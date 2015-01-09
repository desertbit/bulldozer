/*
 *  Bulldozer Framework
 *  Copyright (C) DesertBit
 */

package bulldozer

import (
	textTemplate "text/template"

	"code.desertbit.com/bulldozer/bulldozer/log"
	"code.desertbit.com/bulldozer/bulldozer/settings"
	"net/http"
)

var (
	robotsTxtTemplate *textTemplate.Template
	siteMapTemplate   *textTemplate.Template
)

func init() {
	// Parse the sitemap template.
	var err error
	siteMapTemplate, err = textTemplate.New("sitemap").Parse(siteMapTemplateText)
	if err != nil {
		log.L.Fatalf("sitemap template parsing error: %v", err)
	}

	// Parse the robots template.
	robotsTxtTemplate, err = textTemplate.New("robots").Parse(robotsTxtTemplateText)
	if err != nil {
		log.L.Fatalf("robotstxt template parsing error: %s", err.Error())
	}
}

func setWebCrawlerHtmlFuncs() {
	// Create the sitemap handler
	http.HandleFunc("/sitemap.xml", handleSitemapFunc)

	// Create the robots.txt handler
	http.HandleFunc("/robots.txt", handleRobotsTxtFunc)
}

func handleRobotsTxtFunc(w http.ResponseWriter, req *http.Request) {
	// Create the render data for the template
	data := struct {
		SiteURL             string
		DisallowedPagePaths []string
	}{
		settings.Settings.SiteUrl,
		settings.Settings.DisallowedRobotsUrls,
	}

	// Execute the template
	robotsTxtTemplate.Execute(w, data)
}

func handleSitemapFunc(w http.ResponseWriter, req *http.Request) {
	// Get the disallowed paths.
	disallowedPaths := settings.Settings.DisallowedRobotsUrls

	// Get all paths, without disallowed paths.
	var paths []string
	var found bool
	for _, path := range mainRouter.Paths() {
		found = false
		for _, dpath := range disallowedPaths {
			if path == dpath {
				found = true
				break
			}
		}

		if !found {
			paths = append(paths, path)
		}
	}

	// Create the render data for the template
	data := struct {
		SiteURL string
		URLs    []string
	}{
		settings.Settings.SiteUrl,
		paths,
	}

	// Execute the template
	siteMapTemplate.Execute(w, data)
}

// The sitemap template text
const siteMapTemplateText = `<?xml version="1.0" encoding="UTF-8"?>
<urlset xmlns="http://www.sitemaps.org/schemas/sitemap/0.9">
	{{range $url := $.URLs}}
	<url>
		<loc>{{$.SiteURL}}{{$url}}</loc>
	</url>
	{{end}}
</urlset>`

// The robotstxt template text
const robotsTxtTemplateText = `User-agent: *
{{if $.DisallowedPagePaths}}{{range $path := $.DisallowedPagePaths}}Disallow: {{$path}}${{end}}{{end}}
Sitemap: {{$.SiteURL}}/sitemap.xml`
