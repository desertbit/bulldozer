/*
 *  Goji Framework
 *  Copyright (C) Roland Singer
 */

package editmode

import (
	"github.com/chuckpreslar/emission"
	"github.com/desertbit/bulldozer/log"
	"github.com/desertbit/bulldozer/sessions"
)

const (
	onNewSession       = "OnNewSession"
	onRemoveSession    = "OnRemoveSession"
	onSessionReconnect = "OnSessionReconnect"
)

var (
	emitter *emission.Emitter
)

func init() {
	// Create a new emitter, set the recover function and the max listeners.
	emitter = emission.NewEmitter().
		RecoverWith(recoverEmitter).
		SetMaxListeners(20)
}

//##############//
//### Public ###//
//##############//

// OnNewSession is triggered during each new edit mode session.
func OnNewSession(f func(s *sessions.Session)) {
	emitter.On(onNewSession, f)
}

// OnRemoveSession is triggered if a session leaves the editmode.
func OnRemoveSession(f func(s *sessions.Session)) {
	emitter.On(onRemoveSession, f)
}

// OnSessionReconnect is triggered if an edit mode session reconnects.
func OnSessionReconnect(f func(s *sessions.Session)) {
	emitter.On(onSessionReconnect, f)
}

//###############//
//### Private ###//
//###############//

func recoverEmitter(event interface{}, listener interface{}, err error) {
	log.L.Error("editmode events error: emitter event: %v: listener: %v: %v", event, listener, err)
}

func triggerOnNewSession(s *sessions.Session) {
	emitter.Emit(onNewSession, s)
}

func triggerOnRemoveSession(s *sessions.Session) {
	emitter.Emit(onRemoveSession, s)
}

func triggerOnSessionReconnect(s *sessions.Session) {
	emitter.Emit(onSessionReconnect, s)
}
