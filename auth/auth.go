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
	"encoding/gob"
)

const (
	// Page Urls
	LoginPageUrl    = "/login"
	RegisterPageUrl = "/register"

	authTemplatesUID = "blzAuth"
	authTemplatesDir = "auth/"

	// Template names:
	loginTemplate    = "login" + settings.TemplateSuffix
	registerTemplate = "register" + settings.TemplateSuffix

	// Session value keys.
	sessionValueKeyIsAuth = "blzAuthData"

	// Context Execution keys.
	contextValueKeyIsAuth = "blzAuthData"
)

var (
	backend bulldozerBackend

	// Templates
	templates *template.Template
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

func Init(b bulldozerBackend) (err error) {
	// Set the backend.
	backend = b

	// Create the files slice.
	files := []string{
		settings.GetCoreTemplatePath(authTemplatesDir + loginTemplate),
		settings.GetCoreTemplatePath(authTemplatesDir + registerTemplate),
	}

	// Create and parse the templates.
	templates, err = template.ParseFiles(authTemplatesUID, files...)
	if err != nil {
		return err
	}

	// Customize the templates.
	templates.LookupFatal(loginTemplate).
		AddStyleClass("bulldozer-page").
		RegisterEvents(new(loginEvents))

	templates.LookupFatal(registerTemplate).
		AddStyleClass("bulldozer-page").
		RegisterEvents(new(registerEvents))

	// Set the login route.
	backend.Route(LoginPageUrl, routeLoginPage)
	backend.Route(RegisterPageUrl, routeRegisterPage)

	// Initialize the database.
	if err = initDB(); err != nil {
		return err
	}

	return nil
}

// Release this package.
func Release() {
	releaseDB()
}

// IsAuth returns a boolean if the current session is authenticated
// by a user login.
// You can pass a session or context value to this method.
// If a context value is available, then always pass it instead of the session.
// This will improve the performance and won't retrieve a user value multiple
// times from the database during one template execution cycle.
func IsAuth(i interface{}) bool {
	return GetUser(i) != nil
}

// GetUser returns the logged in user value if logged in.
// Otherwise nil is returned.
// This user value is not updated, if any user data changes!
// Call user.Update() to get the latest state.
// You can pass a session or context value to this method.
// If a context value is available, then always pass it instead of the session.
// This will improve the performance and won't retrieve a user value multiple
// times from the database during one template execution cycle.
func GetUser(i interface{}) *User {
	var s *sessions.Session
	var c *template.Context

	switch i.(type) {
	case *sessions.Session:
		// Set the session pointer.
		s = i.(*sessions.Session)
	case *template.Context:
		// Assert to context value.
		c = i.(*template.Context)

		// Set the session pointer.
		s = c.Session()

		// If the user was already previously obtained and added
		// to the context execution values, then use this value
		// instead of getting it again from the database.
		uI, ok := c.Get(contextValueKeyIsAuth)
		if ok {
			user, ok := uI.(*User)
			if ok {
				return user
			}
		}
	default:
		log.L.Error("invalid auth.GetUser call: called method with invalid interface type!")
		return nil
	}

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

	// Obtain the user from the database.
	u, err := dbGetUserByID(d.UserID)
	if err != nil {
		log.L.Error(err.Error())
		return nil
	} else if u == nil {
		return nil
	}

	// Create a new user value.
	user := newUser(u)

	// If the context exists, then add the user to the execution values.
	if c != nil {
		c.Set(contextValueKeyIsAuth, user)
	}

	return user
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
