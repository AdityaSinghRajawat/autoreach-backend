package shared

import "encoding/xml"

type SitemapDocument struct {
	XMLName  xml.Name          `xml:""`
	Pages    []SitemapLocation `xml:"url"`
	Sitemaps []SitemapLocation `xml:"sitemap"`
}

type SitemapLocation struct {
	Location string `xml:"loc"`
}
