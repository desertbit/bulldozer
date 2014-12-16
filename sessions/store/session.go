/*
 *  Bulldozer Framework
 *  Copyright (C) DesertBit
 */

package store

import (
	"sync"
	"time"
)

const (
	removeSessionFromCacheTimeout = 5 * time.Second
)

//###############//
//### Session ###//
//###############//

type Session struct {
	id        string
	valid     bool
	lockCount int

	// Main value implementation. Values are stored to the database.
	values map[interface{}]interface{}
	mutex  sync.Mutex

	// Cache values, are values only saved as long as this session is in the memory cache.
	cacheValues      map[interface{}]interface{}
	cacheValuesMutex sync.Mutex
}

// ID returns the unique session ID. Don't expose this information.
func (s *Session) ID() string {
	return s.id
}

// Dirty sets the session values to an unsaved state,
// which will trigger the save trigger handler.
// Use this method, if you don't want to always call the
// Set() method for pointer values.
func (s *Session) Dirty() {
	// Register the changed session
	registerChangedSession(s)
}

// Lock increments the lock count. If you call lock, you have to also call unlock.
// This method returns true, if the lock was successful. Otherwise, if this session
// is already released from cache, this will return false. Then, you should obtain
// a new session with store.Get.
// This operation is thread-safe.
func (s *Session) Lock() bool {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	// Check if this session is already released from cache.
	if s.lockCount < 0 {
		return false
	}

	s.lockCount++

	return true
}

// Unlock decrements the lock count. If this session is not more locked by any lock,
// then this session will be released from cache.
// This operation is thread-safe.
func (s *Session) Unlock() {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	s.lockCount--

	if s.lockCount <= 0 {
		// Be sure the lock count is 0
		s.lockCount = 0

		// Remove the session from the cache if not locked after the timeout
		removeSessionFromCacheAfterTimeout(s)
	}
}

//###################//
//### Main Values ###//
//###################//

// Get obtains the value.
// A single variadic argument is accepted, and it is optional:
// if a function is set, this function will be called if no value
// exists for the given key.
// This operation is thread-safe.
func (s *Session) Get(key interface{}, vars ...func() interface{}) (value interface{}, ok bool) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	value, ok = s.values[key]

	// If no value is found and the create function variable
	// is set, then call the function and set the new value.
	if !ok && len(vars) > 0 {
		value = vars[0]()
		s.values[key] = value
		ok = true

		// Register the changed session
		registerChangedSession(s)
	}

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

//####################//
//### Cache Values ###//
//####################//

// CacheGet obtains the cache value.
// A single variadic argument is accepted, and it is optional:
// if a function is set, this function will be called if no value
// exists for the given key.
// This operation is thread-safe.
func (s *Session) CacheGet(key interface{}, vars ...func() interface{}) (value interface{}, ok bool) {
	s.cacheValuesMutex.Lock()
	defer s.cacheValuesMutex.Unlock()

	value, ok = s.cacheValues[key]

	// If no value is found and the create function variable
	// is set, then call the function and set the new value.
	if !ok && len(vars) > 0 {
		value = vars[0]()
		s.cacheValues[key] = value
		ok = true
	}

	return
}

// CacheSet sets the cache value with the given key. This operation is thread-safe.
func (s *Session) CacheSet(key interface{}, value interface{}) {
	s.cacheValuesMutex.Lock()
	defer s.cacheValuesMutex.Unlock()

	s.cacheValues[key] = value
}

// CacheDelete removes the cache value with the given key. This operation is thread-safe.
func (s *Session) CacheDelete(key interface{}) {
	s.cacheValuesMutex.Lock()
	defer s.cacheValuesMutex.Unlock()

	delete(s.cacheValues, key)
}

//###############//
//### Private ###//
//###############//

// removeSessionFromCacheAfterTimeout removes the session from the cache if
// not locked within the timeout.
func removeSessionFromCacheAfterTimeout(s *Session) {
	go func() {
		// Sleep
		time.Sleep(removeSessionFromCacheTimeout)

		// Lock the session mutex to access the lock count variable
		s.mutex.Lock()
		defer s.mutex.Unlock()

		// If not locked by any instance, then remove this session from the cache again
		if s.lockCount == 0 {
			// Set the lockCount to -1, which indicates, that
			// this session is going to be released from cache.
			s.lockCount = -1

			// Lock the main mutex for the sessions map
			mutex.Lock()
			defer mutex.Unlock()

			// Delete the session from the map
			delete(sessions, s.id)
		}
	}()
}
