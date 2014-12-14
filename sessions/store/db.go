/*
 *  Bulldozer Framework
 *  Copyright (C) DesertBit
 */

package store

import (
	"bytes"
	"code.desertbit.com/bulldozer/bulldozer/sessions/store/protobuf"
	"code.desertbit.com/bulldozer/bulldozer/settings"
	"code.google.com/p/gogoprotobuf/proto"
	"encoding/gob"
	"errors"
	"fmt"
	"github.com/boltdb/bolt"
	"github.com/golang/glog"
	"sync"
	"time"
)

const (
	saveLoopTimeout       = 1 * time.Minute
	cleanupExpiredTimeout = 2 * time.Minute

	cleanupExpiredSessionsBatchSize = 100

	bucketName = "s"
)

var (
	// Public
	ErrNotFound = errors.New("store: the session with the corresponding ID does not exists")

	// Private
	db              *bolt.DB
	bucketNameBytes = []byte(bucketName)

	stopSaveLoop      chan struct{} = make(chan struct{})
	stopCleanupDBLoop chan struct{} = make(chan struct{})

	// The previous database iteration key for scanning for expired sessions
	prevExpiredScanKey []byte

	changedSessions      map[string]*Session = make(map[string]*Session)
	changedSessionsMutex sync.Mutex

	removeSessionIDs      []string
	removeSessionIDsMutex sync.Mutex
)

//##############//
//### Public ###//
//##############//

// Init initializes this store package.
// This is handled by the main bulldozer package.
func Init() {
	// Open the sessions database file.
	// It will be created if it doesn't exist.
	var err error
	db, err = bolt.Open(settings.Settings.SessionsDatabasePath, 0600, nil)
	if err != nil {
		glog.Fatalf("failed to open sessions database '%s': %v", settings.Settings.SessionsDatabasePath, err)
	}

	// Create the bucket if not already exists
	err = db.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists(bucketNameBytes)
		return err
	})
	if err != nil {
		glog.Fatalf("failed to create the sessions database bucket: %v", err)
	}

	// Start the loops in a new goroutine
	go saveLoop()
	go cleanupDBLoop()
}

// Release releases this store package.
// This is handled by the main bulldozer package.
func Release() {
	if db == nil {
		return
	}

	// Stop the loops by triggering the quit trigger
	close(stopSaveLoop)
	close(stopCleanupDBLoop)

	// Finally save all unsaved sessions before exiting
	saveUnsavedSessions()

	// Remove all the manual removed sessions without scanning for expired sessions
	cleanupDBSessions(true)

	// Close the database on exit
	db.Close()
}

//###############//
//### Private ###//
//###############//

type dbSessionBuffer struct {
	id    []byte
	value []byte
}

// registerChangedSession notifies the daemon to save the sessions' changes
func registerChangedSession(s *Session) {
	// Start this in a new goroutine to not block the calling function...
	go func() {
		// Lock the mutex
		changedSessionsMutex.Lock()
		defer changedSessionsMutex.Unlock()

		// Add the session pointer to the map
		changedSessions[s.id] = s
	}()
}

func saveLoop() {
	// Create a new ticker
	ticker := time.NewTicker(saveLoopTimeout)

	defer func() {
		// Stop the ticker
		ticker.Stop()
	}()

	for {
		select {
		case <-ticker.C:
			// Save all unsaved sessions
			saveUnsavedSessions()
		case <-stopSaveLoop:
			// Just exit the loop
			return
		}
	}
}

func saveUnsavedSessions() {
	// Skip if the session max age is not set
	if settings.Settings.SessionMaxAge <= 0 {
		return
	}

	err := func() (err error) {
		// Lock the mutex
		changedSessionsMutex.Lock()

		// Return if the map is empty
		if len(changedSessions) == 0 {
			// Unlock the mutex again befure returning
			changedSessionsMutex.Unlock()
			return nil
		}

		// Create a copy of the changed sessions map
		tmpChangedSessions := changedSessions

		// Clear the changed sessions map
		changedSessions = make(map[string]*Session)

		// Unlock the mutex again
		changedSessionsMutex.Unlock()

		// Create a temporary database buffer for the batched write procedure
		var dbBuffer []dbSessionBuffer

		// Create the expire timestamp
		expiresAt := time.Now().Unix() + int64(settings.Settings.SessionMaxAge)

		// Iterate over all changed session and save them to the database
		for _, s := range tmpChangedSessions {
			// Skip if this session if flagged as invalid
			if !s.valid {
				continue
			}

			// Prepare the session values data to be encoded
			var buf bytes.Buffer
			enc := gob.NewEncoder(&buf)

			func() {
				// Lock the mutex
				s.mutex.Lock()
				defer s.mutex.Unlock()

				// Encode the data
				err = enc.Encode(s.values)
			}()

			// Catch any encoding error
			if err != nil {
				return
			}

			// Create a new proto session
			protoSession := &protobuf.Session{
				Values:    buf.Bytes(),
				ExpiresAt: &expiresAt,
			}

			// Marshal the proto session to a bytes slice
			var data []byte
			data, err = proto.Marshal(protoSession)
			if err != nil {
				return
			}

			// Add the data to the temporary database buffer
			dbBuffer = append(dbBuffer, dbSessionBuffer{[]byte(s.id), data})
		}

		// Now save everything to the database
		err = db.Update(func(tx *bolt.Tx) (err error) {
			// Get the bucket
			b := tx.Bucket(bucketNameBytes)
			if b == nil {
				return fmt.Errorf("no bucket '%s' found!", bucketName)
			}

			// Save all the buffered sessions data
			for _, buf := range dbBuffer {
				err = b.Put(buf.id, buf.value)
				if err != nil {
					return err
				}
			}

			return nil
		})

		return
	}()

	if err != nil {
		glog.Errorf("sessions database save error: %v", err)
	}
}

func cleanupDBLoop() {
	// Create a new ticker
	ticker := time.NewTicker(cleanupExpiredTimeout)

	defer func() {
		// Stop the ticker
		ticker.Stop()
	}()

	for {
		select {
		case <-ticker.C:
			// Cleanup all expired and deleted sessions from the database
			cleanupDBSessions(false)
		case <-stopCleanupDBLoop:
			// Just exit the loop
			return
		}
	}
}

func cleanupDBSessions(skipExpiredSessions bool) {
	var err error
	expiredSessionIDs := make([][]byte, 0)

	if !skipExpiredSessions {
		// Cleanup all expired database sessions
		err = db.View(func(tx *bolt.Tx) error {
			// Get the bucket
			b := tx.Bucket(bucketNameBytes)
			if b == nil {
				return fmt.Errorf("no bucket '%s' found!", bucketName)
			}

			c := b.Cursor()
			i := 0
			var isExpired bool

			for k, v := c.Seek(prevExpiredScanKey); ; k, v = c.Next() {
				// If we hit the end of our sessions then
				// exit and start over next time.
				if k == nil {
					prevExpiredScanKey = nil
					return nil
				}

				// Increment the counter
				i++

				// The flag if the session is expired
				isExpired = false

				// Get the proto session value from the session data
				// and check if the session is expired.
				protoSession, err := getProtoSession(v)
				if err != nil {
					// Just remove the session with the invalid session data
					isExpired = true
				} else if protoSessionExpired(protoSession) {
					isExpired = true
				}

				if isExpired {
					// Copy the byte slice key, because this data is
					// not safe outside of this transaction.
					temp := make([]byte, len(k))
					copy(temp, k)

					// Add it to the expired sessios IDs slice
					expiredSessionIDs = append(expiredSessionIDs, temp)
				}

				if i >= cleanupExpiredSessionsBatchSize {
					// Store the current key to the previous key.
					// Copy the byte slice key, because this data is
					// not safe outside of this transaction.
					prevExpiredScanKey = make([]byte, len(k))
					copy(prevExpiredScanKey, k)
					return nil
				}
			}

			return nil
		})

		if err != nil {
			glog.Errorf("sessions database: obtain expired sessions error: %v", err)
		}
	}

	// Add all session IDs to the expired map,
	// which should be removed from the database.
	if len(removeSessionIDs) > 0 {
		func() {
			// Lock the mutex
			removeSessionIDsMutex.Lock()
			defer removeSessionIDsMutex.Unlock()

			for _, id := range removeSessionIDs {
				expiredSessionIDs = append(expiredSessionIDs, []byte(id))
			}

			// Clear the slice again
			removeSessionIDs = nil
		}()
	}

	if len(expiredSessionIDs) > 0 {
		// Remove the expired sessions from the database
		err = db.Update(func(tx *bolt.Tx) error {
			// Get the bucket
			b := tx.Bucket(bucketNameBytes)
			if b == nil {
				return fmt.Errorf("no bucket '%s' found!", bucketName)
			}

			// Remove all expired sessions in the slice
			for _, id := range expiredSessionIDs {
				err = b.Delete(id)
				if err != nil {
					return err
				}
			}

			return nil
		})

		if err != nil {
			glog.Errorf("sessions database: remove expired sessions error: %v", err)
		}
	}
}

func getSessionFromDB(id string) (*Session, error) {
	// Check if the ID is flagged to be removed
	if idIsRemoved(id) {
		return nil, ErrNotFound
	}

	var values map[interface{}]interface{}

	// Try to obtain the session from the database
	err := db.View(func(tx *bolt.Tx) error {
		idb := []byte(id)

		// Get the bucket
		b := tx.Bucket(bucketNameBytes)
		if b == nil {
			return fmt.Errorf("no bucket '%s' found!", bucketName)
		}

		// Obtain the session data
		data := b.Get(idb)
		if data == nil {
			return ErrNotFound
		}

		// Get the proto session value from the session data
		protoSession, err := getProtoSession(data)
		if err != nil {
			return err
		}

		// Check if the session is epxired
		if protoSessionExpired(protoSession) {
			// This session is expired. Just return a not found error.
			// The cleanupExpiredLoop will handle deletion of it.
			return ErrNotFound
		}

		// Decode the session data and set the values map
		dec := gob.NewDecoder(bytes.NewBuffer(protoSession.GetValues()))
		err = dec.Decode(&values)
		if err != nil {
			return fmt.Errorf("failed to gob decode session database values: %v", err)
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	// Create a new session and set the values map
	s := &Session{
		id:          id,
		valid:       true,
		values:      values,
		cacheValues: make(map[interface{}]interface{}),
	}

	return s, nil
}

func removeSessionFromDB(id string) {
	// Lock the mutex
	removeSessionIDsMutex.Lock()
	defer removeSessionIDsMutex.Unlock()

	// Add the id to the slice
	removeSessionIDs = append(removeSessionIDs, id)

	// Lock the mutex
	changedSessionsMutex.Lock()
	defer changedSessionsMutex.Unlock()

	// Remove the session also from the changed sessions map if present
	delete(changedSessions, id)
}

// idIsRemoved checks if the ID is flagged to be removed
func idIsRemoved(id string) bool {
	// Lock the mutex
	removeSessionIDsMutex.Lock()
	defer removeSessionIDsMutex.Unlock()

	// Check if the ID is in the slice for the removed session IDs
	for _, rId := range removeSessionIDs {
		if id == rId {
			return true
		}
	}

	return false
}

func sessionIDExistsInDB(id string) (exists bool, err error) {
	exists = true

	// Check if the ID exists in the database
	err = db.View(func(tx *bolt.Tx) error {
		// Get the bucket
		b := tx.Bucket(bucketNameBytes)
		if b == nil {
			return fmt.Errorf("no bucket '%s' found!", bucketName)
		}

		// Try to obtain the session data
		data := b.Get([]byte(id))
		if data == nil {
			exists = false
		}

		return nil
	})

	return
}

// getProtoSession converts the byte slice to the proto session struct
func getProtoSession(data []byte) (s *protobuf.Session, err error) {
	s = &protobuf.Session{}
	err = proto.Unmarshal(data, s)
	return
}

// protoSessionExpired checks if the session is expired.
func protoSessionExpired(s *protobuf.Session) bool {
	expiresAt := s.GetExpiresAt()

	// The session is expired if the value is invalid
	if expiresAt <= 0 {
		return true
	}

	return expiresAt <= time.Now().Unix()
}
