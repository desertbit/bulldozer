/*
 *  Bulldozer Framework
 *  Copyright (C) DesertBit
 */

package auth

import (
	"code.desertbit.com/bulldozer/bulldozer/log"
	"code.desertbit.com/bulldozer/bulldozer/router"
	"code.desertbit.com/bulldozer/bulldozer/sessions"
	"code.desertbit.com/bulldozer/bulldozer/settings"
	"code.desertbit.com/bulldozer/bulldozer/template"
	"code.desertbit.com/bulldozer/bulldozer/template/store"
)

const (
	// Page Urls
	LoginPageUrl = "/login"

	// Authentication template directory name.
	authTemplatesDir = "auth"

	// Template names:
	loginTemplate = "login.tmpl"
)

var (
	backend bulldozerBackend

	// Template Stores
	templatesStore *store.Store
)

//###################################//
//### Bulldozer backend interface ###//
//###################################//

type bulldozerBackend interface {
	Route(path string, f func(*sessions.Session, *router.Data) (string, string, error))
}

//##############//
//### Public ###//
//##############//

func Init(backend bulldozerBackend) error {
	// Set the backend.
	backend = backend

	// Create a new store and parse it.
	s, err := store.New(settings.Settings.BulldozerCoreTemplatesPath + "/" + authTemplatesDir)
	if err != nil {
		return err
	}
	s.Parse()

	// Set the templates store.
	templatesStore = s

	// Register the template events.
	lookupMust(s.Templates, loginTemplate).RegisterEvents(new(loginEvents))

	// Set the login route.
	backend.Route(LoginPageUrl, routeLoginPage)

	return nil
}

//###############//
//### Private ###//
//###############//

func lookupMust(t *template.Template, name string) *template.Template {
	if t == nil {
		log.L.Fatalf("failed to find template '%s'", name)
	}

	t = t.Lookup(name)
	if t == nil {
		log.L.Fatalf("failed to find template '%s'", name)
	}

	return t
}
