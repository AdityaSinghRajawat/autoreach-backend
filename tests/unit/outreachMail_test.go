package unit

import (
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	sharedTypes "github.com/AdityaSinghRajawat/autoreach-backend/pkg/customtypes/shared"
	"github.com/AdityaSinghRajawat/autoreach-backend/pkg/emailTemplates"
	"github.com/AdityaSinghRajawat/autoreach-backend/pkg/utils"
)

func TestPlaceholdersRenderKnownValuesAndReportUnknownOnes(t *testing.T) {
	text := "Hi {{ contactFirstName }}, about {{jobTitle}} {{favourite}}"

	assert.Equal(
		t,
		[]string{"favourite"},
		utils.UnknownPlaceholders(text, []string{"contactFirstName", "jobTitle"}),
	)
	assert.Equal(
		t,
		"Hi Sarah, about Go Engineer ",
		utils.RenderPlaceholders(
			text,
			map[string]string{"contactFirstName": "Sarah", "jobTitle": "Go Engineer"},
		),
	)
}

func TestSendWindowWaitsForTheNextWeekdayMorningInTheCandidatesZone(t *testing.T) {
	kolkata, err := time.LoadLocation("Asia/Kolkata")
	require.NoError(t, err)

	fridayEvening := time.Date(2026, 11, 13, 19, 0, 0, 0, kolkata)
	next := utils.NextSendTime(fridayEvening, kolkata, 9, 17, false)
	assert.Equal(t, time.Date(2026, 11, 16, 9, 0, 0, 0, kolkata).UTC(), next, "skips the weekend")

	mondayNoon := time.Date(2026, 11, 16, 12, 0, 0, 0, kolkata)
	assert.Equal(
		t,
		mondayNoon,
		utils.NextSendTime(mondayNoon, kolkata, 9, 17, false),
		"inside the window",
	)
}

func TestComposedEmailThreadsFollowUpsAndParsesBackAsAReply(t *testing.T) {
	raw, err := utils.BuildEmailMessage(sharedTypes.OutgoingEmail{
		FromAddress: "alex@example.com",
		FromName:    "Alex Candidate",
		ToAddress:   "sarah@acme.test",
		Subject:     "Re: Go role – follow-up",
		Body:        "Hello Sarah,\nfollowing up.",
		MessageID:   "<two@example.com>",
		InReplyTo:   "<one@example.com>",
		Date:        time.Date(2026, 11, 16, 9, 0, 0, 0, time.UTC),
	})
	require.NoError(t, err)

	text := string(raw)
	assert.Contains(t, text, "In-Reply-To: <one@example.com>\r\n")
	assert.Contains(t, text, "References: <one@example.com>\r\n")
	assert.Contains(t, text, "Subject: =?utf-8?q?")

	parsed, err := utils.ParseInboundEmail(7, raw)
	require.NoError(t, err)
	assert.Equal(t, "<two@example.com>", parsed.MessageID)
	assert.Equal(t, []string{"<one@example.com>"}, parsed.ReferencedIDs)
	assert.False(t, parsed.IsBounce)
	assert.False(t, parsed.IsAutoReply)
}

func TestAutoRepliesAndBounceReportsAreRecognised(t *testing.T) {
	autoReply, err := utils.ParseInboundEmail(
		1,
		[]byte("From: sarah@acme.test\r\nAuto-Submitted: auto-replied\r\n"+
			"In-Reply-To: <one@example.com>\r\n\r\nOut of office"),
	)
	require.NoError(t, err)
	assert.True(t, autoReply.IsAutoReply)

	bounce, err := utils.ParseInboundEmail(2, []byte(strings.Join([]string{
		"From: MAILER-DAEMON@mx.acme.test",
		"Content-Type: multipart/report; report-type=delivery-status; boundary=\"b\"",
		"",
		"--b",
		"Content-Type: message/delivery-status",
		"",
		"Final-Recipient: rfc822; Sarah@acme.test",
		"--b",
		"Content-Type: text/rfc822-headers",
		"",
		"Message-ID: <one@example.com>",
		"--b--",
	}, "\r\n")))
	require.NoError(t, err)
	assert.True(t, bounce.IsBounce)
	assert.Equal(t, []string{"<one@example.com>"}, bounce.BouncedMessageIDs)
	assert.Equal(t, []string{"sarah@acme.test"}, bounce.BouncedAddresses)
}

func TestOutreachEmailTemplateEscapesTextAndRoutesLinksThroughTracking(t *testing.T) {
	data := emailTemplates.NewOutreachEmailData(
		"See <b> https://acme.test/jobs?id=1&x=2\nThanks",
		func(link string) string { return "https://api.test/t/c/abc?u=" + url.QueryEscape(link) },
		"https://api.test/t/o/abc",
	)

	body, err := emailTemplates.GetOutreachEmailHTML(data)
	require.NoError(t, err)

	assert.Contains(t, body, "See &lt;b&gt; ")
	assert.Contains(t, body, `<img src="https://api.test/t/o/abc"`)
	assert.Contains(t, body, "<br>Thanks")
	assert.Contains(
		t,
		body,
		`href="https://api.test/t/c/abc?u=https%3A%2F%2Facme.test%2Fjobs%3Fid%3D1%26x%3D2"`,
	)
	assert.Contains(t, body, ">https://acme.test/jobs?id=1&amp;x=2</a>")
}

func TestSignaturesOnlyVerifyTheSignedValues(t *testing.T) {
	key := utils.DeriveKey([]byte("secret"), "label")
	signature := utils.Sign(key, "abc", "https://acme.test")

	assert.True(t, utils.VerifySignature(key, signature, "abc", "https://acme.test"))
	assert.False(t, utils.VerifySignature(key, signature, "abc", "https://evil.test"))
}
