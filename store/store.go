/*
 *  Bulldozer Framework
 *  Copyright (C) DesertBit
 */

package store

import (
	tr "code.desertbit.com/bulldozer/bulldozer/translate"

	"code.desertbit.com/bulldozer/bulldozer/editmode"
	"code.desertbit.com/bulldozer/bulldozer/log"
	"code.desertbit.com/bulldozer/bulldozer/sessions"
	"code.desertbit.com/bulldozer/bulldozer/template"
	"code.desertbit.com/bulldozer/bulldozer/ui/messagebox"
	"fmt"
	"strconv"
	"sync"
	"time"
)

const (
	cleanupLocksTimeout            = 30 * time.Second
	removeExpiredLocksAfterTimeout = 10 * time.Second

	// Value keys.
	clientKeyStoreState        = "blzStoreState"
	contextValueKeyStorePrefix = "blzStore_"

	saveTemporaryChangesCallback = "blzSaveTmpChanges"
)

var (
	backend bulldozerBackend

	lockMutex            sync.Mutex
	stopCleanupLocksLoop chan struct{} = make(chan struct{})
)

//###################################//
//### Bulldozer backend interface ###//
//###################################//

type bulldozerBackend interface {
	ReloadPage(s *sessions.Session)
}

//##################//
//### Store type ###//
//##################//

type store struct {
	data  *dbStore
	mutex sync.Mutex
}

func newStore(data *dbStore) *store {
	return &store{
		data: data,
	}
}

//##############//
//### Public ###//
//##############//

func Init(b bulldozerBackend) error {
	// Set the backend.
	backend = b

	// Attach the event listeners.
	editmode.OnNewSession(onNewEditModeSession)
	editmode.OnSessionReconnect(onEditModeSessionReconnect)

	// Register the messagebox callback.
	messagebox.RegisterCallback(saveTemporaryChangesCallback, saveTemporaryChanges)

	// Initialize the database.
	err := initDB()
	if err != nil {
		return err
	}

	// Start the cleanup loop in a new goroutine.
	go cleanupLocksLoop()

	return nil
}

func Release() {
	// Stop the loop by triggering the quit trigger
	close(stopCleanupLocksLoop)
}

// Lock the context for the current session.
// A boolean is returned, if the lock request was successful.
// This operation is thread-safe.
func Lock(c *template.Context) bool {
	id := c.ID()
	session := c.Session()
	instanceID := session.InstanceID()

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
	broadcastChangedContext(id, session)

	return true
}

// Unlock the context again.
// This operation is thread-safe.
func Unlock(c *template.Context) {
	id := c.ID()
	session := c.Session()
	instanceID := session.InstanceID()

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
	broadcastChangedContext(id, session)
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
		broadcastChangedContext(id, c.Session())

		return value, true, nil
	}

	return nil, false, nil
}

// Set the context value to the store.
// The data is also flushed to the database.
// This operation is thread-safe.
func Set(c *template.Context, value interface{}) error {
	id := c.ID()

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
	s.data.Values[id] = newDBStoreData(value)

	// Update data to the database.
	err = flushUpdatesToDB(s)
	if err != nil {
		return err
	}

	// Broadcast changes to other sessions in edit mode.
	broadcastChangedContext(id, c.Session())

	return nil
}

// Delete removes the context value from the store.
// This operation is thread-safe.
func Delete(c *template.Context) error {
	id := c.ID()

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
	delete(s.data.Values, id)

	// Update data to the database.
	err = flushUpdatesToDB(s)
	if err != nil {
		return err
	}

	// Broadcast changes to other sessions in edit mode.
	broadcastChangedContext(id, c.Session())

	return nil
}

func SaveTemporaryChanges(s *sessions.Session) {
	// Check if there are temporary changes to save.
	hasChanges, err := dbHasTemporaryChanges()
	if err != nil {
		// Show a messagebox.
		messagebox.New().
			SetTitle(tr.S("blz.core.saveFailedTitle")).
			SetText(tr.S("blz.core.saveFailedText")).
			SetType(messagebox.TypeAlert).
			Show(s)
		return
	} else if !hasChanges {
		// Show a messagebox.
		messagebox.New().
			SetTitle(tr.S("blz.core.nothingToSaveTitle")).
			SetText(tr.S("blz.core.nothingToSaveText")).
			SetType(messagebox.TypeInfo).
			Show(s)
		return
	}

	// Show a messagebox to continue.
	messagebox.New().
		SetTitle(tr.S("blz.core.saveChangesTitle")).
		SetText(tr.S("blz.core.saveChangesText")).
		SetType(messagebox.TypeQuestion).
		SetButtons(messagebox.ButtonYes | messagebox.ButtonNo).
		SetCallback(saveTemporaryChangesCallback).
		Show(s)
}

//###############//
//### Private ###//
//###############//

func saveTemporaryChanges(s *sessions.Session, b messagebox.Button) {
	if b != messagebox.ButtonYes {
		return
	}

	// Check if anything is
	locks, err := dbGetLocks()
	if err != nil {
		log.L.Error("failed to get all locks: %v", err)

		// Show a messagebox.
		messagebox.New().
			SetTitle(tr.S("blz.core.saveFailedTitle")).
			SetText(tr.S("blz.core.saveFailedText")).
			SetType(messagebox.TypeAlert).
			Show(s)
		return
	} else if len(locks) > 0 {
		// Show a messagebox.
		messagebox.New().
			SetTitle(tr.S("blz.core.objectsLockedTitle")).
			SetText(tr.S("blz.core.objectsLockedText")).
			SetType(messagebox.TypeWarning).
			Show(s)
		return
	}

	// Save all temporary changes
	err = dbSaveTemporaryChanges()
	if err != nil {
		log.L.Error("failed to save temporary changes: %v", err)

		// Show a messagebox.
		messagebox.New().
			SetTitle(tr.S("blz.core.saveFailedTitle")).
			SetText(tr.S("blz.core.saveFailedText")).
			SetType(messagebox.TypeAlert).
			Show(s)
		return
	}

	// Show a success messagebox.
	messagebox.New().
		SetTitle(tr.S("blz.core.successSaveTitle")).
		SetText(tr.S("blz.core.successSaveText")).
		SetType(messagebox.TypeSuccess).
		Show(s)
}

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

		return s
	})

	// Assert and return.
	return storeI.(*store), nil
}

func flushUpdatesToDB(s *store) error {
	// Update the temporary store in the database.
	err := dbUpdateStore(s.data, true)
	if err != nil {
		return fmt.Errorf("failed to update the store data to the database: %v", err)
	}

	return nil
}

func broadcastChangedContext(contextID string, currentSession ...*sessions.Session) {
	// Update the last changed timestamp.
	timestamp, err := dbUpdateStoreInfoLastChanged(dbTmpStoreTable)
	if err != nil {
		log.L.Error("failed to update session store information in database: %v", err)
		return
	}

	// Get the session ID of the current session if passed.
	var curSid string
	if len(currentSession) > 0 {
		curSid = currentSession[0].SessionID()
	}

	// Start this in a new go-routine. Don't block any calling mutexes...
	go func() {
		// Get all sessions which are in the edit mode.
		activeSessions := editmode.GetSessions()

		for _, s := range activeSessions {
			// Update the store state in the client values.
			// This way, we can detect, if a session is out-of-sync on reconnect.
			s.ClientSet(clientKeyStoreState, strconv.FormatInt(timestamp, 10))

			// Skip if this is the current session.
			if s.SessionID() == curSid {
				continue
			}

			// Recover panics and log the error message.
			defer func() {
				if e := recover(); e != nil {
					log.L.Warning("update session context with ID '%s': panic: %v", contextID, e)
				}
			}()

			// Get the context store of the session.
			store := template.GetSessionContextStore(s)
			if store == nil {
				log.L.Warning("failed to update session context with ID '%s': failed to get context store!", contextID)
				sessionOutOfSync(s)
				return
			}

			cc, ok := store.Get(contextID)
			if !ok {
				log.L.Warning("failed to update session context with ID '%s': failed to get context!", contextID)
				sessionOutOfSync(s)
				return
			}

			err = cc.Update()
			if err != nil {
				log.L.Warning("failed to update session context with ID '%s': %v", contextID, err)
				sessionOutOfSync(s)
				return
			}
		}
	}()
}

func onNewEditModeSession(s *sessions.Session) {
	// Get the store info.
	info, err := dbGetStoreInfo(dbTmpStoreTable)
	if err != nil {
		log.L.Error("store on new edit mode session: %v", err)
		return
	} else if info == nil {
		// No info data present.
		return
	}

	// Update the store state in the client values.
	// This way, we can detect, if a session is out-of-sync on reconnect.
	s.ClientSet(clientKeyStoreState, strconv.FormatInt(info.LastChange, 10))
}

func onEditModeSessionReconnect(s *sessions.Session) {
	// Get the store info.
	info, err := dbGetStoreInfo(dbTmpStoreTable)
	if err != nil {
		log.L.Error("store on edit mode session reconnect: %v", err)
		return
	} else if info == nil {
		// No info data present.
		return
	}

	// Get the session store state.
	s.ClientGet(clientKeyStoreState, func(data string) {
		// Check if the session state is out of sync.
		if data != strconv.FormatInt(info.LastChange, 10) {
			sessionOutOfSync(s)
		}
	})
}

func sessionOutOfSync(s *sessions.Session) {
	// Reload the page to get in sync again.
	backend.ReloadPage(s)

	// Show a messagebox
	messagebox.New().
		SetTitle(tr.S("blz.core.sessionOutOfSyncTitle")).
		SetText(tr.S("blz.core.sessionOutOfSyncText")).
		SetType(messagebox.TypeWarning).
		Show(s)
}

func cleanupLocksLoop() {
	// Create a new ticker
	ticker := time.NewTicker(cleanupLocksTimeout)

	defer func() {
		// Stop the ticker
		ticker.Stop()
	}()

	for {
		select {
		case <-ticker.C:
			cleanupLocks()
		case <-stopCleanupLocksLoop:
			// Just exit the loop
			return
		}
	}
}

func cleanupLocks() {
	// Just skip this check if there are no active edit mode sessions.
	// This removes some overhead.
	if !editmode.HasActiveSessions() {
		return
	}

	// Recover panics and log the error message.
	defer func() {
		if e := recover(); e != nil {
			log.L.Error("store cleanup locks: panic: %v", e)
		}
	}()

	// Get all the locks from the database.
	locks, err := dbGetLocks()
	if err != nil {
		log.L.Error("store cleanup locks: failed to obtain locks from database: %v", err)
		return
	}

	if len(locks) == 0 {
		return
	}

	// Get all sessions which are in the edit mode
	// and create a session map sorted with instance IDs.
	getActiveInstanceIDs := func() map[string]*sessions.Session {
		activeSessions := editmode.GetSessions()
		m := make(map[string]*sessions.Session)
		for _, s := range activeSessions {
			m[s.InstanceID()] = s
		}
		return m
	}
	activeInstanceIDs := getActiveInstanceIDs()

	// Find locks which are not locked by an active session.
	var expiredLocks []*dbLockData
	var found bool
	for _, lock := range locks {
		_, found = activeInstanceIDs[lock.Value]
		if !found {
			expiredLocks = append(expiredLocks, lock)
		}
	}

	if len(expiredLocks) == 0 {
		return
	}

	// Give the sessions a chance to reconnect.
	time.Sleep(removeExpiredLocksAfterTimeout)

	// Update the active sessions.
	activeInstanceIDs = getActiveInstanceIDs()

	for _, lock := range expiredLocks {
		_, found = activeInstanceIDs[lock.Value]
		if found {
			continue
		}

		// Unlock the lock.
		dbUnlock(lock.ID)

		// Broadcast changes to other sessions in edit mode.
		broadcastChangedContext(lock.ID)
	}
}
