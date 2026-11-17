package main

import (
	"bufio"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/AdityaSinghRajawat/autoreach-backend/config"
	aiIntegration "github.com/AdityaSinghRajawat/autoreach-backend/pkg/integrations/ai"
	"github.com/AdityaSinghRajawat/autoreach-backend/pkg/utils"
)

const (
	labelGood     = "good"
	labelBad      = "bad"
	maxLineBytes  = 4 << 20
	embeddingPair = 2
)

type labelledPair struct {
	Resume string `json:"resume"`
	Job    string `json:"job"`
	Label  string `json:"label"`
}

func main() {
	config.Init()

	path := flag.String("pairs", "", "JSONL file of labelled resume and job pairs")

	flag.Parse()

	if err := run(context.Background(), *path); err != nil {
		fmt.Fprintln(os.Stderr, "calibration failed:", err)
		os.Exit(1)
	}
}

func run(ctx context.Context, path string) error {
	pairs, err := readPairs(path)
	if err != nil {
		return err
	}

	embedder := aiIntegration.NewAIIntegration(config.GetAIBaseURL())
	maxChars := config.GetAIEmbeddingMaxInputChars()
	similarities := map[string][]float64{}

	for index, pair := range pairs {
		vectors, err := embedder.CreateEmbeddings(ctx, []string{
			truncate(pair.Resume, maxChars),
			truncate(pair.Job, maxChars),
		})
		if err != nil {
			return fmt.Errorf("embed pair %d: %w", index+1, err)
		}

		if len(vectors) != embeddingPair {
			return fmt.Errorf("embed pair %d: %w", index+1, config.ErrEmbeddingCountMismatch)
		}

		similarities[pair.Label] = append(
			similarities[pair.Label],
			utils.CosineSimilarity(vectors[0], vectors[1]),
		)

		fmt.Printf("\rembedded %d / %d pairs", index+1, len(pairs))
	}

	printReport(similarities[labelGood], similarities[labelBad])

	return nil
}

func readPairs(path string) ([]labelledPair, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("open pairs file: %w", err)
	}

	defer func() { _ = file.Close() }()

	scanner := bufio.NewScanner(file)
	scanner.Buffer(make([]byte, 0, maxLineBytes), maxLineBytes)

	pairs := make([]labelledPair, 0, 64)

	for line := 1; scanner.Scan(); line++ {
		if strings.TrimSpace(scanner.Text()) == "" {
			continue
		}

		pair := labelledPair{}
		if err := json.Unmarshal(scanner.Bytes(), &pair); err != nil {
			return nil, fmt.Errorf("line %d: %w", line, err)
		}

		pair.Label = strings.ToLower(strings.TrimSpace(pair.Label))
		if !isUsable(pair) {
			return nil, fmt.Errorf("line %d: %w", line, config.ErrInvalidCalibrationPair)
		}

		pairs = append(pairs, pair)
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("read pairs file: %w", err)
	}

	if len(pairs) == 0 {
		return nil, fmt.Errorf("read pairs file: %w", config.ErrInvalidCalibrationPair)
	}

	return pairs, nil
}

func isUsable(pair labelledPair) bool {
	return pair.Resume != "" && pair.Job != "" &&
		(pair.Label == labelGood || pair.Label == labelBad)
}

func truncate(text string, maxChars int) string {
	characters := []rune(text)
	if len(characters) > maxChars {
		characters = characters[:maxChars]
	}

	return string(characters)
}

func printReport(good, bad []float64) {
	low, high := utils.SuggestCalibration(good, bad)

	fmt.Println()
	fmt.Println()
	fmt.Println("Raw cosine similarity by label")
	fmt.Println("------------------------------")
	printDistribution("good", good)
	printDistribution("bad", bad)
	fmt.Println()
	fmt.Println("Suggested settings")
	fmt.Println("------------------")
	fmt.Printf("MATCH_CALIBRATION_LOW=%.4f\n", low)
	fmt.Printf("MATCH_CALIBRATION_HIGH=%.4f\n", high)
	fmt.Println()
	fmt.Println("With these, three quarters of the pairs labelled good score 0.80 or more,")
	fmt.Println("and the median pair labelled bad scores 0.")
}

func printDistribution(label string, values []float64) {
	fmt.Printf(
		"%-5s n=%-4d p10=%.3f p25=%.3f p50=%.3f p75=%.3f p90=%.3f\n",
		label,
		len(values),
		utils.Percentile(values, 0.10),
		utils.Percentile(values, 0.25),
		utils.Percentile(values, 0.50),
		utils.Percentile(values, 0.75),
		utils.Percentile(values, 0.90),
	)
}
