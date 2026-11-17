//go:build integration

package integration

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"path"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/AdityaSinghRajawat/autoreach-backend/config"
	jobTypes "github.com/AdityaSinghRajawat/autoreach-backend/pkg/customtypes/jobManagement"
	jobModel "github.com/AdityaSinghRajawat/autoreach-backend/pkg/models/jobManagement"
	"github.com/AdityaSinghRajawat/autoreach-backend/pkg/services/enrichmentManagement/providers"
	"github.com/AdityaSinghRajawat/autoreach-backend/pkg/services/jobManagement/jobBoards"
	"github.com/AdityaSinghRajawat/autoreach-backend/routes"
	"github.com/AdityaSinghRajawat/autoreach-backend/tests/helpers"
)

const jobSourcesPath = "/api/v1/admin/job-sources"

const greenhouseBoardBothJobs = `{"jobs":[
	{"id":1,"title":"Senior Go Engineer","absolute_url":"https://boards.greenhouse.io/acme/jobs/1",
	 "content":"&lt;p&gt;Build &lt;strong&gt;APIs&lt;/strong&gt; in Go.&lt;/p&gt;&lt;ul&gt;&lt;li&gt;Postgres&lt;/li&gt;&lt;/ul&gt;",
	 "updated_at":"2026-11-01T10:00:00-05:00","location":{"name":"Remote"},"departments":[{"name":"Engineering"}]},
	{"id":2,"title":"Staff Go Engineer","absolute_url":"https://boards.greenhouse.io/acme/jobs/2",
	 "content":"&lt;p&gt;Lead the platform team.&lt;/p&gt;","location":{"name":"Berlin"},"departments":[]},
	{"id":3,"title":"","absolute_url":"https://boards.greenhouse.io/acme/jobs/3","content":"untitled",
	 "location":{"name":""},"departments":[]}
]}`

const greenhouseBoardOneJob = `{"jobs":[
	{"id":1,"title":"Senior Go Engineer","absolute_url":"https://boards.greenhouse.io/acme/jobs/1",
	 "content":"&lt;p&gt;Build &lt;strong&gt;APIs&lt;/strong&gt; in Go.&lt;/p&gt;","location":{"name":"Remote"},"departments":[]}
]}`

const leverPostings = `[{"id":"abc-123","text":"Backend Engineer","hostedUrl":"https://jobs.lever.co/globex/abc-123",
	"descriptionPlain":"Join Globex.","additionalPlain":"Equal opportunity employer.","createdAt":1762000000000,
	"lists":[{"text":"Requirements","content":"<li>Go</li><li>Kubernetes</li>"}],
	"categories":{"location":"London","team":"Platform"}}]`

const ashbyBoard = `{"apiVersion":"1","jobs":[
	{"id":"ash-1","title":"Platform Engineer","department":"Engineering","team":"Infra","location":"Austin",
	 "isListed":true,"publishedAt":"2026-11-02T09:00:00.000+00:00","jobUrl":"https://jobs.ashbyhq.com/initech/ash-1",
	 "descriptionPlain":"","descriptionHtml":"<p>Run&nbsp;the platform.</p>"},
	{"id":"ash-2","title":"Hidden Role","location":"Austin","isListed":false,
	 "jobUrl":"https://jobs.ashbyhq.com/initech/ash-2","descriptionPlain":"Not public."}
]}`

const smartRecruitersPostingsFirstPage = `{"offset":0,"limit":1,"totalFound":2,"content":[
	{"id":"sr-1","name":"Data Engineer","releasedDate":"2026-11-03T08:00:00.000Z",
	 "location":{"city":"Pune","fullLocation":"Pune, Maharashtra, India"},"department":{"label":"Data"}}]}`

const smartRecruitersPostingsSecondPage = `{"offset":1,"limit":1,"totalFound":2,"content":[
	{"id":"sr-2","name":"ML Engineer","location":{"city":"Bengaluru"},"department":{"label":"AI"}}]}`

const smartRecruitersPosting = `{"id":"%s","name":"%s","postingUrl":"https://jobs.smartrecruiters.com/Hooli/%s",
	"location":{"city":"Pune","fullLocation":"Pune, Maharashtra, India"},"department":{"label":"Data"},
	"jobAd":{"sections":{"companyDescription":{"title":"Company Description","text":"<p>Hooli builds things.</p>"},
	"jobDescription":{"title":"Job Description","text":"<p>Own the pipelines.</p>"},
	"qualifications":{"title":"Qualifications","text":"<ul><li>SQL</li></ul>"},
	"additionalInformation":{"title":"Additional Information","text":""}}}}`

const workdayJobs = `{"total":2,"jobPostings":[
	{"title":"Platform Engineer","externalPath":"/job/Austin/Platform-Engineer_R1","locationsText":"Austin, TX"},
	{"title":"Data Engineer","externalPath":"/job/Remote/Data-Engineer_R2","locationsText":"Remote"}]}`

const workdayJob = `{"jobPostingInfo":{"title":"%s","jobDescription":"<p>Work on %s.</p>","location":"%s",
	"startDate":"2026-11-10","externalUrl":"https://initech.wd1.myworkdayjobs.com/careers%s"}}`

const adzunaResults = `{"count":2,"results":[
	{"id":"5922025200","title":"<strong>Go</strong> Developer","description":"Build services in Go for payments.",
	 "redirect_url":"https://www.adzuna.in/land/ad/1","created":"2026-11-12T10:00:00Z",
	 "salary_min":1500000,"salary_max":2500000,"salary_is_predicted":"0",
	 "company":{"display_name":"Initrode Ltd"},"location":{"display_name":"Pune, Maharashtra"},
	 "category":{"label":"IT Jobs"}},
	{"id":5923277084,"title":"Go Engineer","description":"Mystery company role.",
	 "redirect_url":"https://www.adzuna.in/land/ad/2","company":{"display_name":"Mystery Co"},
	 "location":{"display_name":"Remote"}}]}`

const careerRobots = "User-agent: *\nDisallow: /private/\n"

const careerSitemapIndex = `<?xml version="1.0" encoding="UTF-8"?>
<sitemapindex xmlns="http://www.sitemaps.org/schemas/sitemap/0.9">
	<sitemap><loc>https://careers.initech.com/sitemap-jobs.xml</loc></sitemap>
</sitemapindex>`

const careerSitemap = `<?xml version="1.0" encoding="UTF-8"?>
<urlset xmlns="http://www.sitemaps.org/schemas/sitemap/0.9">
	<url><loc>https://careers.initech.com/jobs/1</loc></url>
	<url><loc>https://careers.initech.com/jobs/2</loc></url>
	<url><loc>https://careers.initech.com/private/jobs/3</loc></url>
	<url><loc>https://careers.initech.com/about</loc></url>
</urlset>`

const careerJobOne = `<html><head><script type="application/ld+json">
{"@context":"https://schema.org","@type":"JobPosting","title":"Backend Engineer",
 "description":"&lt;p&gt;Design &lt;b&gt;APIs&lt;/b&gt;.&lt;/p&gt;","datePosted":"2026-11-01",
 "url":"https://careers.initech.com/jobs/1","identifier":{"@type":"PropertyValue","value":"BE-1"},
 "jobLocation":{"@type":"Place","address":{"addressLocality":"Austin","addressRegion":"TX","addressCountry":"US"}},
 "baseSalary":{"@type":"MonetaryAmount","currency":"USD","value":{"minValue":150000,"maxValue":190000}}}
</script></head><body>Backend Engineer</body></html>`

const careerJobTwo = `<html><head><script type='application/ld+json'>
{"@context":"https://schema.org","@graph":[{"@type":"Organization","name":"Initech"},
 {"@type":["JobPosting"],"title":"Support Engineer","description":"Help customers.",
  "url":"/jobs/2","jobLocationType":"TELECOMMUTE"}]}
</script></head></html>`

type boardsStub struct {
	server                 *httptest.Server
	jobRemoved             atomic.Bool
	boardFailing           atomic.Bool
	smartRecruitersDetails atomic.Int32
	careerServer           *httptest.Server
	careerPageReads        atomic.Int32
}

func newBoardsStub(t *testing.T) *boardsStub {
	t.Helper()

	stub := &boardsStub{}
	stub.server = httptest.NewServer(
		http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
			writer.Header().Set("Content-Type", "application/json")

			switch {
			case stub.boardFailing.Load():
				writer.WriteHeader(http.StatusNotFound)
			case request.URL.Path == "/greenhouse/boards/acme/jobs" && request.URL.Query().Get("content") == "true":
				if stub.jobRemoved.Load() {
					_, _ = io.WriteString(writer, greenhouseBoardOneJob)

					return
				}

				_, _ = io.WriteString(writer, greenhouseBoardBothJobs)
			case request.URL.Path == "/lever-eu/postings/globex" && request.URL.Query().Get("mode") == "json":
				_, _ = io.WriteString(writer, leverPostings)
			case request.URL.Path == "/ashby/job-board/initech":
				_, _ = io.WriteString(writer, ashbyBoard)
			case request.URL.Path == "/smartrecruiters/companies/hooli/postings":
				if request.URL.Query().Get("offset") == "1" {
					_, _ = io.WriteString(writer, smartRecruitersPostingsSecondPage)

					return
				}

				_, _ = io.WriteString(writer, smartRecruitersPostingsFirstPage)
			case strings.HasPrefix(request.URL.Path, "/smartrecruiters/companies/hooli/postings/"):
				stub.smartRecruitersDetails.Add(1)

				postingID := path.Base(request.URL.Path)
				name := map[string]string{"sr-1": "Data Engineer", "sr-2": "ML Engineer"}[postingID]
				_, _ = fmt.Fprintf(writer, smartRecruitersPosting, postingID, name, postingID)
			case request.Method == http.MethodPost &&
				request.URL.Path == "/workday/initech/wd1/wday/cxs/initech/careers/jobs":
				_, _ = io.WriteString(writer, workdayJobs)
			case strings.HasPrefix(request.URL.Path, "/workday/initech/wd1/wday/cxs/initech/careers/job/"):
				externalPath := strings.TrimPrefix(
					request.URL.Path,
					"/workday/initech/wd1/wday/cxs/initech/careers",
				)
				title, location := "Platform Engineer", "Austin, TX"

				if strings.Contains(externalPath, "Data-Engineer") {
					title, location = "Data Engineer", "Remote"
				}

				_, _ = fmt.Fprintf(writer, workdayJob, title, title, location, externalPath)
			case request.URL.Path == "/adzuna/jobs/in/search/1" &&
				request.URL.Query().Get("app_key") == "adzuna-key" &&
				request.URL.Query().Get("what") == "golang":
				_, _ = io.WriteString(writer, adzunaResults)
			default:
				writer.WriteHeader(http.StatusNotFound)
			}
		}),
	)
	t.Cleanup(stub.server.Close)

	stub.careerServer = httptest.NewTLSServer(
		http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
			pages := map[string]string{
				"/robots.txt":       careerRobots,
				"/sitemap.xml":      careerSitemapIndex,
				"/sitemap-jobs.xml": careerSitemap,
				"/jobs/1":           careerJobOne,
				"/jobs/2":           careerJobTwo,
				"/about":            "<html>About us</html>",
				"/private/jobs/3":   careerJobOne,
			}

			page, found := pages[request.URL.Path]
			if !found {
				writer.WriteHeader(http.StatusNotFound)

				return
			}

			if strings.HasPrefix(request.URL.Path, "/jobs/") ||
				strings.HasPrefix(request.URL.Path, "/private/") {
				stub.careerPageReads.Add(1)
			}

			_, _ = io.WriteString(writer, page)
		}),
	)
	t.Cleanup(stub.careerServer.Close)

	return stub
}

type jobSourceFixture struct {
	jobFixture
	stub       *boardsStub
	adminToken string
}

func newJobSourceFixture(t *testing.T) jobSourceFixture {
	t.Helper()

	fixture := newJobFixture(t)
	stub := newBoardsStub(t)

	fixture.deps.JobBoards = jobBoards.NewJobBoardRegistry(jobBoards.RegistryOptions{
		GreenhouseAPIURL:      stub.server.URL + "/greenhouse",
		LeverAPIURL:           stub.server.URL + "/lever",
		LeverEUAPIURL:         stub.server.URL + "/lever-eu",
		AshbyAPIURL:           stub.server.URL + "/ashby",
		SmartRecruitersAPIURL: stub.server.URL + "/smartrecruiters",
		WorkdayHostURLFormat:  stub.server.URL + "/workday/%s/%s",
		AdzunaAPIURL:          stub.server.URL + "/adzuna",
		AdzunaAppID:           "adzuna-id",
		AdzunaAppKey:          "adzuna-key",
		CareerPageTransport:   stub.careerTransport(),
	})
	require.NoError(t, fixture.deps.Wire())
	fixture.router = routes.NewRoutes(fixture.deps)

	t.Cleanup(func() {
		fixture.deps.JobBoards = nil
		require.NoError(t, fixture.deps.Wire())
	})

	return jobSourceFixture{
		jobFixture: fixture,
		stub:       stub,
		adminToken: loginAdmin(t, fixture.router),
	}
}

func (f jobSourceFixture) createSource(t *testing.T, body map[string]any) jobModel.JobSource {
	t.Helper()

	created := helpers.Do(t, f.router, http.MethodPost, jobSourcesPath, body, bearer(f.adminToken))
	require.Equal(t, http.StatusCreated, created.Recorder.Code, string(created.Body))
	helpers.AssertMatchesSpec(t, created)

	return helpers.DecodeData[jobModel.JobSource](t, created)
}

func (f jobSourceFixture) poll(t *testing.T, sourceID string) helpers.Exchange {
	t.Helper()

	polled := helpers.Do(
		t,
		f.router,
		http.MethodPost,
		jobSourcesPath+"/"+sourceID+"/poll",
		nil,
		bearer(f.adminToken),
	)
	helpers.AssertMatchesSpec(t, polled)

	return polled
}

func (f jobSourceFixture) sourceJobs(t *testing.T, sourceID string) map[string]jobModel.JobPosting {
	t.Helper()

	rows, err := f.deps.Postgres.GetPool().Query(context.Background(), `
		SELECT external_job_id, job_title, job_description, COALESCE(location, ''), COALESCE(department, ''),
			source::text, status::text, external_url
		FROM job_postings WHERE job_source_id = $1
	`, sourceID)
	require.NoError(t, err)

	defer rows.Close()

	jobs := map[string]jobModel.JobPosting{}

	for rows.Next() {
		var (
			job                              jobModel.JobPosting
			externalID, location, department string
		)
		require.NoError(t, rows.Scan(
			&externalID,
			&job.JobTitle,
			&job.JobDescription,
			&location,
			&department,
			&job.Source,
			&job.Status,
			&job.ExternalURL,
		))
		job.Location, job.Department = &location, &department
		jobs[externalID] = job
	}

	require.NoError(t, rows.Err())

	return jobs
}

func TestGreenhouseBoardPollIngestsJobsAndClosesRemovedOnes(t *testing.T) {
	fixture := newJobSourceFixture(t)
	source := fixture.createSource(t, map[string]any{
		"boardType":     "GREENHOUSE",
		"boardToken":    "Acme",
		"companyName":   "Acme Corp",
		"companyDomain": "acme.com",
	})
	assert.Equal(t, "acme", source.BoardToken)

	duplicate := helpers.Do(t, fixture.router, http.MethodPost, jobSourcesPath, map[string]any{
		"boardType":     "GREENHOUSE",
		"boardToken":    "acme",
		"companyName":   "Acme",
		"companyDomain": "acme.com",
	}, bearer(fixture.adminToken))
	assert.Equal(t, http.StatusConflict, duplicate.Recorder.Code)

	first := fixture.poll(t, source.ID)
	require.Equal(t, http.StatusOK, first.Recorder.Code, string(first.Body))
	firstResult := helpers.DecodeData[jobTypes.PollJobSourceResult](t, first)
	assert.Equal(
		t,
		jobTypes.PollJobSourceResult{SourceID: source.ID, Fetched: 3, Ingested: 2, Skipped: 1},
		firstResult,
	)

	jobs := fixture.sourceJobs(t, source.ID)
	require.Len(t, jobs, 2)
	assert.Equal(t, "Build APIs in Go.\nPostgres", jobs["1"].JobDescription)
	assert.Equal(t, "Engineering", *jobs["1"].Department)
	assert.Equal(t, "Berlin", *jobs["2"].Location)
	assert.Equal(t, "GREENHOUSE", jobs["1"].Source)
	assert.Equal(t, "https://job-boards.greenhouse.io/acme/jobs/1", jobs["1"].ExternalURL,
		"board jobs link to the canonical Greenhouse page, whatever absolute_url says")

	fixture.stub.jobRemoved.Store(true)
	second := fixture.poll(t, source.ID)
	require.Equal(t, http.StatusOK, second.Recorder.Code, string(second.Body))
	assert.Equal(t, 1, helpers.DecodeData[jobTypes.PollJobSourceResult](t, second).Closed)

	jobs = fixture.sourceJobs(t, source.ID)
	assert.Equal(t, "ACTIVE", string(jobs["1"].Status))
	assert.Equal(t, "CLOSED", string(jobs["2"].Status))

	fixture.stub.boardFailing.Store(true)
	failed := fixture.poll(t, source.ID)
	assert.NotEqual(t, http.StatusOK, failed.Recorder.Code)
	assert.Equal(t, "ACTIVE", string(fixture.sourceJobs(t, source.ID)["1"].Status),
		"a failed fetch must not close any job")

	listed := helpers.Do(
		t,
		fixture.router,
		http.MethodGet,
		jobSourcesPath,
		nil,
		bearer(fixture.adminToken),
	)
	require.Equal(t, http.StatusOK, listed.Recorder.Code)
	helpers.AssertMatchesSpec(t, listed)
	sources := helpers.DecodeData[[]jobModel.JobSource](t, listed)
	require.Len(t, sources, 1)
	require.NotNil(t, sources[0].LastJobCount)
	assert.Equal(t, 1, *sources[0].LastJobCount)
	require.NotNil(t, sources[0].LastError)
}

func TestLeverEUBoardPollMapsPostings(t *testing.T) {
	fixture := newJobSourceFixture(t)
	source := fixture.createSource(t, map[string]any{
		"boardType":     "LEVER",
		"boardToken":    "globex",
		"region":        "EU",
		"companyName":   "Globex",
		"companyDomain": "globex.com",
	})

	polled := fixture.poll(t, source.ID)
	require.Equal(t, http.StatusOK, polled.Recorder.Code, string(polled.Body))

	job := fixture.sourceJobs(t, source.ID)["abc-123"]
	assert.Equal(t, "Backend Engineer", job.JobTitle)
	assert.Equal(
		t,
		"Join Globex.\n\nRequirements\nGo\nKubernetes\n\nEqual opportunity employer.",
		job.JobDescription,
	)
	assert.Equal(t, "London", *job.Location)
	assert.Equal(t, "Platform", *job.Department)
}

func TestJobSourcesCanBePausedAndDeleted(t *testing.T) {
	fixture := newJobSourceFixture(t)
	source := fixture.createSource(t, map[string]any{
		"boardType":     "GREENHOUSE",
		"boardToken":    "acme",
		"companyName":   "Acme",
		"companyDomain": "acme.com",
	})
	sourcePath := jobSourcesPath + "/" + source.ID

	paused := helpers.Do(
		t,
		fixture.router,
		http.MethodPatch,
		sourcePath,
		map[string]any{"isActive": false},
		bearer(fixture.adminToken),
	)
	require.Equal(t, http.StatusOK, paused.Recorder.Code, string(paused.Body))
	helpers.AssertMatchesSpec(t, paused)
	assert.False(t, helpers.DecodeData[jobModel.JobSource](t, paused).IsActive)

	scheduled, customErr := fixture.deps.Services.Job.ScheduleJobSourcePolls(context.Background())
	require.Nil(t, customErr)
	assert.Zero(t, scheduled, "paused sources are not polled")

	empty := helpers.Do(
		t,
		fixture.router,
		http.MethodPatch,
		sourcePath,
		map[string]any{},
		bearer(fixture.adminToken),
	)
	assert.Equal(t, http.StatusBadRequest, empty.Recorder.Code)

	deleted := helpers.Do(
		t,
		fixture.router,
		http.MethodDelete,
		sourcePath,
		nil,
		bearer(fixture.adminToken),
	)
	assert.Equal(t, http.StatusNoContent, deleted.Recorder.Code)

	missing := helpers.Do(
		t,
		fixture.router,
		http.MethodDelete,
		sourcePath,
		nil,
		bearer(fixture.adminToken),
	)
	assert.Equal(t, http.StatusNotFound, missing.Recorder.Code)
	helpers.AssertMatchesSpec(t, missing)
}

func TestJobSourceRoutesRejectNonAdmins(t *testing.T) {
	fixture := newJobFixture(t)

	forbidden := helpers.Do(
		t,
		fixture.router,
		http.MethodGet,
		jobSourcesPath,
		nil,
		bearer(fixture.userToken),
	)
	anonymous := helpers.Do(t, fixture.router, http.MethodGet, jobSourcesPath, nil, nil)

	assert.Equal(t, http.StatusForbidden, forbidden.Recorder.Code)
	assert.Equal(t, http.StatusUnauthorized, anonymous.Recorder.Code)
}

func TestNaukriAndWellfoundJobsCanBeIngested(t *testing.T) {
	fixture := newJobFixture(t)

	ingested := fixture.ingest(t,
		withFields(validJob(), map[string]any{
			"source":      "NAUKRI",
			"atsType":     "OTHER",
			"externalUrl": "https://www.naukri.com/job-listings-go-1",
		}),
		withFields(validJob(), map[string]any{
			"source":      "WELLFOUND",
			"atsType":     "OTHER",
			"externalUrl": "https://wellfound.com/jobs/2-go",
		}),
	)

	assert.Equal(t, 2, ingested.IngestedCount)
}

func TestAshbyBoardPollSkipsUnlistedJobs(t *testing.T) {
	fixture := newJobSourceFixture(t)
	source := fixture.createSource(t, map[string]any{
		"boardType":     "ASHBY",
		"boardToken":    "initech",
		"companyName":   "Initech",
		"companyDomain": "initech.com",
	})

	polled := fixture.poll(t, source.ID)
	require.Equal(t, http.StatusOK, polled.Recorder.Code, string(polled.Body))

	jobs := fixture.sourceJobs(t, source.ID)
	require.Len(t, jobs, 1, "unlisted Ashby jobs are not public")
	assert.Equal(t, "Run the platform.", jobs["ash-1"].JobDescription)
	assert.Equal(t, "Engineering", *jobs["ash-1"].Department)
	assert.Equal(t, "ASHBY", jobs["ash-1"].Source)
}

func TestSmartRecruitersBoardPollReadsOnlyNewPostings(t *testing.T) {
	fixture := newJobSourceFixture(t)
	source := fixture.createSource(t, map[string]any{
		"boardType":     "SMARTRECRUITERS",
		"boardToken":    "Hooli",
		"companyName":   "Hooli",
		"companyDomain": "hooli.com",
	})

	first := fixture.poll(t, source.ID)
	require.Equal(t, http.StatusOK, first.Recorder.Code, string(first.Body))
	assert.Equal(t, 2, helpers.DecodeData[jobTypes.PollJobSourceResult](t, first).Ingested)

	jobs := fixture.sourceJobs(t, source.ID)
	require.Len(t, jobs, 2)
	assert.Equal(
		t,
		"Company Description\nHooli builds things.\n\nJob Description\nOwn the pipelines.\n\nQualifications\nSQL",
		jobs["sr-1"].JobDescription,
	)
	assert.Equal(t, "https://jobs.smartrecruiters.com/Hooli/sr-1", jobs["sr-1"].ExternalURL)
	assert.Equal(t, "Pune, Maharashtra, India", *jobs["sr-1"].Location)
	assert.Equal(t, int32(2), fixture.stub.smartRecruitersDetails.Load())

	second := fixture.poll(t, source.ID)
	require.Equal(t, http.StatusOK, second.Recorder.Code, string(second.Body))
	secondResult := helpers.DecodeData[jobTypes.PollJobSourceResult](t, second)

	assert.Equal(t, 2, secondResult.Unchanged)
	assert.Zero(t, secondResult.Closed, "postings kept but not re-read must stay active")
	assert.Equal(t, int32(2), fixture.stub.smartRecruitersDetails.Load(),
		"known postings are not fetched again")
}

// careerTransport sends every career-site request to the TLS stub, whatever host it names.
func (s *boardsStub) careerTransport() http.RoundTripper {
	stubURL, _ := url.Parse(s.careerServer.URL)
	inner := s.careerServer.Client().Transport

	return roundTripFunc(func(request *http.Request) (*http.Response, error) {
		redirected := request.Clone(request.Context())
		redirected.URL.Host = stubURL.Host

		return inner.RoundTrip(redirected)
	})
}

type roundTripFunc func(request *http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(
	request *http.Request,
) (*http.Response, error) {
	return f(request)
}

func (f jobSourceFixture) sourceJobColumn(
	t *testing.T,
	sourceID, externalID, column string,
) string {
	t.Helper()

	var value string
	require.NoError(t, f.deps.Postgres.GetPool().QueryRow(context.Background(),
		`SELECT COALESCE(`+column+`::text, '') FROM job_postings WHERE job_source_id = $1 AND external_job_id = $2`,
		sourceID, externalID,
	).
		Scan(&value))

	return value
}

func TestJobSourceRequestsAreValidatedPerBoardType(t *testing.T) {
	fixture := newJobSourceFixture(t)

	for name, body := range map[string]map[string]any{
		"company board without a company": {"boardType": "GREENHOUSE", "boardToken": "acme"},
		"search without a query":          {"boardType": "ADZUNA", "boardToken": "in"},
		"search with a company": {
			"boardType": "ADZUNA", "boardToken": "in", "searchQuery": "golang",
			"companyName": "Acme", "companyDomain": "acme.com",
		},
		"unsupported search country": {"boardType": "ADZUNA", "boardToken": "xx", "searchQuery": "golang"},
		"query on a company board": {
			"boardType": "LEVER", "boardToken": "acme", "searchQuery": "go",
			"companyName": "Acme", "companyDomain": "acme.com",
		},
		"malformed workday site": {
			"boardType": "WORKDAY", "boardToken": "initech", "companyName": "Initech", "companyDomain": "initech.com",
		},
		"career page on a private host": {
			"boardType": "CAREER_PAGE", "boardToken": "https://localhost/jobs",
			"companyName": "Initech", "companyDomain": "initech.com",
		},
	} {
		rejected := helpers.Do(
			t,
			fixture.router,
			http.MethodPost,
			jobSourcesPath,
			body,
			bearer(fixture.adminToken),
		)
		assert.Equal(
			t,
			http.StatusBadRequest,
			rejected.Recorder.Code,
			"%s: %s",
			name,
			string(rejected.Body),
		)
	}
}

func TestWorkdayBoardPollReadsEachNewPosting(t *testing.T) {
	fixture := newJobSourceFixture(t)
	source := fixture.createSource(t, map[string]any{
		"boardType":     "WORKDAY",
		"boardToken":    "Initech.wd1/Careers",
		"companyName":   "Initech",
		"companyDomain": "initech.com",
	})
	assert.Equal(t, "initech.wd1/careers", source.BoardToken)

	polled := fixture.poll(t, source.ID)
	require.Equal(t, http.StatusOK, polled.Recorder.Code, string(polled.Body))
	assert.Equal(t, 2, helpers.DecodeData[jobTypes.PollJobSourceResult](t, polled).Ingested)

	jobs := fixture.sourceJobs(t, source.ID)
	job := jobs["/job/Austin/Platform-Engineer_R1"]
	assert.Equal(t, "Work on Platform Engineer.", job.JobDescription)
	assert.Equal(
		t,
		"https://initech.wd1.myworkdayjobs.com/careers/job/Austin/Platform-Engineer_R1",
		job.ExternalURL,
	)
	assert.Equal(t, "WORKDAY", job.Source)

	again := fixture.poll(t, source.ID)
	require.Equal(t, http.StatusOK, again.Recorder.Code, string(again.Body))
	assert.Equal(t, 2, helpers.DecodeData[jobTypes.PollJobSourceResult](t, again).Unchanged)
}

func TestAdzunaSearchPollResolvesCompanyDomainsOnce(t *testing.T) {
	fixture := newJobSourceFixture(t)
	enrichment := helpers.NewEnrichmentStub(t)
	enrichment.Set(func(stub *helpers.EnrichmentStub) {
		stub.Organizations["Initrode Ltd"] = []map[string]any{
			{"name": "Initrode Labs", "primary_domain": "initrode-labs.com"},
			{"name": "Initrode Limited", "primary_domain": "initrode.com"},
		}
	})

	fixture.deps.EnrichmentProviders = providers.NewProviderRegistry(
		enrichment.Policies(),
		config.GetSignalHireCallbackURL(),
	)
	require.NoError(t, fixture.deps.Wire())
	fixture.router = routes.NewRoutes(fixture.deps)

	t.Cleanup(func() {
		fixture.deps.EnrichmentProviders = []providers.Provider{}
		require.NoError(t, fixture.deps.Wire())
	})

	source := fixture.createSource(t, map[string]any{
		"boardType":   "ADZUNA",
		"boardToken":  "IN",
		"searchQuery": "golang",
	})

	first := fixture.poll(t, source.ID)
	require.Equal(t, http.StatusOK, first.Recorder.Code, string(first.Body))
	firstResult := helpers.DecodeData[jobTypes.PollJobSourceResult](t, first)
	assert.Equal(t, 1, firstResult.Ingested)
	assert.Equal(t, 1, firstResult.Skipped, "a company whose domain is unknown cannot be ingested")

	assert.Equal(
		t,
		"initrode.com",
		fixture.sourceJobColumn(t, source.ID, "5922025200", "company_domain"),
	)
	assert.Equal(
		t,
		"Go Developer",
		fixture.sourceJobColumn(t, source.ID, "5922025200", "job_title"),
	)
	assert.Equal(t, "INR", fixture.sourceJobColumn(t, source.ID, "5922025200", "salary_currency"))
	assert.Equal(t, "ADZUNA", fixture.sourceJobColumn(t, source.ID, "5922025200", "source"))
	assert.Equal(t, 2, enrichment.Calls("apollo.organizations"))

	second := fixture.poll(t, source.ID)
	require.Equal(t, http.StatusOK, second.Recorder.Code, string(second.Body))
	assert.Equal(t, 2, enrichment.Calls("apollo.organizations"),
		"resolved and unresolvable companies are both cached")
	assert.Zero(t, helpers.DecodeData[jobTypes.PollJobSourceResult](t, second).Closed)
}

func TestCareerPageSitemapPollReadsJobPostingDataAndRespectsRobots(t *testing.T) {
	fixture := newJobSourceFixture(t)
	source := fixture.createSource(t, map[string]any{
		"boardType":     "CAREER_PAGE",
		"boardToken":    "https://careers.initech.com/sitemap.xml",
		"searchQuery":   "/jobs/",
		"companyName":   "Initech",
		"companyDomain": "initech.com",
	})

	polled := fixture.poll(t, source.ID)
	require.Equal(t, http.StatusOK, polled.Recorder.Code, string(polled.Body))
	assert.Equal(t, 2, helpers.DecodeData[jobTypes.PollJobSourceResult](t, polled).Ingested)
	assert.Equal(t, int32(2), fixture.stub.careerPageReads.Load(),
		"pages outside the filter or disallowed by robots.txt are not read")

	jobs := fixture.sourceJobs(t, source.ID)
	require.Len(t, jobs, 2)

	backend := jobs["https://careers.initech.com/jobs/1"]
	assert.Equal(t, "Backend Engineer", backend.JobTitle)
	assert.Equal(t, "Design APIs.", backend.JobDescription)
	assert.Equal(t, "Austin, TX, US", *backend.Location)
	assert.Equal(t, "DIRECT", backend.Source)
	assert.Equal(t, "190000.00",
		fixture.sourceJobColumn(t, source.ID, "https://careers.initech.com/jobs/1", "salary_max"))

	support := jobs["https://careers.initech.com/jobs/2"]
	assert.Equal(t, "Remote", *support.Location)
	assert.Equal(t, "https://careers.initech.com/jobs/2", support.ExternalURL)

	again := fixture.poll(t, source.ID)
	require.Equal(t, http.StatusOK, again.Recorder.Code, string(again.Body))
	assert.Equal(t, int32(2), fixture.stub.careerPageReads.Load(), "known pages are not read again")
}
