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
	"fmt"
)

const (
	templatesUID     = "budTmpls"
	coreTemplatesUID = "budCoreTmpls"

	// Core Template Names:
	errorTemplate    = "error"
	notFoundTemplate = "notfound"
)

var (
	Templates *template.Template

	templatesParseError error
	coreTemplates       *template.Template
)

//###############//
//### Private ###//
//###############//

func loadTemplates() (err error) {
	// Create a slice of directories which should be excluded.
	excludeDirs := []string{
		settings.Settings.BulldozerTemplatesPath,
	}

	// Load all the template files recursivly.
	Templates, err = template.ParseRec(templatesUID, settings.Settings.TemplatesPath, excludeDirs...)
	if err != nil && err != template.ErrNoFilesFound && err != template.ErrPatternMatchesNoFiles {
		// Just store the templates parse error.
		// The application startup should not be interrupted...
		templatesParseError = fmt.Errorf("failed to load templates: %v", err)
		log.L.Error(templatesParseError.Error())
	}

	// Load the core templates.
	coreTemplateFiles := []string{
		settings.LookupInternalTemplatePath(errorTemplate + settings.TemplateExtension),
		settings.LookupInternalTemplatePath(notFoundTemplate + settings.TemplateExtension),
	}
	coreTemplates, err = template.ParseFiles(coreTemplatesUID, coreTemplateFiles...)
	if err != nil {
		return fmt.Errorf("failed to load core templates: %v", err)
	}

	// Customize the templates.
	coreTemplates.LookupFatal(notFoundTemplate).
		AddStyleClass("bud-page").
		AddStyleClass("bud-not-found-page")

	coreTemplates.LookupFatal(errorTemplate).
		AddStyleClass("bud-page").
		AddStyleClass("bud-error-page")

	return nil
}

func execNotFoundTemplate(s *sessions.Session) (int, string, string) {
	// Execute the not found page
	out, _, _, err := coreTemplates.ExecuteTemplateToString(s, notFoundTemplate)
	if err != nil {
		return execErrorTemplate(s, err.Error())
	}

	return 404, out, tr.S("bud.page.notFound.pageTitle")
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
		return 500, "Internal Server Error", tr.S("bud.page.error.pageTitle")
	}

	return 500, out, tr.S("bud.page.error.pageTitle")
}
