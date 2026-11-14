package main

import (
	"context"
	"flag"
	"fmt"
	"math/rand/v2"
	"os"
	"strconv"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/AdityaSinghRajawat/autoreach-backend/config"
	"github.com/AdityaSinghRajawat/autoreach-backend/pkg/db"
	"github.com/AdityaSinghRajawat/autoreach-backend/pkg/utils"
)

const (
	benchTable            = "vector_bench_jobs"
	benchIndex            = "idx_vector_bench_jobs_embedding"
	centresTable          = "vector_bench_centres"
	clusterCount          = 256
	clusterSpread         = 0.35
	querySpread           = 0.10
	distributionUniform   = "uniform"
	distributionClustered = "clustered"
	generationChunk       = 5000
	recallSampleSize      = 20
)

type options struct {
	rows          int
	queries       int
	topK          int
	searchBreadth int
	workMemory    string
	distribution  string
	keep          bool
}

type report struct {
	rows          int
	loadDuration  time.Duration
	indexDuration time.Duration
	indexSize     string
	tableSize     string
	latencies     []float64
	recall        float64
}

func main() {
	config.Init()

	settings := options{}
	flag.IntVar(&settings.rows, "rows", 50000, "number of synthetic job vectors")
	flag.IntVar(&settings.queries, "queries", 200, "number of timed searches")
	flag.IntVar(&settings.topK, "k", config.GetMatchTopK(), "neighbours per search")
	flag.IntVar(&settings.searchBreadth, "ef", config.GetMatchEFSearch(), "hnsw.ef_search")
	flag.StringVar(&settings.workMemory, "work-mem", "1GB", "maintenance_work_mem for the build")
	flag.StringVar(
		&settings.distribution,
		"distribution",
		distributionClustered,
		"clustered (like real embeddings) or uniform (worst case)",
	)
	flag.BoolVar(&settings.keep, "keep", false, "keep the benchmark table afterwards")
	flag.Parse()

	if err := run(context.Background(), settings); err != nil {
		fmt.Fprintln(os.Stderr, "vector benchmark failed:", err)
		os.Exit(1)
	}
}

func run(ctx context.Context, settings options) error {
	pool, err := pgxpool.New(ctx, config.GetDBDSN())
	if err != nil {
		return fmt.Errorf("connect: %w", err)
	}
	defer pool.Close()

	conn, err := pool.Acquire(ctx)
	if err != nil {
		return fmt.Errorf("acquire connection: %w", err)
	}
	defer conn.Release()

	result := report{rows: settings.rows}

	if err := prepareTable(ctx, conn.Conn()); err != nil {
		return err
	}

	if !settings.keep {
		defer func() { _, _ = pool.Exec(ctx, `DROP TABLE IF EXISTS `+benchTable) }()
	}

	if result.loadDuration, err = loadVectors(ctx, conn.Conn(), settings); err != nil {
		return err
	}

	if result.indexDuration, err = buildIndex(ctx, conn.Conn(), settings.workMemory); err != nil {
		return err
	}

	if err := readSizes(ctx, conn.Conn(), &result); err != nil {
		return err
	}

	warmCache(ctx, conn.Conn())

	queries, err := buildQueries(ctx, conn.Conn(), settings)
	if err != nil {
		return err
	}

	if result.latencies, err = timeSearches(ctx, conn.Conn(), settings, queries); err != nil {
		return err
	}

	if result.recall, err = measureRecall(ctx, conn.Conn(), settings, queries); err != nil {
		return err
	}

	printReport(settings, result)

	return nil
}

func prepareTable(ctx context.Context, conn *pgx.Conn) error {
	statements := []string{
		`DROP TABLE IF EXISTS ` + benchTable,
		`CREATE TABLE ` + benchTable + ` (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			status job_status NOT NULL DEFAULT 'ACTIVE',
			embedding vector(` + strconv.Itoa(config.GetOpenAIEmbeddingDimensions()) + `)
		)`,
	}
	for _, statement := range statements {
		if _, err := conn.Exec(ctx, statement); err != nil {
			return fmt.Errorf("prepare table: %w", err)
		}
	}

	return nil
}

func loadVectors(ctx context.Context, conn *pgx.Conn, settings options) (time.Duration, error) {
	startedAt := time.Now()
	dimensions := config.GetOpenAIEmbeddingDimensions()

	if err := prepareCentres(ctx, conn, dimensions); err != nil {
		return 0, err
	}

	spread := 1.0
	if settings.distribution == distributionClustered {
		spread = clusterSpread
	}

	for loaded := 0; loaded < settings.rows; loaded += generationChunk {
		chunk := min(generationChunk, settings.rows-loaded)

		_, err := conn.Exec(ctx, `
			INSERT INTO `+benchTable+` (embedding)
			SELECT centre.embedding + ARRAY(
				SELECT (random() * 2 - 1) * $3::float8
				FROM generate_series(1, $1::int) AS component
				WHERE row_number IS NOT NULL
			)::vector
			FROM generate_series(1, $2::int) AS row_number
			JOIN `+centresTable+` AS centre ON centre.id = 1 + (row_number % $4::int)
		`, dimensions, chunk, spread, clusterCount)
		if err != nil {
			return 0, fmt.Errorf("load vectors: %w", err)
		}

		fmt.Printf("\rloaded %d / %d rows", loaded+chunk, settings.rows)
	}

	fmt.Println()

	return time.Since(startedAt), nil
}

func prepareCentres(ctx context.Context, conn *pgx.Conn, dimensions int) error {
	statements := []string{
		`DROP TABLE IF EXISTS ` + centresTable,
		`CREATE TEMP TABLE ` + centresTable + ` AS
			SELECT centre_number AS id, ARRAY(
				SELECT random() * 2 - 1
				FROM generate_series(1, ` + strconv.Itoa(dimensions) + `) AS component
				WHERE centre_number IS NOT NULL
			)::vector AS embedding
			FROM generate_series(1, ` + strconv.Itoa(clusterCount) + `) AS centre_number`,
	}
	for _, statement := range statements {
		if _, err := conn.Exec(ctx, statement); err != nil {
			return fmt.Errorf("prepare centres: %w", err)
		}
	}

	return nil
}

func buildQueries(ctx context.Context, conn *pgx.Conn, settings options) ([]string, error) {
	total := settings.queries + 2*recallSampleSize

	if settings.distribution == distributionUniform {
		queries := make([]string, 0, total)
		for range total {
			queries = append(queries, randomVector(nil, 1))
		}

		return queries, nil
	}

	rows, err := conn.Query(ctx, `
		SELECT embedding::float4[] FROM `+benchTable+` TABLESAMPLE BERNOULLI (5) LIMIT $1
	`, total)
	if err != nil {
		return nil, fmt.Errorf("sample query vectors: %w", err)
	}

	bases, err := pgx.CollectRows(rows, pgx.RowTo[[]float32])
	if err != nil {
		return nil, fmt.Errorf("sample query vectors: %w", err)
	}

	if len(bases) == 0 {
		return nil, fmt.Errorf("sample query vectors: %w", pgx.ErrNoRows)
	}

	queries := make([]string, 0, total)
	for index := range total {
		queries = append(queries, randomVector(bases[index%len(bases)], querySpread))
	}

	return queries, nil
}

func buildIndex(ctx context.Context, conn *pgx.Conn, workMemory string) (time.Duration, error) {
	startedAt := time.Now()

	statements := []string{
		`SET maintenance_work_mem = '` + workMemory + `'`,
		`SET statement_timeout = 0`,
		`CREATE INDEX ` + benchIndex + ` ON ` + benchTable + `
			USING hnsw (embedding vector_cosine_ops) WITH (m = 16, ef_construction = 64)
			WHERE status = 'ACTIVE'`,
		`ANALYZE ` + benchTable,
	}
	for _, statement := range statements {
		if _, err := conn.Exec(ctx, statement); err != nil {
			return 0, fmt.Errorf("build index: %w", err)
		}
	}

	return time.Since(startedAt), nil
}

func warmCache(ctx context.Context, conn *pgx.Conn) {
	statements := []string{
		`CREATE EXTENSION IF NOT EXISTS pg_prewarm WITH SCHEMA public`,
		`SELECT public.pg_prewarm('` + benchIndex + `'::regclass, 'read')`,
		`SELECT public.pg_prewarm('` + benchTable + `'::regclass, 'read')`,
	}
	for _, statement := range statements {
		if _, err := conn.Exec(ctx, statement); err != nil {
			fmt.Println("cache warm-up skipped:", err)

			return
		}
	}
}

func readSizes(ctx context.Context, conn *pgx.Conn, result *report) error {
	err := conn.QueryRow(ctx, `
		SELECT pg_size_pretty(pg_relation_size($1::regclass)),
			pg_size_pretty(pg_total_relation_size($2::regclass) - pg_indexes_size($2::regclass))
	`, benchIndex, benchTable).Scan(&result.indexSize, &result.tableSize)
	if err != nil {
		return fmt.Errorf("read sizes: %w", err)
	}

	return nil
}

func timeSearches(
	ctx context.Context,
	conn *pgx.Conn,
	settings options,
	queries []string,
) ([]float64, error) {
	latencies := make([]float64, 0, settings.queries+recallSampleSize)

	for index := range settings.queries + recallSampleSize {
		startedAt := time.Now()

		if _, err := nearest(ctx, conn, queries[index], settings, true); err != nil {
			return nil, err
		}

		latencies = append(latencies, float64(time.Since(startedAt).Microseconds())/1000)
	}

	return latencies[recallSampleSize:], nil
}

func measureRecall(
	ctx context.Context,
	conn *pgx.Conn,
	settings options,
	queries []string,
) (float64, error) {
	found, expected := 0, 0

	for index := range recallSampleSize {
		vector := queries[len(queries)-1-index]

		approximate, err := nearest(ctx, conn, vector, settings, true)
		if err != nil {
			return 0, err
		}

		exact, err := nearest(ctx, conn, vector, settings, false)
		if err != nil {
			return 0, err
		}

		exactSet := make(map[string]struct{}, len(exact))
		for _, id := range exact {
			exactSet[id] = struct{}{}
		}

		for _, id := range approximate {
			if _, hit := exactSet[id]; hit {
				found++
			}
		}

		expected += len(exact)
	}

	if expected == 0 {
		return 0, nil
	}

	return float64(found) / float64(expected), nil
}

func nearest(
	ctx context.Context,
	conn *pgx.Conn,
	vector string,
	settings options,
	useIndex bool,
) ([]string, error) {
	transaction, err := conn.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("begin search: %w", err)
	}

	defer func() { _ = transaction.Rollback(ctx) }()

	tuning := `SET LOCAL hnsw.ef_search = ` + strconv.Itoa(settings.searchBreadth)
	if !useIndex {
		tuning = `SET LOCAL enable_indexscan = off`
	}

	if _, err := transaction.Exec(ctx, tuning); err != nil {
		return nil, fmt.Errorf("tune search: %w", err)
	}

	rows, err := transaction.Query(ctx, `
		SELECT id::text
		FROM `+benchTable+`
		WHERE status = 'ACTIVE' AND embedding IS NOT NULL
		ORDER BY embedding <=> $1::vector
		LIMIT $2
	`, vector, settings.topK)
	if err != nil {
		return nil, fmt.Errorf("search: %w", err)
	}

	ids, err := pgx.CollectRows(rows, pgx.RowTo[string])
	if err != nil {
		return nil, fmt.Errorf("search: %w", err)
	}

	return ids, nil
}

func randomVector(base []float32, spread float32) string {
	vector := make([]float32, config.GetOpenAIEmbeddingDimensions())
	for index := range vector {
		vector[index] = (rand.Float32()*2 - 1) * spread
		if base != nil {
			vector[index] += base[index]
		}
	}

	return db.FormatVector(vector)
}

func printReport(settings options, result report) {
	dimensions := config.GetOpenAIEmbeddingDimensions()

	fmt.Println()
	fmt.Println("Vector search benchmark")
	fmt.Println("-----------------------")
	fmt.Printf("rows              %d (%d dimensions)\n", result.rows, dimensions)
	fmt.Printf("distribution      %s\n", settings.distribution)
	fmt.Printf("table size        %s\n", result.tableSize)
	fmt.Printf("index size        %s\n", result.indexSize)
	fmt.Printf("load time         %s\n", result.loadDuration.Round(time.Second))
	fmt.Printf("index build time  %s\n", result.indexDuration.Round(time.Second))
	fmt.Printf(
		"searches          %d, top-%d, ef_search=%d\n",
		settings.queries, settings.topK, settings.searchBreadth,
	)
	fmt.Printf("latency p50       %.2f ms\n", utils.Percentile(result.latencies, 0.50))
	fmt.Printf("latency p95       %.2f ms\n", utils.Percentile(result.latencies, 0.95))
	fmt.Printf("latency p99       %.2f ms\n", utils.Percentile(result.latencies, 0.99))
	fmt.Printf(
		"recall@%d         %.3f (against an exact scan, %d samples)\n",
		settings.topK, result.recall, recallSampleSize,
	)
}
