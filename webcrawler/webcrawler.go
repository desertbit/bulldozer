/*
 *  Bulldozer Framework
 *  Copyright (C) DesertBit
 */

package webcrawler

import (
	textTemplate "text/template"

	"net/http"
	"strings"

	"code.desertbit.com/bulldozer/bulldozer/log"
	"code.desertbit.com/bulldozer/bulldozer/settings"
)

var (
	robotsTxtTemplate *textTemplate.Template
	sitemapTemplate   *textTemplate.Template
)

var (
	sitemapPaths    []string
	disallowedPaths []string
)

func init() {
	// Parse the sitemap template.
	var err error
	sitemapTemplate, err = textTemplate.New("sitemap").Parse(sitemapTemplateText)
	if err != nil {
		log.L.Fatalf("sitemap template parsing error: %v", err)
	}

	// Parse the robots template.
	robotsTxtTemplate, err = textTemplate.New("robots").Parse(robotsTxtTemplateText)
	if err != nil {
		log.L.Fatalf("robotstxt template parsing error: %s", err.Error())
	}

	// Create the sitemap handler
	http.HandleFunc("/sitemap.xml", handleSitemapFunc)

	// Create the robots.txt handler
	http.HandleFunc("/robots.txt", handleRobotsTxtFunc)
}

//##############//
//### Public ###//
//##############//

// AddSitemapPath adds the path to the sitemap.
// The path should not contain the site URL. This is prepended automatically.
func AddSitemapPath(path string) {
	// Normalize the path.
	path = toPath(path)

	// Check if already present in slice.
	for _, p := range sitemapPaths {
		if p == path {
			log.L.Warning("webcrawler: the sitemap path '%s' is already in the list. Duplicate call...", path)
			return
		}
	}

	// Add the path to the slice.
	sitemapPaths = append(sitemapPaths, path)
}

// RemoveSitemapPath removes the sitemap path from the list if present.
func RemoveSitemapPath(path string) {
	// Normalize the path.
	path = toPath(path)

	// Remove the path from the sitemap slice if present.
	for i, p := range sitemapPaths {
		if p == path {
			sitemapPaths = append(sitemapPaths[:i], sitemapPaths[i+1:]...)
			break
		}
	}
}

// AddDisallowedPath adds the path to the robots disallowed paths.
// The path should not contain the site URL. This is prepended automatically.
func AddDisallowedPath(path string) {
	// Normalize the path.
	path = toPath(path)

	// If the path does not end with a slash, then append the '$' symbol.
	if !strings.HasSuffix(path, "/") {
		path += "$"
	}

	// Check if already present in slice.
	for _, p := range disallowedPaths {
		if p == path {
			log.L.Warning("webcrawler: the robots disallowed path '%s' is already in the list. Duplicate call...", path)
			return
		}
	}

	// Add the path to the slice.
	disallowedPaths = append(disallowedPaths, path)
}

//###############//
//### Private ###//
//###############//

// toPath normalizes the path.
func toPath(path string) string {
	path = strings.TrimSpace(strings.ToLower(path))

	if !strings.HasPrefix(path, "/") {
		path += "/" + path
	}

	return path
}

func handleRobotsTxtFunc(w http.ResponseWriter, req *http.Request) {
	// Create the render data for the template
	data := struct {
		SiteURL             string
		DisallowedPagePaths []string
	}{
		settings.Settings.SiteUrl,
		disallowedPaths,
	}

	// Execute the template
	robotsTxtTemplate.Execute(w, data)
}

func handleSitemapFunc(w http.ResponseWriter, req *http.Request) {
	// Get all sitemap urls.
	urls := append(sitemapPaths, callSitemapHooks()...)

	// Create the render data for the template
	data := struct {
		SiteURL string
		URLs    []string
	}{
		settings.Settings.SiteUrl,
		urls,
	}

	// Execute the template
	sitemapTemplate.Execute(w, data)
}

// The sitemap template text
const sitemapTemplateText = `<?xml version="1.0" encoding="UTF-8"?>
<urlset xmlns="http://www.sitemaps.org/schemas/sitemap/0.9">
	{{range $url := $.URLs}}
	<url>
		<loc>{{$.SiteURL}}{{$url}}</loc>
	</url>
	{{end}}
</urlset>`

// The robotstxt template text
const robotsTxtTemplateText = `User-agent: *
{{if $.DisallowedPagePaths}}{{range $path := $.DisallowedPagePaths}}Disallow: {{$path}}{{end}}{{end}}
Sitemap: {{$.SiteURL}}/sitemap.xml`
