/*
 *  Bulldozer Framework
 *  Copyright (C) DesertBit
 */

package bulldozer

import (
	"github.com/desertbit/bulldozer/mux"
	"github.com/desertbit/bulldozer/sessions"
	"github.com/desertbit/bulldozer/templates"
	"github.com/desertbit/bulldozer/topbar"
	"github.com/desertbit/bulldozer/utils"
)

var (
	backendI *backendInterface = new(backendInterface)
)

//##################################################//
//### Private bulldozer backend for sub packages ###//
//##################################################//

type backendInterface struct{}

// This navigates the session to the given route path.
func (i *backendInterface) NavigateFunc(s *sessions.Session, path string) {
	// Execute the route.
	_, body, title, path := mux.ExecRoute(s, path)

	// Render the page.
	renderPage(s, title, body, path)
}

func (i *backendInterface) ShowErrorPage(s *sessions.Session, errorMessage string, vars ...bool) {
	// Execute the template.
	_, body, title := templates.ExecError(s, errorMessage, vars...)

	// Render the page.
	renderPage(s, title, body, s.CurrentPath())
}

func (i *backendInterface) ShowNotFoundPage(s *sessions.Session) {
	// Execute the template.
	_, body, title := templates.ExecNotFound(s)

	// Render the page.
	renderPage(s, title, body, s.CurrentPath())
}

func (i *backendInterface) ExecTopBar(ti interface{}) (string, error) {
	return topbar.ExecTopBar(ti)
}

//###############//
//### Private ###//
//###############//

func renderPage(s *sessions.Session, title string, body string, path string) {
	// Create the client command.
	cmd := `Bulldozer.render.page('` +
		utils.EscapeJS(body) + `','` +
		utils.EscapeJS(title) + `','` +
		utils.EscapeJS(path) + `');`

	// Send the new render request to the client.
	s.SendCommand(cmd)
}
