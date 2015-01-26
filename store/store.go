/*
 *  Bulldozer Framework
 *  Copyright (C) DesertBit
 */

package store

import (
	"code.desertbit.com/bulldozer/bulldozer/editmode"
	"code.desertbit.com/bulldozer/bulldozer/log"
	"code.desertbit.com/bulldozer/bulldozer/sessions"
	"code.desertbit.com/bulldozer/bulldozer/template"
	"fmt"
	"sync"
)

const (
	// Context Execution keys.
	contextValueKeyStorePrefix = "blzStore_"
)

var (
	lockedContextIDs      map[string]string = make(map[string]string)
	lockedContextIDsMutex sync.Mutex
)

//##################//
//### Store type ###//
//##################//

type store struct {
	data           *dbStore
	mutex          sync.Mutex
	editModeActive bool
	existsInDB     bool
}

func newStore(data *dbStore) *store {
	return &store{
		data:           data,
		editModeActive: false,
		existsInDB:     false,
	}
}

//##############//
//### Public ###//
//##############//

func Init() error {
	// Initialize the database.
	return initDB()
}

// Lock the context for the current session.
// A boolean is returned, if the lock request was successful.
// This operation is thread-safe.
func Lock(c *template.Context) bool {
	id := c.ID()
	currentSid := c.Session().SessionID()

	// Lock the mutex.
	lockedContextIDsMutex.Lock()
	defer lockedContextIDsMutex.Unlock()

	// Check if already locked by another session.
	sid, ok := lockedContextIDs[id]
	if ok {
		// Check if already locked by the current session.
		if sid == currentSid {
			return true
		}

		// Check if the session which is holding the lock
		// is still active. Otherwise just take over the lock.
		_, ok = sessions.GetSession(sid)
		if ok {
			return false
		}
	}

	// Lock the context ID for the current session.
	lockedContextIDs[id] = currentSid

	return true
}

// Unlock the context again.
// This operation is thread-safe.
func Unlock(c *template.Context) {
	id := c.ID()
	currentSid := c.Session().SessionID()

	// Lock the mutex.
	lockedContextIDsMutex.Lock()
	defer lockedContextIDsMutex.Unlock()

	// Get the locked session ID.
	sid, ok := lockedContextIDs[id]
	if ok {
		// Check if locked by the current session.
		if sid != currentSid {
			log.L.Error("store: failed to unlock store cotext: the calling session is not the session which holds the lock!")
			return
		}

		// Unlock the lock.
		delete(lockedContextIDs, id)
	}
}

// IsLocked returns a boolean whenever the context is
// locked by the current session.
// This operation is thread-safe.
func IsLocked(c *template.Context) bool {
	// Lock the mutex.
	lockedContextIDsMutex.Lock()
	defer lockedContextIDsMutex.Unlock()

	// Check if locked by the current session.
	sid, ok := lockedContextIDs[c.ID()]
	if !ok {
		return false
	}

	// The locked session ID has to match.
	return sid == c.Session().SessionID()
}

// Get obtains the store value for the context.
// A single variadic argument is accepted, and it is optional:
// if a function is set, this function will be called if no value
// exists for the given key.
// This operation is thread-safe.
func Get(c *template.Context, vars ...func() interface{}) (interface{}, bool, error) {
	// Get the store.
	s, err := getStore(c)
	if err != nil {
		return nil, false, err
	}

	// The key is the context's ID.
	id := c.ID()

	// Lock the mutex.
	s.mutex.Lock()
	defer s.mutex.Unlock()

	// Create the map if it doesn't exists.
	s.data.createMapIfNil()

	// Get the value.
	data, ok := s.data.Values[id]
	if ok {
		return data.Data, true, nil
	}

	// If no data is found and the create function variable
	// is set, then call the function and set the new value.
	if len(vars) > 0 {
		// The context has to be locked to perform any changes.
		if !IsLocked(c) {
			return nil, false, fmt.Errorf("store.Get: create function: can't set store data if the context is not locked by the session!")
		}

		value := vars[0]()
		s.data.Values[id] = newDBStoreData(value)

		// Update data to the database.
		err = flushUpdatesToDB(s)
		if err != nil {
			return nil, false, err
		}

		return value, true, nil
	}

	return nil, false, nil
}

// Set the context value to the store.
// The data is also flushed to the database.
// This operation is thread-safe.
func Set(c *template.Context, value interface{}) error {
	// Get the store.
	s, err := getStore(c)
	if err != nil {
		return err
	}

	// The context has to be locked to perform any changes.
	if !IsLocked(c) {
		return fmt.Errorf("store.Set: can't set store data if the context is not locked by the session!")
	}

	// Lock the mutex.
	s.mutex.Lock()
	defer s.mutex.Unlock()

	// Create the map if it doesn't exists.
	s.data.createMapIfNil()

	// Set the value. The key is the context's ID.
	s.data.Values[c.ID()] = newDBStoreData(value)

	// Update data to the database.
	return flushUpdatesToDB(s)
}

// Delete removes the context value from the store.
// This operation is thread-safe.
func Delete(c *template.Context) error {
	// Get the store.
	s, err := getStore(c)
	if err != nil {
		return err
	}

	// The context has to be locked to perform any changes.
	if !IsLocked(c) {
		return fmt.Errorf("store.Delete: can't remove store data if the context is not locked by the session!")
	}

	// Lock the mutex.
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if s.data.Values == nil {
		// Nothing to do.
		return nil
	}

	// Remove the value. The key is the context's ID.
	delete(s.data.Values, c.ID())

	// Update data to the database.
	return flushUpdatesToDB(s)
}

//###############//
//### Private ###//
//###############//

// getStore returns the store for the current context.
// This operation is thread-safe.
func getStore(c *template.Context) (st *store, err error) {
	// Recover panics and return the error message.
	defer func() {
		if e := recover(); e != nil {
			err = fmt.Errorf("%v", e)
		}
	}()

	// The store ID is the context's root ID.
	id := c.RootID()

	// Create the context execution value key.
	cKey := contextValueKeyStorePrefix + id

	// If the store was already previously obtained and added
	// to the context execution values, then use this value
	// instead of getting it again from the database.
	storeI, _ := c.Get(cKey, func() interface{} {
		// Get the session pointer.
		session := c.Session()

		// Get a boolean if the edit mode is active for
		// the current session.
		editModeActive := editmode.IsActive(session)

		// The store was not found in the context
		// execution values. Obtain it...
		existsInDB := true
		data, err := dbGetStore(id, editModeActive)
		if err != nil {
			panic(err)
		} else if data == nil {
			// Create a new db store.
			data = newDBStore(id)
			existsInDB = false
		}

		// Create a new store value.
		s := newStore(data)

		// Set the flags.
		s.editModeActive = editModeActive
		s.existsInDB = existsInDB

		return s
	})

	// Assert and return.
	return storeI.(*store), nil
}

func flushUpdatesToDB(s *store) error {
	if s.existsInDB {
		// Update the store in the database.
		err := dbUpdateStore(s.data, s.editModeActive)
		if err != nil {
			return fmt.Errorf("failed to update the store data to the database: %v", err)
		}
	} else {
		// Insert the store to the database.
		err := dbInsertStore(s.data, s.editModeActive)
		if err != nil {
			return fmt.Errorf("failed to insert the store data to the database: %v", err)
		}

		// Update the flag.
		s.existsInDB = true
	}

	return nil
}
