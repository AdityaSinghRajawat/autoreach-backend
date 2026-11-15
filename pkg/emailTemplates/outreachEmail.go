package emailTemplates

import (
	"fmt"
	"html/template"
	"regexp"
	"strings"
)

type OutreachEmailSegment struct {
	Text string
	Href string
}

type OutreachEmailData struct {
	Lines        [][]OutreachEmailSegment
	OpenPixelURL string
}

const outreachEmailTemplate = `<!doctype html>
<html>
  <body style="margin:0;padding:0;font-family:Helvetica,Arial,sans-serif;font-size:14px;line-height:22px;color:#18181b;">
    <p style="margin:0;">
      {{- range $index, $line := .Lines}}{{if $index}}<br>{{end}}{{range $line}}{{if .Href}}<a href="{{.Href}}">{{.Text}}</a>{{else}}{{.Text}}{{end}}{{end}}{{end -}}
    </p>
    {{if .OpenPixelURL}}<img src="{{.OpenPixelURL}}" width="1" height="1" alt="">{{end}}
  </body>
</html>
`

var (
	outreachEmail = template.Must(template.New("outreachEmail").Parse(outreachEmailTemplate))
	linkPattern   = regexp.MustCompile(`https?://[^\s<>"]+`)
)

func NewOutreachEmailData(
	plainBody string,
	trackLink func(link string) string,
	openPixelURL string,
) OutreachEmailData {
	data := OutreachEmailData{OpenPixelURL: openPixelURL}

	for _, line := range strings.Split(plainBody, "\n") {
		segments := []OutreachEmailSegment{}
		last := 0

		for _, match := range linkPattern.FindAllStringIndex(line, -1) {
			segments = append(
				segments,
				OutreachEmailSegment{Text: line[last:match[0]]},
				OutreachEmailSegment{
					Text: line[match[0]:match[1]],
					Href: trackLink(line[match[0]:match[1]]),
				},
			)
			last = match[1]
		}

		data.Lines = append(data.Lines, append(segments, OutreachEmailSegment{Text: line[last:]}))
	}

	return data
}

func GetOutreachEmailHTML(data OutreachEmailData) (string, error) {
	var body strings.Builder
	if err := outreachEmail.Execute(&body, data); err != nil {
		return "", fmt.Errorf("render outreach email: %w", err)
	}

	return body.String(), nil
}
