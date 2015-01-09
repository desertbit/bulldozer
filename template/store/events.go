/*
 *  Bulldozer Framework
 *  Copyright (C) DesertBit
 */

package store

import (
	"code.desertbit.com/bulldozer/bulldozer/log"
)

const (
	onBeforeParse = "beforeParse"
	onAfterParse  = "afterParse"
)

//#############################//
//### Template Struct Event ###//
//#############################//

// OnBeforeParse is triggered before the templates get parsed.
func (s *Store) OnBeforeParse(f func(s *Store)) {
	s.emitter.On(onBeforeParse, f)
}

// Triggere the event.
func (s *Store) triggerOnBeforeParse() {
	s.emitter.Emit(onBeforeParse, s)
}

// OnAfterParse is triggered after the templates got parsed.
func (s *Store) OnAfterParse(f func(s *Store)) {
	s.emitter.On(onAfterParse, f)
}

// Triggere the event.
func (s *Store) triggerOnAfterParse() {
	s.emitter.Emit(onAfterParse, s)
}

//###############//
//### Private ###//
//###############//

func recoverEmitter(event interface{}, listener interface{}, err error) {
	log.L.Error("bulldozer templates store events error: emitter event: %v: listener: %v: %v", event, listener, err)
}
