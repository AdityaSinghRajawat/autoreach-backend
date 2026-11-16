package utils

import (
	"encoding/json"
	"encoding/xml"
	"html"
	"regexp"
	"strings"

	sharedTypes "github.com/AdityaSinghRajawat/autoreach-backend/pkg/customtypes/shared"
)

var jsonLDScriptPatternI = regexp.MustCompile(
	`(?is)<script[^>]*type\s*=\s*["']application/ld\+json["'][^>]*>(.*?)</script>`,
)

// ExtractJSONLDObjects returns every schema.org object of the given type in a page's JSON-LD,
// including ones nested in arrays and @graph.
func ExtractJSONLDObjects(page []byte, typeName string) []map[string]any {
	objects := []map[string]any{}

	for _, match := range jsonLDScriptPatternI.FindAllSubmatch(page, -1) {
		var document any
		if json.Unmarshal([]byte(strings.TrimSpace(string(match[1]))), &document) != nil {
			continue
		}

		collectJSONLD(document, typeName, &objects)
	}

	return objects
}

func collectJSONLD(node any, typeName string, objects *[]map[string]any) {
	switch value := node.(type) {
	case []any:
		for _, item := range value {
			collectJSONLD(item, typeName, objects)
		}
	case map[string]any:
		if hasJSONLDType(value["@type"], typeName) {
			*objects = append(*objects, value)
		}

		if graph, found := value["@graph"]; found {
			collectJSONLD(graph, typeName, objects)
		}
	}
}

func hasJSONLDType(declared any, typeName string) bool {
	switch value := declared.(type) {
	case string:
		return value == typeName
	case []any:
		for _, item := range value {
			if text, isText := item.(string); isText && text == typeName {
				return true
			}
		}
	}

	return false
}

// ParseSitemap reads a sitemap or sitemap index; ok is false for anything else.
func ParseSitemap(body []byte) (pages, sitemaps []string, ok bool) {
	document := sharedTypes.SitemapDocument{}
	if xml.Unmarshal(body, &document) != nil {
		return nil, nil, false
	}

	if name := document.XMLName.Local; name != "urlset" && name != "sitemapindex" {
		return nil, nil, false
	}

	for _, page := range document.Pages {
		if location := strings.TrimSpace(html.UnescapeString(page.Location)); location != "" {
			pages = append(pages, location)
		}
	}

	for _, sitemap := range document.Sitemaps {
		if location := strings.TrimSpace(html.UnescapeString(sitemap.Location)); location != "" {
			sitemaps = append(sitemaps, location)
		}
	}

	return pages, sitemaps, true
}
