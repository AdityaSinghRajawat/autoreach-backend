package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"mime/multipart"
	"net"
	"net/http"
	"os"
	"slices"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/google/uuid"

	"github.com/AdityaSinghRajawat/autoreach-backend/config"
	"github.com/AdityaSinghRajawat/autoreach-backend/pkg/db"
	"github.com/AdityaSinghRajawat/autoreach-backend/pkg/utils"
	workerConfigs "github.com/AdityaSinghRajawat/autoreach-backend/worker/configs"
	workerCustomTypes "github.com/AdityaSinghRajawat/autoreach-backend/worker/customtypes"
	workerProviders "github.com/AdityaSinghRajawat/autoreach-backend/worker/providers"
)

const (
	modeAPI          = "api"
	modeApply        = "apply"
	stubBoardToken   = "loadtest"
	stubCompany      = "loadtest.dev"
	loadtestPassword = "loadtest-password-1"
	pollInterval     = time.Second
)

type options struct {
	mode        string
	baseURL     string
	duration    time.Duration
	concurrency int
	jobs        int
	stubAddr    string
	timeout     time.Duration
}

type session struct {
	token       string
	candidateID string
}

func main() {
	config.Init()

	settings := options{}
	flag.StringVar(
		&settings.mode,
		"mode",
		modeAPI,
		"api (request latency) or apply (application throughput)",
	)
	flag.StringVar(&settings.baseURL, "base-url", "http://127.0.0.1:8080", "API server")
	flag.DurationVar(
		&settings.duration,
		"duration",
		time.Minute,
		"api mode: how long to send requests",
	)
	flag.IntVar(&settings.concurrency, "concurrency", 20, "api mode: concurrent clients")
	flag.IntVar(&settings.jobs, "jobs", 500, "apply mode: applications to submit")
	flag.StringVar(
		&settings.stubAddr,
		"stub-addr",
		":18090",
		"apply mode: address of the stub Greenhouse API",
	)
	flag.DurationVar(
		&settings.timeout,
		"timeout",
		30*time.Minute,
		"apply mode: give up after this long",
	)
	flag.Parse()

	ctx := context.Background()

	var err error

	switch settings.mode {
	case modeAPI:
		err = runAPI(ctx, settings)
	case modeApply:
		err = runApply(ctx, settings)
	default:
		err = fmt.Errorf("unknown mode %q", settings.mode)
	}

	if err != nil {
		fmt.Fprintln(os.Stderr, "loadtest:", err)
		os.Exit(1)
	}
}

// runAPI sends a mix of dashboard and profile reads for one candidate and reports latency.
func runAPI(ctx context.Context, settings options) error {
	user, err := register(ctx, settings.baseURL)
	if err != nil {
		return err
	}

	candidate := "/api/v1/candidates/" + user.candidateID
	paths := []string{
		"/health",
		candidate,
		candidate + "/metrics",
		candidate + "/jobs",
		candidate + "/applications",
		"/api/v1/dashboard/funnel",
		"/api/v1/dashboard/activity",
	}

	var (
		mutex     sync.Mutex
		latencies []time.Duration
		statuses  = map[int]int{}
		requests  atomic.Int64
	)

	deadline := time.Now().Add(settings.duration)
	client := &http.Client{
		Timeout: 10 * time.Second,
		Transport: &http.Transport{
			MaxIdleConns:        settings.concurrency,
			MaxIdleConnsPerHost: settings.concurrency,
		},
	}
	failures := map[string]int{}
	clients := make([]utils.ConcurrentTask, 0, settings.concurrency)

	for worker := range settings.concurrency {
		clients = append(clients, utils.ConcurrentTask{
			Name: fmt.Sprintf("client %d", worker),
			Run: func(ctx context.Context) error {
				for index := worker; time.Now().Before(deadline); index++ {
					startedAt := time.Now()
					status, failure := get(
						ctx,
						client,
						settings.baseURL+paths[index%len(paths)],
						user.token,
					)
					elapsed := time.Since(startedAt)

					requests.Add(1)
					mutex.Lock()

					latencies = append(latencies, elapsed)
					statuses[status]++

					if failure != "" {
						failures[failure]++
					}
					mutex.Unlock()
				}

				return nil
			},
		})
	}

	if err := utils.RunConcurrently(ctx, clients...); err != nil {
		return err
	}

	slices.Sort(latencies)

	fmt.Printf(
		"requests:    %d in %s (%.0f/s) with %d clients\n",
		requests.Load(),
		settings.duration,
		float64(requests.Load())/settings.duration.Seconds(),
		settings.concurrency,
	)
	fmt.Printf("latency:     p50 %s  p95 %s  p99 %s  max %s\n", percentile(latencies, 50),
		percentile(latencies, 95), percentile(latencies, 99), latencies[len(latencies)-1])
	fmt.Printf("statuses:    %v\n", statuses)

	for failure, count := range failures {
		fmt.Printf("no response: %d x %s\n", count, failure)
	}

	return nil
}

// runApply seeds matched Greenhouse jobs for one candidate, lets the worker apply to all of them
// through a stub Greenhouse API, and reports applications per hour.
func runApply(ctx context.Context, settings options) error {
	stub, err := startStubGreenhouse(settings.stubAddr)
	if err != nil {
		return err
	}
	defer func() { _ = stub.Close() }()

	user, err := register(ctx, settings.baseURL)
	if err != nil {
		return err
	}

	postgres, err := db.ConnectPostgres(ctx, 4)
	if err != nil {
		return err
	}
	defer postgres.Close()

	if err := prepareCandidate(ctx, settings.baseURL, user, postgres); err != nil {
		return err
	}

	if err := seedMatchedJobs(ctx, postgres, user.candidateID, settings.jobs); err != nil {
		return err
	}

	strategy, err := workerProviders.NewWorkerStrategy()
	if err != nil {
		return err
	}

	client, err := strategy.InitializeClient(ctx)
	if err != nil {
		return err
	}
	defer func() { _ = client.Close() }()

	startedAt := time.Now()

	if _, err := client.Publish(ctx, workerConfigs.PublishParams{
		TaskID:  workerConfigs.TaskQueueApplications,
		Payload: workerCustomTypes.QueueApplicationsPayload{CandidateID: user.candidateID},
	}); err != nil {
		return err
	}

	for time.Since(startedAt) < settings.timeout {
		counts := map[string]int{}

		rows, err := postgres.GetPool().Query(ctx,
			`SELECT status::text, COUNT(*) FROM applications WHERE candidate_id = $1 GROUP BY status`,
			user.candidateID)
		if err != nil {
			return err
		}

		for rows.Next() {
			var (
				status string
				count  int
			)
			if err := rows.Scan(&status, &count); err != nil {
				return err
			}

			counts[status] = count
		}

		rows.Close()

		finished := counts["SUBMITTED"] + counts["ACTION_REQUIRED"] + counts["FAILED"]
		fmt.Printf("\r%s  %v   ", time.Since(startedAt).Round(time.Second), counts)

		if finished >= settings.jobs {
			elapsed := time.Since(startedAt)
			fmt.Printf(
				"\napplications: %d finished in %s, %d submitted\n",
				finished,
				elapsed.Round(time.Millisecond),
				counts["SUBMITTED"],
			)
			fmt.Printf(
				"throughput:   %.0f applications/hour\n",
				float64(counts["SUBMITTED"])/elapsed.Hours(),
			)
			fmt.Printf(
				"stub calls:   %d form reads, %d submissions\n",
				stub.reads.Load(),
				stub.submissions.Load(),
			)

			return nil
		}

		time.Sleep(pollInterval)
	}

	return errors.New("timed out before every application finished")
}

func register(ctx context.Context, baseURL string) (session, error) {
	email := "loadtest-" + uuid.NewString()[:8] + "@example.com"
	body, _ := json.Marshal(
		map[string]string{"email": email, "password": loadtestPassword, "fullName": "Load Test"},
	)

	response := struct {
		Data struct {
			AccessToken string `json:"accessToken"`
			User        struct {
				CandidateID string `json:"candidateId"`
			} `json:"user"`
		} `json:"data"`
	}{}

	if err := send(ctx, http.MethodPost, baseURL+"/api/v1/auth/register", "", "application/json",
		bytes.NewReader(body), &response); err != nil {
		return session{}, fmt.Errorf("register: %w", err)
	}

	return session{
		token:       response.Data.AccessToken,
		candidateID: response.Data.User.CandidateID,
	}, nil
}

// prepareCandidate turns on automatic applying, saves an application profile and stores a resume,
// which is marked ready directly because resume embedding needs an OpenAI key.
func prepareCandidate(
	ctx context.Context,
	baseURL string,
	user session,
	postgres *db.PostgresConn,
) error {
	candidate := baseURL + "/api/v1/candidates/" + user.candidateID

	if err := sendJSON(ctx, http.MethodPatch, candidate, user.token,
		map[string]any{"autoApplyEnabled": true, "minMatchScore": 0.8}); err != nil {
		return fmt.Errorf("enable auto-apply: %w", err)
	}

	if err := sendJSON(
		ctx,
		http.MethodPut,
		candidate+"/application-profile",
		user.token,
		map[string]any{
			"phone":          "+1 555 0100",
			"linkedinUrl":    "https://www.linkedin.com/in/load-test",
			"workAuthorized": true,
		},
	); err != nil {
		return fmt.Errorf("save application profile: %w", err)
	}

	form := bytes.Buffer{}
	writer := multipart.NewWriter(&form)

	part, err := writer.CreateFormFile(config.GetResumeFormField(), "resume.pdf")
	if err != nil {
		return err
	}

	_, _ = part.Write(minimalPDF("Load Test", "Go engineer"))
	_ = writer.Close()

	if err := send(
		ctx,
		http.MethodPost,
		candidate+"/resumes",
		user.token,
		writer.FormDataContentType(),
		&form,
		nil,
	); err != nil {
		return fmt.Errorf("upload resume: %w", err)
	}

	_, err = postgres.GetPool().
		Exec(ctx, `UPDATE candidate_resumes SET is_active = TRUE, parse_status = 'EMBEDDED'
		WHERE candidate_id = $1`, user.candidateID)

	return err
}

func seedMatchedJobs(
	ctx context.Context,
	postgres *db.PostgresConn,
	candidateID string,
	count int,
) error {
	_, err := postgres.GetPool().Exec(ctx, `
		WITH jobs AS (
			INSERT INTO job_postings (job_hash, source, company_name, company_domain, job_title, normalized_title,
				job_description, description_hash, ats_type, external_url, external_job_id)
			SELECT md5($1::text || n::text) || md5(n::text), 'GREENHOUSE', 'Load Test Inc', $2, 'Engineer ' || n,
				'engineer ' || n, 'A load test job.', md5('A load test job.'), 'GREENHOUSE',
				'https://boards.greenhouse.io/' || $3 || '/jobs/' || n, n::text
			FROM generate_series(1, $4) AS n
			RETURNING id
		)
		INSERT INTO job_matches (candidate_id, job_id, resume_id, raw_similarity, match_score)
		SELECT $1::uuid, jobs.id, (SELECT id FROM candidate_resumes WHERE candidate_id = $1::uuid AND is_active), 0.95, 0.95
		FROM jobs
	`, candidateID, stubCompany, stubBoardToken, count)
	if err != nil {
		return fmt.Errorf("seed matched jobs: %w", err)
	}

	return nil
}

type stubGreenhouse struct {
	server      *http.Server
	reads       atomic.Int64
	submissions atomic.Int64
}

// startStubGreenhouse answers the Job Board API: each job's form asks only for standard fields, and
// every submission is accepted.
func startStubGreenhouse(address string) (*stubGreenhouse, error) {
	stub := &stubGreenhouse{}
	form := `{"questions":[
		{"required":true,"label":"First Name","fields":[{"name":"first_name","type":"input_text"}]},
		{"required":true,"label":"Last Name","fields":[{"name":"last_name","type":"input_text"}]},
		{"required":true,"label":"Email","fields":[{"name":"email","type":"input_text"}]},
		{"required":true,"label":"Resume/CV","fields":[{"name":"resume","type":"input_file"}]}]}`

	stub.server = &http.Server{
		ReadHeaderTimeout: 5 * time.Second,
		Handler: http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
			writer.Header().Set("Content-Type", "application/json")

			if !strings.Contains(request.URL.Path, "/boards/"+stubBoardToken+"/jobs/") {
				writer.WriteHeader(http.StatusNotFound)

				return
			}

			if request.Method == http.MethodPost {
				stub.submissions.Add(1)

				_, _ = io.Copy(io.Discard, request.Body)
				_, _ = io.WriteString(writer, `{}`)

				return
			}

			stub.reads.Add(1)

			_, _ = io.WriteString(writer, form)
		}),
	}

	listener, err := (&net.ListenConfig{}).Listen(context.Background(), "tcp", address)
	if err != nil {
		return nil, fmt.Errorf("start stub greenhouse: %w", err)
	}

	utils.RunInBackground(context.Background(), "stub greenhouse", func(context.Context) {
		_ = stub.server.Serve(listener)
	})

	return stub, nil
}

func (s *stubGreenhouse) Close() error {
	return s.server.Close()
}

// get returns the status, or 0 and the reason when no response arrived.
func get(ctx context.Context, client *http.Client, url, token string) (int, string) {
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return 0, err.Error()
	}

	request.Header.Set("Authorization", "Bearer "+token)

	response, err := client.Do(request)
	if err != nil {
		return 0, err.Error()
	}
	defer func() { _ = response.Body.Close() }()

	_, _ = io.Copy(io.Discard, response.Body)

	return response.StatusCode, ""
}

func sendJSON(ctx context.Context, method, url, token string, payload any) error {
	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	return send(ctx, method, url, token, "application/json", bytes.NewReader(body), nil)
}

func send(
	ctx context.Context,
	method, url, token, contentType string,
	body io.Reader,
	result any,
) error {
	request, err := http.NewRequestWithContext(ctx, method, url, body)
	if err != nil {
		return err
	}

	request.Header.Set("Content-Type", contentType)

	if token != "" {
		request.Header.Set("Authorization", "Bearer "+token)
	}

	response, err := http.DefaultClient.Do(request)
	if err != nil {
		return err
	}
	defer func() { _ = response.Body.Close() }()

	raw, _ := io.ReadAll(response.Body)
	if response.StatusCode >= http.StatusMultipleChoices {
		return fmt.Errorf("%s %s: %d %s", method, url, response.StatusCode, raw)
	}

	if result == nil {
		return nil
	}

	return json.Unmarshal(raw, result)
}

func percentile(sorted []time.Duration, percent int) time.Duration {
	if len(sorted) == 0 {
		return 0
	}

	return sorted[min(len(sorted)-1, len(sorted)*percent/100)].Round(100 * time.Microsecond)
}

func minimalPDF(lines ...string) []byte {
	content := "BT /F1 12 Tf 72 720 Td 14 TL"
	for _, line := range lines {
		content += " (" + line + ") Tj T*"
	}

	content += " ET"

	objects := []string{
		"<< /Type /Catalog /Pages 2 0 R >>",
		"<< /Type /Pages /Kids [3 0 R] /Count 1 >>",
		"<< /Type /Page /Parent 2 0 R /MediaBox [0 0 612 792] /Contents 4 0 R " +
			"/Resources << /Font << /F1 5 0 R >> >> >>",
		fmt.Sprintf("<< /Length %d >>\nstream\n%s\nendstream", len(content)+1, content),
		"<< /Type /Font /Subtype /Type1 /BaseFont /Helvetica >>",
	}

	document := bytes.Buffer{}
	document.WriteString("%PDF-1.4\n")

	offsets := make([]int, 0, len(objects))
	for index, object := range objects {
		offsets = append(offsets, document.Len())
		fmt.Fprintf(&document, "%d 0 obj\n%s\nendobj\n", index+1, object)
	}

	startXref := document.Len()
	fmt.Fprintf(&document, "xref\n0 %d\n0000000000 65535 f \n", len(objects)+1)

	for _, offset := range offsets {
		fmt.Fprintf(&document, "%010d 00000 n \n", offset)
	}

	fmt.Fprintf(
		&document,
		"trailer\n<< /Size %d /Root 1 0 R >>\nstartxref\n%d\n%%%%EOF\n",
		len(objects)+1,
		startXref,
	)

	return document.Bytes()
}
