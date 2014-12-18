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

var (
	pageTemplates *template.Template
)

//###############//
//### Private ###//
//###############//

// TODO: Lock mutex?

func parsePages() error {
	// Create the pattern string
	pattern := settings.Settings.PagesPath + "/" + "*" + settings.TemplateSuffix

	// Parse the templates files in the pages directory
	templates, err := template.ParseGlob(pattern)
	if err != nil {
		return fmt.Errorf("failed to parse page templates: %v", err)
	}

	// Set the pageTemplates pointer on success
	pageTemplates = templates

	return nil
}
