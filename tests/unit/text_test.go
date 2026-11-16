package unit

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/AdityaSinghRajawat/autoreach-backend/pkg/utils"
)

func TestHTMLToTextKeepsLineBreaksAndUnescapes(t *testing.T) {
	cases := map[string]string{
		"<p>Hello <b>world</b></p><p>Next</p>":        "Hello world\nNext",
		"&lt;p&gt;Fish &amp;amp; chips&lt;/p&gt;":     "Fish & chips",
		"<ul><li>Go</li><li>SQL</li></ul>":            "Go\nSQL",
		"Line one<br/>Line two<br><br><br>Line three": "Line one\nLine two\n\nLine three",
		"plain text": "plain text",
		"<div>  spaced   out  </div>\n\n\n<div>end</div>": "spaced out\n\nend",
	}

	for markup, expected := range cases {
		assert.Equal(t, expected, utils.HTMLToText(markup), markup)
	}
}
