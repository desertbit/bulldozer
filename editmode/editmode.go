/*
 *  Bulldozer Framework
 *  Copyright (C) DesertBit
 */

package editmode

import (
	tr "github.com/desertbit/bulldozer/translate"

	"github.com/desertbit/bulldozer/sessions"
	"github.com/desertbit/bulldozer/template"
	"sync"
)

const (
	// Session value keys.
	sessionValueKeyIsActive = "budEditModeActive"
)

var (
	// Sessions active in editmode.
	activeSessions      map[string]*sessions.Session = make(map[string]*sessions.Session)
	activeSessionsMutex sync.Mutex
)

func init() {
	// Attach the event listener.
	sessions.OnNewSession(onNewSessionFunc)
}

//##############//
//### Public ###//
//##############//

// Start the edit mode.
func Start(s *sessions.Session) {
	// Set the edit mode session data value.
	s.InstanceSet(sessionValueKeyIsActive, true)

	// Enable the session context store.
	template.EnableSessionContextStore(s)

	// Confirm on exit.
	s.SetExitMessage(tr.S("bud.core.exitMessage"))

	// Add the session to the active sessions.
	addSession(s, true)

	// Reload the current page.
	s.Reload()
}

// Stop the edit mode.
func Stop(s *sessions.Session) {
	// Remove the edit mode session data value.
	s.InstanceDelete(sessionValueKeyIsActive)

	// Disable the session context store again.
	template.DisableSessionContextStore(s)

	// Remove the confirm message on exit.
	s.ResetExitMessage()

	// Remove the session from the active sessions.
	removeSession(s)

	// Reload the current page.
	s.Reload()
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

// HasActiveSessions returns a boolean if there are
// active sessions in the edit mode.
func HasActiveSessions() bool {
	return len(activeSessions) > 0
}

// GetSessions returns a slice of all active edit mode sessions.
func GetSessions() []*sessions.Session {
	// Lock the mutex.
	activeSessionsMutex.Lock()
	defer activeSessionsMutex.Unlock()

	// Create the slice.
	l := make([]*sessions.Session, len(activeSessions))

	i := 0
	for _, s := range activeSessions {
		l[i] = s
		i++
	}

	return l
}

//###############//
//### Private ###//
//###############//

func addSession(s *sessions.Session, triggerEvent bool) {
	// Remove the session if closed from the map.
	s.OnceClose(removeSession)

	// Lock the mutex.
	activeSessionsMutex.Lock()
	defer activeSessionsMutex.Unlock()

	// Add the session to the map.
	activeSessions[s.SessionID()] = s

	if triggerEvent {
		// Trigger the event.
		triggerOnNewSession(s)
	}
}

func removeSession(s *sessions.Session) {
	// Detach the event again.
	s.OffClose(removeSession)

	// Lock the mutex.
	activeSessionsMutex.Lock()
	defer activeSessionsMutex.Unlock()

	// Remove the session from the map.
	delete(activeSessions, s.SessionID())

	// Trigger the event.
	triggerOnRemoveSession(s)
}

func onNewSessionFunc(s *sessions.Session) {
	// Skip if no edit mode is active.
	if !IsActive(s) {
		return
	}

	// Readd the session to the active session map.
	addSession(s, false)

	// Trigger the event.
	triggerOnSessionReconnect(s)
}
