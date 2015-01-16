/*
 *  Bulldozer Framework
 *  Copyright (C) DesertBit
 */

package auth

import (
	"fmt"
	"sync"
	"time"
)

const (
	expireCacheLoopTimeout      = 5 * time.Minute
	removeFromCacheAfterTimeout = 60 * 5 // 5 minutes.
)

var (
	stopExpireCacheLoop chan struct{} = make(chan struct{})

	chachedDBUsers      map[string]*cachedUser = make(map[string]*cachedUser)
	chachedDBUsersMutex sync.Mutex
)

//#####################//
//### Private Types ###//
//#####################//

type cachedUser struct {
	expiresAt int64
	user      *dbUser
}

func newCachedUser(u *dbUser) *cachedUser {
	c := &cachedUser{
		user: u,
	}

	c.ResetExpireTime()

	return c
}

func (c *cachedUser) ResetExpireTime() {
	c.expiresAt = time.Now().Unix() + removeFromCacheAfterTimeout
}

//#######################//
//### Private Methods ###//
//#######################//

// cacheGetDBUser gets the user from the cache or if not found
// loads it from the database and adds it to the cache.
// If the user does not exists, then nil is returned.
func cacheGetDBUser(id string) (*dbUser, error) {
	// Lock the mutex.
	chachedDBUsersMutex.Lock()
	defer chachedDBUsersMutex.Unlock()

	// Check if the value is in the map.
	c, ok := chachedDBUsers[id]
	if ok {
		// Reset the expire time in a new goroutine.
		// We don't have to wait for this...
		go c.ResetExpireTime()

		return c.user, nil
	}

	// Obtain the user from the database.
	u, err := dbGetUserByID(id)
	if err != nil {
		return nil, fmt.Errorf("failed to get session user by ID: %v", err)
	} else if u == nil {
		return nil, nil
	}

	// Add the user to the cache.
	chachedDBUsers[u.ID] = newCachedUser(u)

	return u, nil
}

// cacheUserOutOfDate removes the user from the cache.
// This will force to retrieve the user from the database.
func cacheUserOutOfDate(id string) {
	// Lock the mutex.
	chachedDBUsersMutex.Lock()
	defer chachedDBUsersMutex.Unlock()

	delete(chachedDBUsers, id)
}

//##############//
//### Expire ###//
//##############//

func startExpireCacheLoop() {
	// Start the expire loop in a new goroutine.
	go expireCacheLoop()
}

func releaseCache() {
	// Stop the loop by triggering the quit trigger.
	close(stopExpireCacheLoop)
}

func expireCacheLoop() {
	// Create a new ticker
	ticker := time.NewTicker(expireCacheLoopTimeout)

	defer func() {
		// Stop the ticker
		ticker.Stop()
	}()

	for {
		select {
		case <-ticker.C:
			// Cleanup expired users from the cache.
			removeExpiredCacheUsers()
		case <-stopExpireCacheLoop:
			// Just exit the loop
			return
		}
	}
}

func removeExpiredCacheUsers() {
	// Get the current timestamp.
	timeNow := time.Now().Unix()

	// Lock the mutex.
	chachedDBUsersMutex.Lock()
	defer chachedDBUsersMutex.Unlock()

	for id, u := range chachedDBUsers {
		if u.expiresAt < timeNow {
			delete(chachedDBUsers, id)
		}
	}
}
