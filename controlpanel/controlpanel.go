/*
 *  Bulldozer Framework
 *  Copyright (C) DesertBit
 */

package controlpanel

import (
	ht "html/template"

	"github.com/desertbit/bulldozer/auth"
	"github.com/desertbit/bulldozer/mux"
	"github.com/desertbit/bulldozer/sessions"
	"github.com/desertbit/bulldozer/template"
	"github.com/desertbit/bulldozer/templates"
	"github.com/desertbit/bulldozer/webcrawler"

	"fmt"
	"strings"
)

const (
	PageUrl = "/controlpanel"

	templateName = "bud/controlpanel/controlpanel"
)

//##############//
//### Public ###//
//##############//

// Init initializes this package. This is managed by the bulldozer main package.
func Init() (err error) {
	// Disallow the control panel page to be indexed.
	webcrawler.AddDisallowedPath(PageUrl + "/")

	// Obtain the control center template.
	t := templates.Templates.Lookup(templateName)
	if t == nil {
		return fmt.Errorf("failed to lookup control center template!")
	}

	// Set the custom ID.
	t.SetStaticDomID("bud-ctrl")

	// Add the control panel routes.
	mux.Route(PageUrl, routePage)
	mux.Route(PageUrl+"/*", routePage)

	return nil
}

// IsCurrentPage returns a boolean whenever the control panel
// is the current page.
func IsCurrentPage(s *sessions.Session) bool {
	return strings.HasPrefix(s.CurrentPath(), PageUrl)
}

//###############//
//### Private ###//
//###############//

type item struct {
	Url      string
	Title    string
	Icon     string
	IsActive bool
}

func routePage(s *sessions.Session, req *mux.Request) {
	// Get the user
	u := auth.GetUser(s)
	if u == nil {
		req.Error(fmt.Errorf("control panel: access not allowed"))
		return
	}

	// Get the requested control panel page ID.
	id := req.RouteData.RestPath
	if len(id) == 0 {
		if len(pages) == 0 {
			req.Error(fmt.Errorf("control panel: no pages"))
			return
		}

		// If no ID is set, then use the first page as default page.
		id = pages[0].ID
	}

	// The current page pointer.
	var currentPage *Page

	// Create a item slice of all items which the user has access to.
	var items []item
	var isActive bool
	for _, page := range pages {
		// Set the active flag if this is the current active item.
		// Also set the current page pointer.
		isActive = false
		if page.ID == id {
			isActive = true
			currentPage = page
		}

		// Skip if the user has no access permission.
		if !u.IsInGroup(page.AuthGroups...) {
			continue
		}

		// Create a new item.
		i := item{
			Url:      PageUrl + "/" + page.ID,
			Title:    page.Title,
			Icon:     page.Icon,
			IsActive: isActive,
		}

		items = append(items, i)
	}

	// Create the page render data.
	data := struct {
		Items        []item
		Body         ht.HTML
		CurrentTitle string
	}{
		Items: items,
	}

	// Śhow the not found error page if the current page is nil.
	if currentPage == nil {
		req.NotFoundError()
		return
	}

	// Check if the user is allowed to access the current page.
	if !u.IsInGroup(currentPage.AuthGroups...) {
		req.Error(fmt.Errorf("control panel: access not allowed"))
		return
	}

	// Check if the template is not nil.
	if currentPage.Template == nil {
		req.Error(fmt.Errorf("control panel: client requested the control panel page with ID '%s', but the template is nil!", id))
		return
	}

	// Execute the control panel page template.
	body, _, err := currentPage.Template.ExecuteToString(s)
	if err != nil {
		req.Error(fmt.Errorf("control panel: client requested the control panel page with ID '%s', but the template execution failed: %v", id, err))
		return
	}

	// Unescape the body.
	data.Body = ht.HTML(body)

	// Set the control panel and page title.
	req.Title = currentPage.Title
	data.CurrentTitle = currentPage.Title

	// Create the template execute options.
	opts := template.ExecOpts{
		Data: data,
	}

	// Execute the template page.
	req.Body, _, _, err = templates.Templates.ExecuteTemplateToString(s, templateName, opts)
	if err != nil {
		req.Error(err)
	}
}
