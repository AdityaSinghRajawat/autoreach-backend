package emailTemplates

import (
	"fmt"
	"html/template"
	"strings"
)

type UnsubscribePageData struct {
	ActionURL    string
	Unsubscribed bool
}

const unsubscribePageTemplate = `<!doctype html>
<html>
  <head><meta charset="utf-8"><meta name="viewport" content="width=device-width,initial-scale=1"><title>Unsubscribe</title></head>
  <body style="margin:0;padding:48px 16px;font-family:Helvetica,Arial,sans-serif;color:#18181b;text-align:center;">
    {{if .Unsubscribed}}
    <p style="font-size:16px;">You have been unsubscribed. You will not receive further emails from this sender.</p>
    {{else}}
    <p style="font-size:16px;">Stop receiving emails from this sender?</p>
    <form method="post" action="{{.ActionURL}}">
      <button type="submit" style="padding:12px 20px;font-size:14px;border:0;border-radius:8px;background:#18181b;color:#fff;cursor:pointer;">Unsubscribe</button>
    </form>
    {{end}}
  </body>
</html>
`

var unsubscribePage = template.Must(template.New("unsubscribePage").Parse(unsubscribePageTemplate))

func GetUnsubscribePageHTML(data UnsubscribePageData) (string, error) {
	var page strings.Builder
	if err := unsubscribePage.Execute(&page, data); err != nil {
		return "", fmt.Errorf("render unsubscribe page: %w", err)
	}

	return page.String(), nil
}
