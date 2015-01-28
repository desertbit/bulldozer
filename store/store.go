/*
 *  Bulldozer Framework
 *  Copyright (C) DesertBit
 */

package store

import (
	"code.desertbit.com/bulldozer/bulldozer/editmode"
	"code.desertbit.com/bulldozer/bulldozer/log"
	"code.desertbit.com/bulldozer/bulldozer/template"
	"fmt"
	"sync"
)

const (
	// Context Execution keys.
	contextValueKeyStorePrefix = "blzStore_"
)

var (
	lockMutex sync.Mutex
)

//##################//
//### Store type ###//
//##################//

type store struct {
	data           *dbStore
	mutex          sync.Mutex
	editModeActive bool
}

func newStore(data *dbStore) *store {
	return &store{
		data:           data,
		editModeActive: false,
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
	instanceID := c.Session().InstanceID()

	// Lock the mutex.
	lockMutex.Lock()
	defer lockMutex.Unlock()

	// Check if already locked by another session.
	locked, err := dbIsLockedByAnotherValue(id, instanceID)
	if err != nil {
		log.L.Error("store: failed to lock context: %v", err)
		return false
	} else if locked {
		return false
	}

	// Lock the context ID for the current session.
	err = dbLock(id, instanceID)
	if err != nil {
		log.L.Error("store: failed to lock context: %v", err)
		return false
	}

	// Broadcast changes to other sessions in edit mode.
	go broadcastChangedContext(c)

	return true
}

// Unlock the context again.
// This operation is thread-safe.
func Unlock(c *template.Context) {
	id := c.ID()
	instanceID := c.Session().InstanceID()

	// Lock the mutex.
	lockMutex.Lock()
	defer lockMutex.Unlock()

	// Check if locked by another session.
	locked, err := dbIsLockedByAnotherValue(id, instanceID)
	if err != nil {
		log.L.Error("store: failed to unlock context: %v", err)
		return
	} else if locked {
		log.L.Error("store: failed to unlock store context: the calling session is not the session which holds the lock!")
		return
	}

	// Unlock the lock.
	err = dbUnlock(id)
	if err != nil {
		log.L.Error("store: failed to unlock context: %v", err)
		return
	}

	// Broadcast changes to other sessions in edit mode.
	go broadcastChangedContext(c)
}

// IsLocked returns a boolean whenever the context is
// locked by the current session.
// This operation is thread-safe.
func IsLocked(c *template.Context) bool {
	id := c.ID()
	instanceID := c.Session().InstanceID()

	// Lock the mutex.
	lockMutex.Lock()
	defer lockMutex.Unlock()

	// Check if locked by the current session.
	locked, err := dbIsLocked(id, instanceID)
	if err != nil {
		log.L.Error("store: failed to get lock state: %v", err)
		return false
	}

	return locked
}

// IsLockedByAnotherSession returns a boolean whenever the context is
// locked by another session.
// This operation is thread-safe.
func IsLockedByAnotherSession(c *template.Context) bool {
	id := c.ID()
	instanceID := c.Session().InstanceID()

	// Lock the mutex.
	lockMutex.Lock()
	defer lockMutex.Unlock()

	// Check if locked by the current session.
	locked, err := dbIsLockedByAnotherValue(id, instanceID)
	if err != nil {
		log.L.Error("store: failed to get lock state: %v", err)
		return true
	}

	return locked
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

		// Broadcast changes to other sessions in edit mode.
		go broadcastChangedContext(c)

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
	err = flushUpdatesToDB(s)
	if err != nil {
		return err
	}

	// Broadcast changes to other sessions in edit mode.
	go broadcastChangedContext(c)

	return nil
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
	err = flushUpdatesToDB(s)
	if err != nil {
		return err
	}

	// Broadcast changes to other sessions in edit mode.
	go broadcastChangedContext(c)

	return nil
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
		data, err := dbGetStore(id, editModeActive)
		if err != nil {
			panic(err)
		} else if data == nil {
			// Create a new db store.
			data = newDBStore(id)
		}

		// Create a new store value.
		s := newStore(data)

		// Set the flags.
		s.editModeActive = editModeActive

		return s
	})

	// Assert and return.
	return storeI.(*store), nil
}

func flushUpdatesToDB(s *store) error {
	// Update the store in the database.
	err := dbUpdateStore(s.data, s.editModeActive)
	if err != nil {
		return fmt.Errorf("failed to update the store data to the database: %v", err)
	}

	return nil
}

func broadcastChangedContext(c *template.Context) {
	// Get the session ID of the current session.
	curSid := c.Session().SessionID()

	// Get the current context ID.
	id := c.ID()

	// Get all sessions which are in the edit mode.
	activeSessions := editmode.GetSessions()

	var err error
	for _, s := range activeSessions {
		// Skip if this is the current session.
		if s.SessionID() == curSid {
			continue
		}

		// Get the context store of the session.
		store := template.GetSessionContextStore(s)
		if store == nil {
			log.L.Error("failed to update session context with ID '%s': failed to get context store!", id)
			// TODO: log error and refresh the sessions page!
			continue
		}

		cc, ok := store.Get(id)
		if !ok {
			log.L.Error("failed to update session context with ID '%s': failed to get context!", id)
			// TODO: log error and refresh the sessions page!
			continue
		}

		err = cc.Update()
		if err != nil {
			log.L.Error("failed to update session context with ID '%s': %v", id, err)
			// TODO: log error and refresh the sessions page!
			continue
		}
	}
}
