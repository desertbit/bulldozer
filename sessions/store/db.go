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
	cleanupExpiredTimeout = 1 * time.Minute

	bucketName = "s"
)

var (
	// Public
	ErrNotFound = errors.New("store: the session with the corresponding ID does not exists")

	// Private
	db              *bolt.DB
	bucketNameBytes = []byte(bucketName)

	changedSessions      map[string]*Session = make(map[string]*Session)
	changedSessionsMutex sync.Mutex

	stopSaveLoop           chan struct{} = make(chan struct{})
	stopCleanupExpiredLoop chan struct{} = make(chan struct{})
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
	go cleanupExpiredLoop()
}

// Release releases this store package.
// This is handled by the main bulldozer package.
func Release() {
	if db == nil {
		return
	}

	// Close the database on exit
	db.Close()

	// Stop the loops by triggering the quit trigger
	close(stopSaveLoop)
	close(stopCleanupExpiredLoop)

	// Finally save all unsaved sessions before exiting
	saveUnsavedSessions()
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
	// Lock the mutex
	changedSessionsMutex.Lock()
	defer changedSessionsMutex.Unlock()

	// Add the session pointer to the map
	changedSessions[s.id] = s
}

func saveLoop() {
	// Create a new timer
	timer := time.NewTimer(saveLoopTimeout)

	defer func() {
		// Stop the timer
		timer.Stop()
	}()

	for {
		select {
		case <-timer.C:
			// Save all unsaved sessions
			saveUnsavedSessions()
		case <-stopSaveLoop:
			// Just exit the loop
			return
		}
	}
}

func saveUnsavedSessions() {
	err := func() (err error) {
		// Lock the mutex
		changedSessionsMutex.Lock()

		// Get the length of the map and return if it is empty
		l := len(changedSessions)
		if l == 0 {
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
		dbBuffer := make([]dbSessionBuffer, l)

		// Iterate over all changed session and save them to the database
		for id, s := range tmpChangedSessions {
			// Prepare the data
			var buf bytes.Buffer
			enc := gob.NewEncoder(&buf)
			err = enc.Encode(s.values)
			if err != nil {
				return
			}
			var data []byte
			data, err = newProtoSession(buf.Bytes())
			if err != nil {
				return
			}

			// Add the data to the temporary database buffer
			dbBuffer = append(dbBuffer, dbSessionBuffer{[]byte(id), data})
		}

		// Now save everything to the database
		err = db.Update(func(tx *bolt.Tx) (err error) {
			// Get the bucket
			b := tx.Bucket(bucketNameBytes)

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

func cleanupExpiredLoop() {
	// Create a new timer
	timer := time.NewTimer(cleanupExpiredTimeout)

	defer func() {
		// Stop the timer
		timer.Stop()
	}()

	for {
		select {
		case <-timer.C:
			// TODO
		case <-stopCleanupExpiredLoop:
			// Just exit the loop
			return
		}
	}
}

func getSessionFromDB(id string) (*Session, error) {
	var values map[interface{}]interface{}

	// Try to obtain the session from the database
	err := db.View(func(tx *bolt.Tx) error {
		idb := []byte(id)

		// Get the bucket
		b := tx.Bucket(bucketNameBytes)

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
			// TODO: Move this to another goroutine!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!
			// Remove the session data from the database again
			err = db.Update(func(txu *bolt.Tx) error {
				return txu.Bucket(bucketNameBytes).Delete(idb)
			})
			if err != nil {
				return err
			}

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
		id:     id,
		values: values,
	}

	return s, nil
}

func sessionIDExistsInDB(id string) (exists bool, err error) {
	exists = true

	// Check if the ID exists in the database
	err = db.View(func(tx *bolt.Tx) error {
		// Get the bucket
		b := tx.Bucket(bucketNameBytes)

		// Try to obtain the session data
		data := b.Get([]byte(id))
		if data == nil {
			exists = false
		}

		return nil
	})

	return
}

func newProtoSession(values []byte) (data []byte, err error) {
	expiresAt := time.Now().Unix() + int64(settings.Settings.SessionMaxAge)

	s := &protobuf.Session{
		Values:    values,
		ExpiresAt: &expiresAt,
	}

	data, err = proto.Marshal(s)
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
