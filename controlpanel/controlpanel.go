/*
 *  Bulldozer Framework
 *  Copyright (C) DesertBit
 */

package controlpanel

import (
	tr "code.desertbit.com/bulldozer/bulldozer/translate"

	"code.desertbit.com/bulldozer/bulldozer/mux"
	"code.desertbit.com/bulldozer/bulldozer/sessions"
	"code.desertbit.com/bulldozer/bulldozer/settings"
	"code.desertbit.com/bulldozer/bulldozer/template"
	"code.desertbit.com/bulldozer/bulldozer/templates"

	"fmt"
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
	settings.Settings.AddDisallowedRobotsUrls(PageUrl)

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

//###############//
//### Private ###//
//###############//

func routePage(s *sessions.Session, req *mux.Request) {
	// Create the page render data.
	data := struct {
		PageUrl string
		Pages   Pages
	}{
		PageUrl: PageUrl,
		Pages:   pages,
	}

	// Get the requested control panel template and execute it.
	/*if len(rData.RestPath) > 0 {
		id := rData.RestPath

		// Check if the ID is valid and a page exists.
		page, ok := pages[id]
		if !ok {
			return nil, fmt.Errorf("control panel: client requested an invalid control panel page with ID: %s", id)
		}

		if page.Template == nil {
			return nil, fmt.Errorf("control panel: client requested the control panel page with ID '%s', but the template is nil!", id)
		}
	}*/

	// Set the title.
	req.Title = tr.S("bud.controlpanel.pageTitle")

	// Create the template execute options.
	opts := template.ExecOpts{
		Data: data,
	}

	// Execute the template page.
	var err error
	req.Body, _, _, err = templates.Templates.ExecuteTemplateToString(s, templateName, opts)
	if err != nil {
		req.Error(err)
	}
}
