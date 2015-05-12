/*
 *  Bulldozer Framework
 *  Copyright (C) DesertBit
 */

package template

import (
	"encoding/gob"
	"github.com/desertbit/bulldozer/log"
	"github.com/desertbit/bulldozer/sessions"
	"sync"
)

const (
	// Session instance value keys:
	instanceKeyContextStore = "budContextStore"
)

func init() {
	// Register the custom types.
	gob.Register(&contextStore{})
}

//#############//
//### Types ###//
//#############//

type contextStore struct {
	ContextDataMap map[string]*contextData
	mutex          sync.Mutex
	s              *sessions.Session
}

func newContextStore(s *sessions.Session) *contextStore {
	return &contextStore{
		ContextDataMap: make(map[string]*contextData),
		s:              s,
	}
}

func (cs *contextStore) Get(id string) (d *contextData, ok bool) {
	// Lock the mutex.
	cs.mutex.Lock()
	defer cs.mutex.Unlock()

	// Get the data from the map if present.
	d, ok = cs.ContextDataMap[id]
	return
}

func (cs *contextStore) Set(d *contextData) {
	// Lock the mutex.
	cs.mutex.Lock()
	defer cs.mutex.Unlock()

	// Add the data to the map.
	cs.ContextDataMap[d.ID] = d

	// Tell the session, that data has changed.
	cs.s.Dirty()
}

func (cs *contextStore) Remove(d *contextData) {
	// Lock the mutex.
	cs.mutex.Lock()
	defer cs.mutex.Unlock()

	// Remove the data again from the map.
	delete(cs.ContextDataMap, d.ID)

	// Tell the session, that data has changed.
	cs.s.Dirty()
}

func (cs *contextStore) Reset() {
	// Lock the mutex.
	cs.mutex.Lock()
	defer cs.mutex.Unlock()

	// Reset the map.
	cs.ContextDataMap = make(map[string]*contextData)

	// Tell the session, that data has changed.
	cs.s.Dirty()
}

//##########################//
//### Context Store type ###//
//##########################//

type ContextStore struct {
	store *contextStore
}

// Get a context from the store by it's ID if present.
func (cs *ContextStore) Get(id string) (*Context, bool) {
	// Get the context data.
	d, ok := cs.store.Get(id)
	if !ok {
		return nil, false
	}

	c, err := newContextFromData(cs.store.s, d, false)
	if err != nil {
		log.L.Error("context store get: %v", err)
		return nil, false
	}

	return c, true
}

//##############//
//### Public ###//
//##############//

// EnableSessionContextStore enables the context store for the session.
// All current active contexts of the session will be saved to
// the sessions instance values. This adds some overhead during each render request.
// Only activate this if you need access to the sessions contexts.
func EnableSessionContextStore(s *sessions.Session) {
	// We'll use the Get method here.
	// If not present, a new context store will be created.
	s.InstanceGet(instanceKeyContextStore, func() interface{} {
		return newContextStore(s)
	})
}

// DisableSessionContextStore disables the context store for the session.
func DisableSessionContextStore(s *sessions.Session) {
	// Remove the store again if present.
	s.InstanceDelete(instanceKeyContextStore)
}

// GetSessionContextStore returns the session context store if present or nil.
func GetSessionContextStore(s *sessions.Session) *ContextStore {
	// Get the store.
	store := getContextStore(s)
	if store == nil {
		return nil
	}

	// Create the public wrapper.
	return &ContextStore{
		store: store,
	}
}

//###############//
//### Private ###//
//###############//

func getContextStore(s *sessions.Session) *contextStore {
	i, ok := s.InstanceGet(instanceKeyContextStore)
	if !ok {
		return nil
	}

	// Assertion
	store, ok := i.(*contextStore)
	if !ok {
		log.L.Error("failed to assert context store from session instance value!")
		return nil
	}

	// Set the session. This might be nil,
	// if this store is obtained from the store
	// after a fresh application restart.
	store.s = s

	return store
}

func resetContextStore(s *sessions.Session) {
	// Get the store.
	store := getContextStore(s)

	// Reset the store if present.
	if store != nil {
		store.Reset()
	}
}
