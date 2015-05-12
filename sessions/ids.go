/*
 *  Bulldozer Framework
 *  Copyright (C) DesertBit
 */

package sessions

import (
	"encoding/gob"
	"github.com/desertbit/bulldozer/utils"
	"strconv"
	"sync"
)

func init() {
	// Register the custom type to gob
	gob.Register(&uniqueID{})
}

//#############//
//### Types ###//
//#############//

type uniqueID struct {
	IDCount uint64
	mutex   sync.Mutex
}

func (u *uniqueID) NewID() uint64 {
	// Lock the mutex.
	u.mutex.Lock()
	defer u.mutex.Unlock()

	// Increment the unique count.
	u.IDCount++

	return u.IDCount
}

//#######################//
//### Session Methods ###//
//#######################//

// NewUniqueDomID returns a new unique DOM ID.
func (s *Session) NewUniqueDomID() string {
	// Get the unique ID type from the session store.
	uid := s.getUniqueIDType(keyUniqueDomID)

	// Get the new uid as string.
	idStr := "uid_" + strconv.FormatUint(uid.NewID(), 10)

	// Calculate the unique DOM ID with the session key.
	return utils.EncryptDomId(s.domEncryptionKey, idStr)
}

// NewUniqueId returns a new unique ID.
// This ID can be used for unique template IDs...
func (s *Session) NewUniqueId() string {
	// Get the unique ID type from the session store.
	uid := s.getUniqueIDType(keyUniqueID)

	// Return the new uid as string.
	return strconv.FormatUint(uid.NewID(), 10)
}

//###############//
//### Private ###//
//###############//

func (s *Session) getUniqueIDType(key string) *uniqueID {
	// Get or create the unique ID.
	i, _ := s.Get(key, func() interface{} {
		return &uniqueID{
			IDCount: 0,
		}
	})

	// Assertion.
	uid, ok := i.(*uniqueID)
	if !ok {
		uid = &uniqueID{
			IDCount: 0,
		}

		s.Set(key, uid)
	}

	return uid
}
