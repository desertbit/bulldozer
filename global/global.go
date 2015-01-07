/*
 *  Bulldozer Framework
 *  Copyright (C) DesertBit
 */

package global

import (
	"code.desertbit.com/bulldozer/bulldozer/log"
	"code.desertbit.com/bulldozer/bulldozer/sessions"
	"code.desertbit.com/bulldozer/bulldozer/settings"
	"code.desertbit.com/bulldozer/bulldozer/template/store"
)

const (
	// Template Names
	ErrorTemplate            = "error" + settings.TemplateSuffix
	NotFoundTemplate         = "notfound" + settings.TemplateSuffix
	LoadingIndicatorTemplate = "loadingindicator" + settings.TemplateSuffix
	ConnectionLostTemplate   = "connectionlost" + settings.TemplateSuffix
	NoScriptTemplate         = "noscript" + settings.TemplateSuffix
)

var (
	// Template Stores
	TemplatesStore     *store.Store
	CoreTemplatesStore *store.Store
)

//##############//
//### Public ###//
//##############//

func Init() error {
	// Create the store paths.
	paths := []string{
		settings.Settings.PagesPath,
		settings.Settings.TemplatesPath,
	}

	// Create a new store and parse it.
	s, err := store.New(paths)
	if err != nil {
		return err
	}
	s.Parse()

	// Set the templates store.
	TemplatesStore = s

	// Create the store paths.
	paths = []string{
		settings.Settings.CoreTemplatesPath,
	}

	// Create a new store and parse it.
	s, err = store.New(paths)
	if err != nil {
		return err
	}
	s.Parse()

	// Set the core templates store.
	CoreTemplatesStore = s

	return nil
}

func ExecNotFoundTemplate(s *sessions.Session) (int, string) {
	// Execute the not found page
	out, _, err := CoreTemplatesStore.Templates.ExecuteTemplateToString(s, NotFoundTemplate, nil)
	if err != nil {
		return ExecErrorTemplate(s, err.Error())
	}

	return 404, out
}

// ExecNotFoundTemplate executes the error template and shows the error message if the
// user is authenticated. The error message will be also logged.
func ExecErrorTemplate(s *sessions.Session, errorMessage string) (int, string) {
	// Create the template data struct.
	data := struct {
		ErrorMessage string
	}{
		errorMessage,
	}

	// Log the error.
	log.L.Error(errorMessage)

	// Execute the template.
	out, _, err := CoreTemplatesStore.Templates.ExecuteTemplateToString(s, ErrorTemplate, data)
	if err != nil {
		log.L.Error("failed to execute error core template: %v", err)
		return 500, "Internal Server Error"
	}

	return 500, out
}
