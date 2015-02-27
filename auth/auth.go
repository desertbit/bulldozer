/*
 *  Bulldozer Framework
 *  Copyright (C) DesertBit
 */

package auth

import (
	"code.desertbit.com/bulldozer/bulldozer/log"
	"code.desertbit.com/bulldozer/bulldozer/mux"
	"code.desertbit.com/bulldozer/bulldozer/sessions"
	"code.desertbit.com/bulldozer/bulldozer/template"
	"code.desertbit.com/bulldozer/bulldozer/templates"

	"encoding/gob"
	"fmt"
)

const (
	// Page Urls
	LoginPageUrl    = "/login"
	RegisterPageUrl = "/register"

	// Template names:
	loginTemplate                = "bud/auth/login"
	registerTemplate             = "bud/auth/register"
	changePasswordDialogTemplate = "bud/auth/changepassworddialog"

	// Session value keys.
	sessionValueKeyIsAuth = "budAuthData"

	// Context Execution keys.
	contextValueKeyIsAuth = "budAuthData"
)

func init() {
	// Register the custom type.
	gob.Register(new(sessionAuthData))
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

func Init() (err error) {
	// Obtain the login template and prepare it.
	t := templates.Templates.Lookup(loginTemplate)
	if t == nil {
		return fmt.Errorf("failed to lookup auth login template!")
	}
	t.AddStyleClass("bud-page").
		RegisterEvents(new(loginEvents)).
		OnGetData(onLoginTemplateGetData)

	// Obtain the login template and prepare it.
	t = templates.Templates.Lookup(registerTemplate)
	if t == nil {
		return fmt.Errorf("failed to lookup auth register template!")
	}
	t.AddStyleClass("bud-page").
		RegisterEvents(new(registerEvents))

	// Obtain the change password dialog template  and prepare it.
	t = templates.Templates.Lookup(changePasswordDialogTemplate)
	if t == nil {
		return fmt.Errorf("failed to lookup auth change password dialog template!")
	}
	t.RegisterEvents(new(changePasswordDialogEvents))
	changePasswordDialog.SetTemplate(t)

	// Set the login route.
	mux.Route(LoginPageUrl, routeLoginPage)
	mux.Route(RegisterPageUrl, routeRegisterPage)

	// Initialize the database.
	initDB()

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
	s.NavigateHome()

	// Trigger the event
	triggerOnEndAuthenticatedSession(s)
}

func NavigateToLoginPage(s *sessions.Session) {
	s.Navigate(LoginPageUrl)
}

func NavigateToRegisterPage(s *sessions.Session) {
	s.Navigate(RegisterPageUrl)
}
