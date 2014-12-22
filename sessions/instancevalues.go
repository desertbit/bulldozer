/*
 *  Bulldozer Framework
 *  Copyright (C) DesertBit
 */

package sessions

import (
	"code.desertbit.com/bulldozer/bulldozer/settings"
	"code.desertbit.com/bulldozer/bulldozer/utils"
	"encoding/gob"
	"github.com/golang/glog"
	"sync"
	"time"
)

func init() {
	// Register the custom types to gob
	gob.Register(&instances{})
	gob.Register(&instance{})
}

//#####################//
//### Private types ###//
//#####################//

// Key: instance ID
type instances struct {
	Set   map[string]*instance
	mutex sync.Mutex
}

type instance struct {
	// Timestamp when the last change was made
	Timestamp int64

	Values map[interface{}]interface{}
	mutex  sync.Mutex
}

func newInstances() *instances {
	return &instances{
		Set: make(map[string]*instance),
	}
}

func newInstance() *instance {
	return &instance{
		Timestamp: time.Now().Unix(),
		Values:    make(map[interface{}]interface{}),
	}
}

//########################//
//### Sessions methods ###//
//########################//

// InstanceGet obtains the instance value. This instance values
// exists as long as the browser tab session is present.
// A single variadic argument is accepted, and it is optional:
// if a function is set, this function will be called if no value
// exists for the given key.
func (s *Session) InstanceGet(key interface{}, vars ...func() interface{}) (value interface{}, ok bool) {
	// Lock the mutex
	s.sessionInstance.mutex.Lock()
	defer s.sessionInstance.mutex.Unlock()

	value, ok = s.sessionInstance.Values[key]

	// If no value is found and the create function variable
	// is set, then call the function and set the new value.
	if !ok && len(vars) > 0 {
		value = vars[0]()
		s.sessionInstance.Values[key] = value
		ok = true

		// Mark the session values as dirty.
		s.Dirty()
	}

	return
}

// InstanceSet sets the instance value with the given key.
func (s *Session) InstanceSet(key interface{}, value interface{}) {
	// Lock the mutex
	s.sessionInstance.mutex.Lock()
	defer s.sessionInstance.mutex.Unlock()

	// Set the value for the given key.
	s.sessionInstance.Values[key] = value

	// Mark the session values as dirty.
	s.Dirty()
}

// InstanceDelete removes the instance value with the given key.
func (s *Session) InstanceDelete(key interface{}) {
	// Lock the mutex
	s.sessionInstance.mutex.Lock()
	defer s.sessionInstance.mutex.Unlock()

	// Remove the value for the given key.
	delete(s.sessionInstance.Values, key)

	// Mark the session values as dirty.
	s.Dirty()
}

//###############//
//### Private ###//
//###############//

func newUniqueInstanceID(s *Session) (id string) {
	// Get the session instances.
	ii := getInstances(s)

	// Lock the mutex.
	ii.mutex.Lock()
	defer ii.mutex.Unlock()

	// Obtain a new unique instance Id
	for {
		// Get a new instance ID
		id = utils.RandomString(instanceIDLength)

		// Check if the ID is already present.
		// This is very unlikely, but we have to check this!
		_, ok := ii.Set[id]
		if !ok {
			// Break the loop if the ID is unique
			break
		}
	}

	return
}

func getInstance(s *Session) *instance {
	// Get the session instances.
	ii := getInstances(s)

	// Lock the mutex.
	ii.mutex.Lock()
	defer ii.mutex.Unlock()

	// Cleanup expired instances if there exists a bunch of instances.
	if len(ii.Set) >= maxInstancesPerSession/1.5 {
		cleanupExpiredInstances(s, ii)
	}

	// Get the current instance if present.
	i, ok := ii.Set[s.instanceID]
	if !ok {
		// If the maximum instances per sessions are reached,
		// then remove the oldest instance from the map.
		if len(ii.Set) >= maxInstancesPerSession {
			removeOldestInstance(ii)
		}

		// Create a new one and add it to the map.
		i = newInstance()
		ii.Set[s.instanceID] = i
	}

	return i
}

func getInstances(s *Session) *instances {
	// Get the session instances. Create and add it, if not present.
	i, _ := s.Get(keyInstanceValues, func() interface{} {
		return newInstances()
	})

	// Assertion
	ii, ok := i.(*instances)
	if !ok {
		// Log the error
		glog.Errorf("session instance values: failed to assert value to instances map!")

		// Just create a new one and set it to the session.
		ii = newInstances()
		s.Set(keyInstanceValues, ii)
	}

	return ii
}

// cleanupExpiredInstances has to be called with a locked mutex.
func cleanupExpiredInstances(s *Session, ii *instances) {
	isDirty := false

	// Create the release timestamp
	releaseTimestamp := time.Now().Unix() - int64(settings.Settings.SessionMaxAge)

	// Remove expired instances
	for key, i := range ii.Set {
		if i.Timestamp < releaseTimestamp {
			isDirty = true
			delete(ii.Set, key)
		}
	}

	// Save the changed map to the store
	if isDirty {
		s.storeSession.Dirty()
	}
}

// removeOldestInstance has to be called with a locked mutex.
func removeOldestInstance(ii *instances) {
	var oldestKey string
	var oldestTimestamp int64 = 0

	// Find the oldest timestamp and instance key.
	for key, i := range ii.Set {
		if i.Timestamp < oldestTimestamp || oldestTimestamp == 0 {
			oldestKey = key
			oldestTimestamp = i.Timestamp
		}
	}

	// Remove the oldest instance if found.
	if oldestKey != "" {
		delete(ii.Set, oldestKey)
	}
}
