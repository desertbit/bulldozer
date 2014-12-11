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

//##############//
//### Public ###//
//##############//

// New should create and return a new session
func New() (s *Session, id string, err error) {
	// Lock the mutex
	mutex.Lock()
	defer mutex.Unlock()

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

	// Create a new session
	s = &Session{
		id:     id,
		values: make(map[interface{}]interface{}),
	}

	// Add the session to the map
	sessions[id] = s

	// Don't save anything to the backend.
	// This is done automatically as soon as any data is set to the session.

	return
}

// Get should return a session fitting to the session ID
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

	return s, nil
}
