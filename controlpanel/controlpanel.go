/*
 *  Bulldozer Framework
 *  Copyright (C) DesertBit
 */

package controlpanel

import (
	tr "code.desertbit.com/bulldozer/bulldozer/translate"

	"code.desertbit.com/bulldozer/bulldozer/settings"
	"code.desertbit.com/bulldozer/bulldozer/template"

	"fmt"
)

const (
	PageUrl = "/controlpanel"

	pageUID           = "budControlC"
	templatesDir      = "controlpanel/"
	templateName      = "controlpanel"
	templateNamespace = "bulldozer/"
)

var (
	backend bulldozerBackend
)

//###################################//
//### Bulldozer backend interface ###//
//###################################//

type bulldozerBackend interface {
	ParsePageTemplate(templateName string, path string) (*template.Template, error)
	RoutePage(path string, pageTitle string, pageTemplate string, UID string)
}

//##############//
//### Public ###//
//##############//

// Init initializes this package. This is managed by the bulldozer main package.
func Init(b bulldozerBackend) (err error) {
	// Set the backend.
	backend = b

	// Create the template file path.
	file := settings.LookupInternalTemplatePath(templatesDir + templateName + settings.TemplateExtension)

	// Create the template name.
	name := templateNamespace + templateName

	// Parse and add the control panel template page.
	t, err := backend.ParsePageTemplate(name, file)
	if err != nil {
		return fmt.Errorf("failed to parse control panel page template: %v", err)
	}

	// Set the custom ID.
	t.SetStaticDomID("bud-ctrl")

	// Add the control panel route.
	backend.RoutePage(PageUrl, tr.S("bud.controlpanel.pageTitle"), name, pageUID)

	return nil
}
