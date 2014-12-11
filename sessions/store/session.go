/*
 *  Bulldozer Framework
 *  Copyright (C) DesertBit
 */

package store

import (
	"sync"
)

//###############//
//### Session ###//
//###############//

type Session struct {
	id     string
	values map[interface{}]interface{}
	mutex  sync.Mutex
}

// ID returns the unique session ID. Don't expose this.
func (s *Session) ID() string {
	return s.id
}

// Get obtains the value. This operation is thread-safe.
func (s *Session) Get(key interface{}) (value interface{}, ok bool) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	value, ok = s.values[key]
	return
}

// Set sets the value with the given key. This operation is thread-safe.
func (s *Session) Set(key interface{}, value interface{}) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	s.values[key] = value

	// Register the changed session
	registerChangedSession(s)
}

// Delete removes the value with the given key. This operation is thread-safe.
func (s *Session) Delete(key interface{}) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	delete(s.values, key)

	// Register the changed session
	registerChangedSession(s)
}
