/*
 *  Bulldozer Framework
 *  Copyright (C) DesertBit
 */

package global

import (
	tr "code.desertbit.com/bulldozer/bulldozer/translate"

	"code.desertbit.com/bulldozer/bulldozer/log"
	"code.desertbit.com/bulldozer/bulldozer/sessions"
	"code.desertbit.com/bulldozer/bulldozer/settings"
	"code.desertbit.com/bulldozer/bulldozer/template"
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
	//###
	//### Load the Templates.
	//###

	// Create a new store and parse it.
	s, err := store.New(settings.Settings.TemplatesPath, settings.Settings.PagesPath)
	if err != nil {
		return err
	}
	s.Parse()

	// Set the templates store.
	TemplatesStore = s

	//###
	//### Load the Core Templates.
	//###

	// Create a new store and parse it.
	s, err = store.New(settings.Settings.CoreTemplatesPath)
	if err != nil {
		return err
	}

	// Customize the templates after each parse.
	s.OnAfterParse(func(s *store.Store) {
		if t := lookupTemplate(s.Templates, LoadingIndicatorTemplate); t != nil {
			t.SetStaticDomID("bulldozer-loading-indicator")
		}

		if t := lookupTemplate(s.Templates, ConnectionLostTemplate); t != nil {
			t.SetStaticDomID("bulldozer-connection-lost")
		}

		if t := lookupTemplate(s.Templates, NoScriptTemplate); t != nil {
			t.SetStaticDomID("bulldozer-noscript")
		}

		if t := lookupTemplate(s.Templates, NotFoundTemplate); t != nil {
			t.AddStyleClass("bulldozer-page").AddStyleClass("bulldozer-not-found-page")
		}

		if t := lookupTemplate(s.Templates, ErrorTemplate); t != nil {
			t.AddStyleClass("bulldozer-page").AddStyleClass("bulldozer-error-page")
		}
	})

	// Parse the templates.
	s.Parse()

	// Set the core templates store.
	CoreTemplatesStore = s

	return nil
}

func ExecNotFoundTemplate(s *sessions.Session) (int, string, string) {
	// Execute the not found page
	out, _, _, err := CoreTemplatesStore.Templates.ExecuteTemplateToString(s, NotFoundTemplate, nil)
	if err != nil {
		return ExecErrorTemplate(s, err.Error())
	}

	return 404, out, tr.S("blz.page.notFound.pageTitle")
}

// ExecNotFoundTemplate executes the error template and shows the error message if the
// user is authenticated. The error message will be also logged.
// One optional boolean can be set. If set to false, the error message won't be logged.
func ExecErrorTemplate(s *sessions.Session, errorMessage string, vars ...bool) (int, string, string) {
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

	// Execute the template.
	out, _, _, err := CoreTemplatesStore.Templates.ExecuteTemplateToString(s, ErrorTemplate, data)
	if err != nil {
		log.L.Error("failed to execute error core template: %v", err)
		return 500, "Internal Server Error", tr.S("blz.page.error.pageTitle")
	}

	return 500, out, tr.S("blz.page.error.pageTitle")
}

//###############//
//### Private ###//
//###############//

func lookupTemplate(t *template.Template, name string) *template.Template {
	if t == nil {
		log.L.Error("failed to find template '%s'", name)
	}

	t = t.Lookup(name)
	if t == nil {
		log.L.Error("failed to find template '%s'", name)
	}

	return t
}
