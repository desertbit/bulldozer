/*
 *  Bulldozer Framework
 *  Copyright (C) DesertBit
 */

package topbar

import (
	"code.desertbit.com/bulldozer/bulldozer/auth"
	"code.desertbit.com/bulldozer/bulldozer/editmode"
	"code.desertbit.com/bulldozer/bulldozer/sessions"
	"code.desertbit.com/bulldozer/bulldozer/settings"
	"code.desertbit.com/bulldozer/bulldozer/template"
	"fmt"
)

const (
	topbarTemplatesUID = "blzTopBar"
	topbarTemplatesDir = "backend/topbar/"

	// Template names:
	topbarTemplate = "topbar" + settings.TemplateSuffix
)

var (
	templates *template.Template
)

func init() {
	auth.OnEndAuthenticatedSession(onEndAuthenticatedSession)
}

//##############//
//### Public ###//
//##############//

func Init() (err error) {
	// Create the file path.
	file := settings.GetCoreTemplatePath(topbarTemplatesDir + topbarTemplate)

	// Create and parse the templates.
	templates, err = template.ParseFiles(topbarTemplatesUID, file)
	if err != nil {
		return err
	}

	// Register the template events.
	templates.RegisterEvents(new(events))

	return nil
}

// ExecTopBar executes the topbar.
// You can pass a session or context value to this method.
// If a context value is available, then always pass it instead of the session.
// This will improve the performance and won't retrieve a user value multiple
// times from the database during one template execution cycle.
// The context is only used to retrieve the authenticated user information if present.
// The topbar is executed in it's own context anyway.
func ExecTopBar(i interface{}) (string, error) {
	var s *sessions.Session

	switch i.(type) {
	case *sessions.Session:
		// Set the session pointer.
		s = i.(*sessions.Session)
	case *template.Context:
		// Assert to context value.
		c := i.(*template.Context)

		// Set the session pointer.
		s = c.Session()
	default:
		return "", fmt.Errorf("invalid topbar.ExecTopBar call: called method with invalid interface type!")
	}

	// Get the current user if logged in.
	user := auth.GetUser(i)
	if user == nil {
		return "", nil
	}

	// TODO: Only show topbar if the user has the specific group set.

	// The render data.
	data := struct {
		User *auth.User
	}{
		user,
	}

	// Execute the topbar template.
	body, _, _, err := templates.ExecuteTemplateToString(s, topbarTemplate, data)
	if err != nil {
		return "", err
	}

	return body, nil
}

//###############//
//### Private ###//
//###############//

func onEndAuthenticatedSession(s *sessions.Session) {
	s.SendCommand("Bulldozer.topbar.space(false);")
}

//##############//
//### Events ###//
//##############//

type events struct{}

func (e *events) EventLogout(c *template.Context) {
	// Logout.
	auth.Logout(c.Session())
}

func (e *events) EventStartEdit(c *template.Context) {
	// Get the session pointer.
	s := c.Session()

	// Hide the loading indicator on return.
	defer s.HideLoadingIndicator()

	// Start the edit mode.
	editmode.Start(s)
}

func (e *events) EventStopEdit(c *template.Context) {
	// Get the session pointer.
	s := c.Session()

	// Hide the loading indicator on return.
	defer s.HideLoadingIndicator()

	// Stop the edit mode.
	editmode.Stop(s)
}
