/*
 *  Bulldozer Framework
 *  Copyright (C) DesertBit
 */

package editmode

import (
	tr "code.desertbit.com/bulldozer/bulldozer/translate"

	"code.desertbit.com/bulldozer/bulldozer/sessions"
)

const (
	// Session value keys.
	sessionValueKeyIsActive = "blzEditModeActive"
)

var (
	backend bulldozerBackend
)

//###################################//
//### Bulldozer backend interface ###//
//###################################//

type bulldozerBackend interface {
	ReloadPage(s *sessions.Session)
}

//##############//
//### Public ###//
//##############//

func Init(b bulldozerBackend) error {
	// Set the backend.
	backend = b

	return nil
}

// Start the edit mode.
func Start(s *sessions.Session) {
	// Set the edit mode session data value.
	s.InstanceSet(sessionValueKeyIsActive, true)

	// Confirm on exit.
	s.SetExitMessage(tr.S("blz.core.exitMessage"))

	// Reload the current page.
	backend.ReloadPage(s)
}

// Stop the edit mode.
func Stop(s *sessions.Session) {
	// Remove the edit mode session data value.
	s.InstanceDelete(sessionValueKeyIsActive)

	// Remove the confirm message on exit.
	s.ResetExitMessage()

	// Reload the current page.
	backend.ReloadPage(s)
}

// IsActive returns a boolean whenever the edit mode is active.
func IsActive(s *sessions.Session) bool {
	// Get the session data value.
	i, ok := s.InstanceGet(sessionValueKeyIsActive)
	if !ok {
		return false
	}

	// Assertion.
	active, ok := i.(bool)
	if !ok {
		return false
	}

	return active
}
