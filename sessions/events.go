/*
 *  Bulldozer Framework
 *  Copyright (C) DesertBit
 */

package sessions

import (
	"code.desertbit.com/bulldozer/bulldozer/log"
	"github.com/chuckpreslar/emission"
)

const (
	onNewSession   = "OnNewSession"
	onCloseSession = "OnCloseSession"
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

// OnNewSession is triggered during each new session creation
func OnNewSession(f func(s *Session)) {
	emitter.On(onNewSession, f)
}

// OnCloseSession is triggered just before a session is destroyed
func OnCloseSession(f func(s *Session)) {
	emitter.On(onCloseSession, f)
}

// OffNewSession removes the listener again
func OffNewSession(f func(s *Session)) {
	emitter.Off(onNewSession, f)
}

// OffCloseSession removes the listener again
func OffCloseSession(f func(s *Session)) {
	emitter.Off(onCloseSession, f)
}

//###############//
//### Private ###//
//###############//

func recoverEmitter(event interface{}, listener interface{}, err error) {
	log.L.Error("session events error: emitter event: %v: listener: %v: %v", event, listener, err)
}

func triggerOnNewSession(s *Session) {
	emitter.Emit(onNewSession, s)
}

func triggerOnCloseSession(s *Session) {
	emitter.Emit(onCloseSession, s)
}

//############################//
//### Session Struct Event ###//
//############################//

// OnClose is triggered just before a session is destroyed
func (s *Session) OnClose(f func(s *Session)) {
	s.emitter.On(onCloseSession, f)
}

// OnceClose is the same event as OnClose, but the listener is triggered only once
func (s *Session) OnceClose(f func(s *Session)) {
	s.emitter.Once(onCloseSession, f)
}

// OffClose removes the listener again
func (s *Session) OffClose(f func(s *Session)) {
	s.emitter.Off(onCloseSession, f)
}

// triggerOnClose triggeres the event
func (s *Session) triggerOnClose() {
	s.emitter.Emit(onCloseSession, s)
}
