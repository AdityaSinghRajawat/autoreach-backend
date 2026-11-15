package unit

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/AdityaSinghRajawat/autoreach-backend/config"
	jobModel "github.com/AdityaSinghRajawat/autoreach-backend/pkg/models/jobManagement"
	"github.com/AdityaSinghRajawat/autoreach-backend/pkg/services/applicationManagement/appliers"
	"github.com/AdityaSinghRajawat/autoreach-backend/tests/setup"
)

func TestAppliersRecogniseGreenhouseAndLeverJobPages(t *testing.T) {
	setup.Config(t)

	registry := appliers.NewApplierRegistry(appliers.RegistryOptions{
		GreenhouseKeys: map[string]string{"acme": "key"},
		LeverKeys:      map[string]string{},
	})

	for url, expected := range map[string]string{
		"https://boards.greenhouse.io/acme/jobs/10293":     "GREENHOUSE",
		"https://job-boards.greenhouse.io/acme/jobs/10293": "GREENHOUSE",
		"https://jobs.lever.co/acme/abc-123":               "LEVER",
		"https://jobs.eu.lever.co/acme/abc-123/apply":      "LEVER",
		"https://jobs.lever.co/acme/apply":                 "",
		"https://boards.greenhouse.io/acme":                "",
		"https://careers.acme.com/jobs/1":                  "",
	} {
		applier := appliers.Find(registry, &jobModel.JobPosting{ExternalURL: url})
		if expected == "" {
			assert.Nil(t, applier, url)

			continue
		}

		if assert.NotNil(t, applier, url) {
			assert.Equal(t, expected, applier.Code(), url)
		}
	}

	greenhouse := appliers.Find(
		registry,
		&jobModel.JobPosting{ExternalURL: "https://boards.greenhouse.io/acme/jobs/1"},
	)
	assert.True(
		t,
		greenhouse.CanSubmit(
			&jobModel.JobPosting{ExternalURL: "https://boards.greenhouse.io/ACME/jobs/1"},
		),
	)
	assert.False(
		t,
		greenhouse.CanSubmit(
			&jobModel.JobPosting{ExternalURL: "https://boards.greenhouse.io/other/jobs/1"},
		),
	)
}

func TestFormQuestionsMapToProfileFieldsByLabel(t *testing.T) {
	for label, field := range map[string]string{
		"LinkedIn Profile": config.ProfileFieldLinkedin,
		"Will you now or in the future require visa sponsorship?":  config.ProfileFieldSponsorship,
		"Are you legally authorized to work in the United States?": config.ProfileFieldWorkAuth,
		"What are your salary expectations?":                       config.ProfileFieldSalary,
		"Why do you want to work here?":                            "",
	} {
		assert.Equal(t, field, config.ProfileFieldForLabel(label), label)
	}
}
