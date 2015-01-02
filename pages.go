/*
 *  Bulldozer Framework
 *  Copyright (C) DesertBit
 */

package bulldozer

import (
	"code.desertbit.com/bulldozer/bulldozer/settings"
	"code.desertbit.com/bulldozer/bulldozer/template"
	"fmt"
)

const (
	pagesTemplatesUID = "pages"
)

var (
	pageTemplates *template.Template

	pagesEvents = new(PagesEvents)
)

//###############//
//### Private ###//
//###############//

type PagesEvents struct{}

func (p *PagesEvents) EventHallo(c *template.Context, arg1 string, arg2 int) {
	fmt.Println(arg1, arg2, p, c)
	c.TriggerEvent("hallo", "1", 2)
	c.Update(nil)
}

// TODO: Lock mutex?
// TODO: template.ReleaseSessionEvents(c.Session())
// TODO: Separate unique ID in sub or global Template

func parsePages() error {
	// Create the pattern string
	pattern := settings.Settings.PagesPath + "/" + "*" + settings.TemplateSuffix

	// Parse the templates files in the pages directory
	templates, err := template.ParseGlob(pagesTemplatesUID, pattern)
	if err != nil {
		return fmt.Errorf("failed to parse page templates: %v", err)
	}

	templates.RegisterEvents(pagesEvents)

	// Set the pageTemplates pointer on success
	pageTemplates = templates

	return nil
}
