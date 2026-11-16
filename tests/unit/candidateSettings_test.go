package unit

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"

	candidateTypes "github.com/AdityaSinghRajawat/autoreach-backend/pkg/customtypes/candidateManagement"
	candidateModel "github.com/AdityaSinghRajawat/autoreach-backend/pkg/models/candidateManagement"
	"github.com/AdityaSinghRajawat/autoreach-backend/tests/setup"
)

func TestSettingsUpdateMergesAndNeverTouchesTheApplicationProfile(t *testing.T) {
	setup.Config(t)

	existing := candidateModel.NewCandidate()
	existing.Settings = json.RawMessage(
		`{"emailTracking":true,"theme":"dark","applicationProfile":{"phone":"+1 555 0100"}}`,
	)

	request := candidateTypes.UpdateCandidateRequest{
		Settings: json.RawMessage(
			`{"emailTracking":false,"theme":null,"applicationProfile":"x","lang":"en"}`,
		),
	}

	assert.JSONEq(t,
		`{"emailTracking":false,"lang":"en","applicationProfile":{"phone":"+1 555 0100"}}`,
		string(request.ToCandidateModel(existing).Settings),
		"keys are merged, null removes one, and the application profile keeps its own endpoint",
	)
}
