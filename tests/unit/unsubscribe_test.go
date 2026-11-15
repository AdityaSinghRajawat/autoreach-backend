package unit

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	sharedTypes "github.com/AdityaSinghRajawat/autoreach-backend/pkg/customtypes/shared"
	"github.com/AdityaSinghRajawat/autoreach-backend/pkg/utils"
	"github.com/AdityaSinghRajawat/autoreach-backend/tests/setup"
)

func TestOutgoingEmailCarriesOneClickUnsubscribeHeaders(t *testing.T) {
	setup.Config(t)

	raw, err := utils.BuildEmailMessage(sharedTypes.OutgoingEmail{
		FromAddress:         "alex@example.com",
		ToAddress:           "sarah@acme.test",
		Subject:             "Hello",
		Body:                "Hi",
		MessageID:           "<one@example.com>",
		Date:                time.Now(),
		ListUnsubscribe:     "<https://api.test/u/abc?s=sig>, <mailto:alex@example.com?subject=unsubscribe>",
		OneClickUnsubscribe: true,
	})
	require.NoError(t, err)

	text := string(raw)
	assert.Contains(
		t,
		text,
		"List-Unsubscribe: <https://api.test/u/abc?s=sig>, <mailto:alex@example.com?subject=unsubscribe>\r\n",
	)
	assert.Contains(t, text, "List-Unsubscribe-Post: List-Unsubscribe=One-Click\r\n")
}

func TestUnsubscribeRepliesAreRecognisedByTheirSubject(t *testing.T) {
	setup.Config(t)

	for subject, expected := range map[string]bool{
		"unsubscribe":                      true,
		"Re: RE: Unsubscribe ":             true,
		"=?utf-8?q?Fwd=3A_unsubscribe?=":   true,
		"Re: Senior Go Engineer":           false,
		"please unsubscribe me later, Bob": false,
	} {
		parsed, err := utils.ParseInboundEmail(
			1,
			[]byte("From: sarah@acme.test\r\nSubject: "+subject+"\r\n\r\nbody"),
		)
		require.NoError(t, err)
		assert.Equal(t, expected, parsed.IsUnsubscribeRequest, subject)
	}
}
