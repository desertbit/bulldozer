/*
 *  Bulldozer Framework
 *  Copyright (C) DesertBit
 */

package topbar

import (
	"code.desertbit.com/bulldozer/bulldozer/auth"
	"code.desertbit.com/bulldozer/bulldozer/sessions"
	"code.desertbit.com/bulldozer/bulldozer/settings"
	"code.desertbit.com/bulldozer/bulldozer/template"
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

func ExecTopBar(s *sessions.Session) (string, error) {
	// Get the current user if logged in.
	user := auth.GetUser(s)
	if user == nil {
		return "", nil
	}

	// TODO: Only show topbar if the user has the specific group set.

	// The render daza.
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
