/*
 *  Bulldozer Framework
 *  Copyright (C) DesertBit
 */

package topbar

import (
	"code.desertbit.com/bulldozer/bulldozer/auth"
	"code.desertbit.com/bulldozer/bulldozer/controlpanel"
	"code.desertbit.com/bulldozer/bulldozer/editmode"
	"code.desertbit.com/bulldozer/bulldozer/sessions"
	"code.desertbit.com/bulldozer/bulldozer/store"
	"code.desertbit.com/bulldozer/bulldozer/template"
	"code.desertbit.com/bulldozer/bulldozer/templates"
	"fmt"
)

const (
	// Template name
	topbarTemplate = "bud/topbar/topbar"
)

func init() {
	auth.OnEndAuthenticatedSession(onEndAuthenticatedSession)
}

//##############//
//### Public ###//
//##############//

func Init() (err error) {
	// Obtain the control center template.
	t := templates.Templates.Lookup(topbarTemplate)
	if t == nil {
		return fmt.Errorf("failed to lookup topbar template!")
	}

	// Register the topbar events.
	t.RegisterEvents(new(events))

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

	// Template options and data.
	opts := template.ExecOpts{
		Data: struct {
			IsControlPanelPage bool
			ControlPanelUrl    string
		}{
			IsControlPanelPage: controlpanel.IsCurrentPage(s),
			ControlPanelUrl:    controlpanel.PageUrl,
		},
	}

	// Execute the topbar template.
	body, _, _, err := templates.Templates.ExecuteTemplateToString(s, topbarTemplate, opts)
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
	// Hide the loading indicator on return.
	s.HideLoadingIndicator()

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

func (e *events) EventSaveChanges(c *template.Context) {
	// Get the session pointer.
	s := c.Session()

	// Hide the loading indicator on return.
	defer s.HideLoadingIndicator()

	// Save all temporary changes.
	store.SaveTemporaryChanges(s)
}

func (e *events) EventDiscardChanges(c *template.Context) {
	// Get the session pointer.
	s := c.Session()

	// Hide the loading indicator on return.
	defer s.HideLoadingIndicator()

	// Remove all temporary changes.
	//store.DiscardTemporaryChanges(s)
}
