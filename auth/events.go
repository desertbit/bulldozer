/*
 *  Bulldozer Framework
 *  Copyright (C) DesertBit
 */

package auth

import (
	"github.com/chuckpreslar/emission"
	"github.com/desertbit/bulldozer/log"
	"github.com/desertbit/bulldozer/sessions"
)

const (
	onNewAuthenticatedSession = "OnNewAuthSession"
	onEndAuthenticatedSession = "OnEndAuthSession"
)

var (
	emitter *emission.Emitter
)

func init() {
	// Create a new emitter and set the recover function
	emitter = emission.NewEmitter().
		RecoverWith(recoverEmitter)
}

//##############//
//### Public ###//
//##############//

// OnNewAuthenticatedSession sets the function which is triggered during each new authenticated session
func OnNewAuthenticatedSession(f func(s *sessions.Session)) {
	emitter.On(onNewAuthenticatedSession, f)
}

// OnEndAuthenticatedSession sets the function which is triggered during each logout
func OnEndAuthenticatedSession(f func(s *sessions.Session)) {
	emitter.On(onEndAuthenticatedSession, f)
}

// OffNewAuthenticatedSession removes the listener again
func OffNewAuthenticatedSession(f func(s *sessions.Session)) {
	emitter.Off(onNewAuthenticatedSession, f)
}

// OffEndAuthenticatedSession removes the listener again
func OffEndAuthenticatedSession(f func(s *sessions.Session)) {
	emitter.Off(onEndAuthenticatedSession, f)
}

//###############//
//### Private ###//
//###############//

func recoverEmitter(event interface{}, listener interface{}, err error) {
	log.L.Error("authentication events error: emitter event: %v: listener: %v: %v", event, listener, err)
}

func triggerOnNewAuthenticatedSession(s *sessions.Session) {
	emitter.Emit(onNewAuthenticatedSession, s)
}

func triggerOnEndAuthenticatedSession(s *sessions.Session) {
	emitter.Emit(onEndAuthenticatedSession, s)
}
