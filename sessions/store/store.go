/*
 *  Bulldozer Framework
 *  Copyright (C) DesertBit
 */

package store

import (
	"code.desertbit.com/bulldozer/bulldozer/utils"
	"sync"
)

const (
	sessionIDLength = 40
)

var (
	mutex    sync.Mutex
	sessions map[string]*Session = make(map[string]*Session)
)

/* Hint: For debugging purpose
func init() {
	go func() {
		for {
			time.Sleep(5 * time.Second)
			mutex.Lock()
			fmt.Printf("Cache state:  %v\n", sessions)
			mutex.Unlock()
		}
	}()
}*/

//##############//
//### Public ###//
//##############//

// New will create and return a new session.
// This operation is thread-safe.
func New() (*Session, error) {
	// Lock the mutex
	mutex.Lock()
	defer mutex.Unlock()

	id, err := newUniqueSessionID()
	if err != nil {
		return nil, err
	}

	// Create a new session
	s := &Session{
		id:          id,
		valid:       true,
		dirty:       false,
		values:      make(map[interface{}]interface{}),
		cacheValues: make(map[interface{}]interface{}),
	}

	// Add the session to the map
	sessions[id] = s

	// Remove the session from the cache if not locked after the timeout
	removeSessionFromCacheAfterTimeout(s)

	// Don't save anything to the backend.
	// This is done automatically as soon as any data is set to the session.

	return s, nil
}

// Get will return a session fitting to the session ID.
// This operation is thread-safe.
func Get(id string) (*Session, error) {
	// Lock the mutex
	mutex.Lock()
	defer mutex.Unlock()

	// Try to get the session from the cached sessions map
	s, ok := sessions[id]
	if ok {
		return s, nil
	}

	// Try to obtain the session from the database
	var err error
	s, err = getSessionFromDB(id)
	if err != nil {
		return nil, err
	}

	// Add the session to the map
	sessions[id] = s

	// Remove the session from the cache if not locked after the timeout
	removeSessionFromCacheAfterTimeout(s)

	return s, nil
}

// Remove removes the session completly from the cache and database
// This operation is thread-safe.
func Remove(id string) {
	// Lock the mutex
	mutex.Lock()
	defer mutex.Unlock()

	// Try to get the session from the cached sessions map
	s, ok := sessions[id]
	if ok {
		// Set the valid flag to false, to be sure, this
		// session won't be saved to the database
		s.valid = false
	}

	// Delete the session from the map
	delete(sessions, id)

	// Remove the session also from the database
	removeSessionFromDB(id)
}

// AssignNewSessionID invalidates the old ID and creates a new ID for the session.
// This operation is thread-safe.
func AssignNewSessionID(s *Session) error {
	// Lock the session mutex
	s.mutex.Lock()
	defer s.mutex.Unlock()

	// Lock the mutex
	mutex.Lock()
	defer mutex.Unlock()

	// Create a new session ID
	id, err := newUniqueSessionID()
	if err != nil {
		return err
	}

	// Delete the session from the map
	delete(sessions, s.id)

	// Remove the session also from the database
	removeSessionFromDB(s.id)

	// Set the new session ID
	s.id = id

	// Add the session with the new ID to the map
	sessions[id] = s

	// Register the changed session if the values map is not emtpy
	if len(s.values) > 0 {
		registerChangedSession(s)
	}

	return nil
}

//###############//
//### Private ###//
//###############//

// newUniqueSessionID creates a new unique session ID.
// Be sure to lock the mutex before calling this method!
func newUniqueSessionID() (id string, err error) {
	// Obtain a new unique session Id
	for {
		// Get a new session ID
		id = utils.RandomString(sessionIDLength)

		// Check if the session Id is already used.
		// This is very unlikely, but we have to check this!
		_, ok := sessions[id]
		if ok {
			// Continue the loop if the ID exists already
			continue
		}

		// Check if the ID exists in the database
		ok, err = sessionIDExistsInDB(id)
		if err != nil {
			return
		}
		if ok {
			// Continue the loop if the ID exists already
			continue
		}

		// Break the loop if the ID is unique
		break
	}

	return
}
