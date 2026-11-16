package unit

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/AdityaSinghRajawat/autoreach-backend/pkg/utils"
	"github.com/AdityaSinghRajawat/autoreach-backend/tests/setup"
)

func TestTransportErrorsNeverCarryTheQueryString(t *testing.T) {
	setup.Config(t)

	client := utils.NewHttpClient("http://127.0.0.1:1", 2*time.Second)
	result := map[string]any{}

	err := client.Get(
		context.Background(),
		"/jobs/in/search/1?app_id=id&app_key=super-secret-key",
		nil,
		&result,
	)

	require.Error(t, err)
	assert.NotContains(
		t,
		err.Error(),
		"super-secret-key",
		"API keys passed as query parameters must not reach logs",
	)
	assert.Contains(t, err.Error(), "/jobs/in/search/1", "the path stays useful for debugging")
}
