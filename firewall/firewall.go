/*
 *  Bulldozer Framework
 *  Copyright (C) DesertBit
 */

package firewall

import (
	"code.desertbit.com/bulldozer/bulldozer/settings"
	"code.desertbit.com/bulldozer/bulldozer/utils"
	"net/http"
	"sync"
	"time"
)

var (
	blockList      map[string]int64 = make(map[string]int64)
	blockListMutex sync.Mutex

	requestCounter      map[string]int = make(map[string]int)
	requestCounterMutex sync.Mutex
)

func init() {
	// Start the loops in an new goroutine
	go cleanupRequestsCountLoop()
	go releaseBlockedRemoteAddrLoop()
}

//##############//
//### Public ###//
//##############//

// NewRequest tells the firewall, that a new request happened.
// False is returned, if this request should be blocked,
// because the IP is on the block list.
// The remote address is always returned for loggin purpose.
func NewRequest(req *http.Request) (bool, string) {
	// Get the remote addresse of the request
	remoteAddr, _ := utils.RemoteAddress(req)

	// Check if this remote address is blocked
	if isBlocked(remoteAddr) {
		return false, remoteAddr
	}

	// Register the new request in a new goroutine
	go addRequest(remoteAddr)

	return true, remoteAddr
}

//###############//
//### Private ###//
//###############//

// isBlocked checks if the remote addres is blocked
func isBlocked(remoteAddr string) bool {
	// Lock the mutex
	blockListMutex.Lock()
	defer blockListMutex.Unlock()

	// Check if the remote addresse exists in the blocked map
	_, exists := blockList[remoteAddr]
	return exists
}

func addRequest(remoteAddr string) {
	// Lock the mutex
	requestCounterMutex.Lock()
	defer requestCounterMutex.Unlock()

	count, ok := requestCounter[remoteAddr]
	if !ok {
		count = 1
	}

	// Add the remote address to the block map, if the count
	// reached the limit.
	if count > settings.Settings.FirewallMaxRequestsPerMinute {
		// Remove the remote address from the request counter map
		delete(requestCounter, remoteAddr)

		// Get the current timestamp
		timestamp := time.Now().Unix()

		// Lock the mutex
		blockListMutex.Lock()
		defer blockListMutex.Unlock()

		// Add the remote address with a timestamp to the block map
		blockList[remoteAddr] = timestamp

		return
	}

	// Save the incremented count to the map
	requestCounter[remoteAddr] = count + 1
}

func cleanupRequestsCountLoop() {
	for {
		// Sleep one minute
		time.Sleep(time.Minute)

		func() {
			// Lock the mutex
			requestCounterMutex.Lock()
			defer requestCounterMutex.Unlock()

			// Clear the map if not emtpy
			if len(requestCounter) > 0 {
				requestCounter = make(map[string]int)
			}
		}()
	}
}

func releaseBlockedRemoteAddrLoop() {
	for {
		// Sleep one minute
		time.Sleep(time.Minute)

		func() {
			// Lock the mutex
			blockListMutex.Lock()
			defer blockListMutex.Unlock()

			// Return if nothing to do
			if len(blockList) == 0 {
				return
			}

			// Create the release timestamp
			releaseTimestamp := time.Now().Unix() - int64(settings.Settings.FirewallReleaseBlockAfter)

			// Remove expired blocks
			for key, timestamp := range blockList {
				if timestamp < releaseTimestamp {
					delete(blockList, key)
				}
			}
		}()
	}
}
