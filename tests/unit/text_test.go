package unit

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/AdityaSinghRajawat/autoreach-backend/config"
	"github.com/AdityaSinghRajawat/autoreach-backend/pkg/utils"
	"github.com/AdityaSinghRajawat/autoreach-backend/tests/setup"
)

func TestHTMLToTextKeepsLineBreaksAndUnescapes(t *testing.T) {
	cases := map[string]string{
		"<p>Hello <b>world</b></p><p>Next</p>":        "Hello world\nNext",
		"&lt;p&gt;Fish &amp;amp; chips&lt;/p&gt;":     "Fish & chips",
		"<ul><li>Go</li><li>SQL</li></ul>":            "Go\nSQL",
		"Line one<br/>Line two<br><br><br>Line three": "Line one\nLine two\n\nLine three",
		"plain text":                                      "plain text",
		"Design <b>APIs</b>. <em>Now</em>":                "Design APIs. Now",
		"<table><tr><td>Go</td><td>SQL</td></tr></table>": "Go SQL",
		"<div>  spaced   out  </div>\n\n\n<div>end</div>": "spaced out\n\nend",
	}

	for markup, expected := range cases {
		assert.Equal(t, expected, utils.HTMLToText(markup), markup)
	}
}

func TestWebFetcherRefusesPrivateAddresses(t *testing.T) {
	setup.Config(t)

	fetcher := utils.NewWebFetcher(5*time.Second, 1024, "AutoReachBot/1.0", nil)

	for _, target := range []string{
		"http://127.0.0.1:1/",
		"http://169.254.169.254/latest",
		"http://10.0.0.1/",
		"http://100.64.0.1/",
		"http://198.18.0.1/",
		"http://[64:ff9b::a00:1]/",
		"http://[2002:a00:1::]/",
	} {
		_, err := fetcher.Fetch(context.Background(), target)
		assert.ErrorIs(t, err, config.ErrPrivateNetworkAddress, target)
	}
}

func TestLogsNeverCarryEmailAddresses(t *testing.T) {
	setup.Config(t)

	assert.Equal(
		t,
		"smtp rcpt: 550 5.1.1 <[redacted]@acme.com> user unknown",
		utils.RedactEmailAddresses("smtp rcpt: 550 5.1.1 <riley.lee+jobs@acme.com> user unknown"),
	)
	assert.Equal(t, "no address here", utils.RedactEmailAddresses("no address here"))
}
