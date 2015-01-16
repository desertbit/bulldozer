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
	"encoding/gob"
)

const (
	// Page Urls
	LoginPageUrl    = "/login"
	RegisterPageUrl = "/register"

	// Authentication template directory name.
	authTemplatesDir = "auth"

	// Template names:
	loginTemplate    = "login.tmpl"
	registerTemplate = "register.tmpl"

	// Session value keys.
	sessionValueKeyIsAuth = "blzAuthData"
)

var (
	backend bulldozerBackend

	// Template Stores
	templatesStore *store.Store
)

func init() {
	// Register the custom type.
	gob.Register(new(sessionAuthData))
}

//###################################//
//### Bulldozer backend interface ###//
//###################################//

type bulldozerBackend interface {
	NavigateToPath(s *sessions.Session, path string)
	Route(path string, f func(*sessions.Session, *router.Data) (string, string, error))
}

//###################################//
//### Session authentication data ###//
//###################################//

type sessionAuthData struct {
	UserID string
}

//##############//
//### Public ###//
//##############//

func Init(b bulldozerBackend) error {
	// Set the backend.
	backend = b

	// Create a new store and parse it.
	s, err := store.New(settings.Settings.BulldozerCoreTemplatesPath + "/" + authTemplatesDir)
	if err != nil {
		return err
	}

	// Customize the templates after each parse.
	s.OnAfterParse(func(s *store.Store) {
		if t := lookupTemplate(s.Templates, loginTemplate); t != nil {
			t.AddStyleClass("bulldozer-page").RegisterEvents(new(loginEvents))
		}

		if t := lookupTemplate(s.Templates, registerTemplate); t != nil {
			t.AddStyleClass("bulldozer-page").RegisterEvents(new(registerEvents))
		}
	})

	// Parse the templates.
	s.Parse()

	// Set the templates store.
	templatesStore = s

	// Set the login route.
	backend.Route(LoginPageUrl, routeLoginPage)
	backend.Route(RegisterPageUrl, routeRegisterPage)

	// Initialize the database.
	if err = initDB(); err != nil {
		return err
	}

	// Start the expire cache loop.
	startExpireCacheLoop()

	return nil
}

// Release this package and stop all goroutines.
func Release() {
	releaseDB()
	releaseCache()
}

// IsAuth returns a boolean if the current session is authenticated
// by a user login.
func IsAuth(s *sessions.Session) bool {
	return GetUser(s) != nil
}

// GetUser returns the logged in user value if logged in.
// Otherwise nil is returned.
// This user value is not updated, if any user data changes!
// Call user.Update() to get the latest state.
func GetUser(s *sessions.Session) *User {
	// Get the session data value.
	i, ok := s.Get(sessionValueKeyIsAuth)
	if !ok {
		return nil
	}

	// Assertion.
	d, ok := i.(*sessionAuthData)
	if !ok {
		return nil
	}

	// Obtain the user value from the cache or database with the user ID.
	u, err := cacheGetDBUser(d.UserID)
	if err != nil {
		log.L.Error(err.Error())
		return nil
	} else if u == nil {
		return nil
	}

	return newUser(u)
}

// Logout logs out the user if authenticated.
func Logout(s *sessions.Session) {
	// Remove the authenticated user data if present.
	s.Delete(sessionValueKeyIsAuth)

	// Redirect to the default page.
	backend.NavigateToPath(s, "/")

	// Trigger the event
	triggerOnEndAuthenticatedSession(s)
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
