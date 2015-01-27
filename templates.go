/*
 *  Bulldozer Framework
 *  Copyright (C) DesertBit
 */

package bulldozer

import (
	tr "code.desertbit.com/bulldozer/bulldozer/translate"

	"code.desertbit.com/bulldozer/bulldozer/log"
	"code.desertbit.com/bulldozer/bulldozer/sessions"
	"code.desertbit.com/bulldozer/bulldozer/settings"
	"code.desertbit.com/bulldozer/bulldozer/template"
	"code.desertbit.com/bulldozer/bulldozer/template/store"
	"fmt"
)

const (
	coreTemplatesUID = "blzCoreTemplates"

	// Core Template Names:
	errorTemplate    = "error" + settings.TemplateSuffix
	notFoundTemplate = "notfound" + settings.TemplateSuffix
)

var (
	TemplatesStore *store.Store
	coreTemplates  *template.Template
)

//###############//
//### Private ###//
//###############//

func loadTemplates() (err error) {
	// Create a new store and parse it.
	TemplatesStore, err = store.New(settings.Settings.TemplatesPath, settings.Settings.PagesPath)
	if err != nil {
		return err
	}
	TemplatesStore.Parse()

	// Load the core templates.
	coreTemplateFiles := []string{
		settings.GetCoreTemplatePath(errorTemplate),
		settings.GetCoreTemplatePath(notFoundTemplate),
	}
	coreTemplates, err = template.ParseFiles(coreTemplatesUID, coreTemplateFiles...)
	if err != nil {
		return fmt.Errorf("failed to load core templates: %v", err)
	}

	// Customize the templates.
	coreTemplates.LookupFatal(notFoundTemplate).
		AddStyleClass("bulldozer-page").
		AddStyleClass("bulldozer-not-found-page")

	coreTemplates.LookupFatal(errorTemplate).
		AddStyleClass("bulldozer-page").
		AddStyleClass("bulldozer-error-page")

	return nil
}

func execNotFoundTemplate(s *sessions.Session) (int, string, string) {
	// Execute the not found page
	out, _, _, err := coreTemplates.ExecuteTemplateToString(s, notFoundTemplate)
	if err != nil {
		return execErrorTemplate(s, err.Error())
	}

	return 404, out, tr.S("blz.page.notFound.pageTitle")
}

// execErrorTemplate executes the error template and shows the error message if the
// user is authenticated. The error message will be also logged.
// One optional boolean can be set. If set to false, the error message won't be logged.
func execErrorTemplate(s *sessions.Session, errorMessage string, vars ...bool) (int, string, string) {
	// Create the template data struct.
	data := struct {
		ErrorMessage string
	}{
		errorMessage,
	}

	if len(vars) <= 0 || vars[0] != false {
		// Log the error.
		log.L.Error(errorMessage)
	}

	// Custom template options.
	opts := template.ExecOpts{
		Data: data,
	}

	// Execute the template.
	out, _, _, err := coreTemplates.ExecuteTemplateToString(s, errorTemplate, opts)
	if err != nil {
		log.L.Error("failed to execute error core template: %v", err)
		return 500, "Internal Server Error", tr.S("blz.page.error.pageTitle")
	}

	return 500, out, tr.S("blz.page.error.pageTitle")
}
