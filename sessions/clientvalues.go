/*
 *  Bulldozer Framework
 *  Copyright (C) DesertBit
 */

package sessions

import (
	"code.desertbit.com/bulldozer/bulldozer/log"
	"code.desertbit.com/bulldozer/bulldozer/utils"
	"fmt"
	"time"
)

const (
	cacheValueKeyCallbackPrefix = "budClientCB_"

	requestTypeClientData   = "clientData"
	requestKeyDataKey       = "key"
	requestKeyDataRandomKey = "rand"
	requestKeyData          = "data"
)

func init() {
	// Register the route server request.
	err := Request(requestTypeClientData, onClientDataResponse)
	if err != nil {
		log.L.Fatalf("failed to register session client data response request: %v", err)
	}
}

//#############################//
//### Session Client Values ###//
//#############################//

type ClientDataCallback func(data string)
type ClientDataErrorCallback func(err error)

// ClientGet gets a value from the client side value store.
func (s *Session) ClientGet(key string, cb ClientDataCallback, errCb ...ClientDataErrorCallback) {
	// Create a random key.
	randomKey := utils.RandomString(15)

	// Add the prefix to the key.
	cacheKey := cacheValueKeyCallbackPrefix + key + randomKey

	// Set the callback to the cache.
	s.CacheSet(cacheKey, cb)

	// Send the request to the client.
	s.SendCommand(`Bulldozer.data.getAndReply('` + utils.EscapeJS(key) + `','` + utils.EscapeJS(randomKey) + `');`)

	// Start the timeout go-routine.
	go func() {
		// Wait.
		time.Sleep(10 * time.Second)

		// Obtain the value.
		_, ok := s.CachePull(cacheKey)
		if !ok {
			return
		}

		if len(errCb) > 0 {
			// If the value still exists, then the callback was not called.
			// Call the error callback.
			errCb[0](fmt.Errorf("failed to get client data: client did not respond!"))
		}
	}()
}

// ClientSet sets a value to the client side value store.
func (s *Session) ClientSet(key string, data string) {
	// Send the request to the client.
	s.SendCommand(`Bulldozer.data.set('` + utils.EscapeJS(key) + `','` + utils.EscapeJS(data) + `');`)

}

// ClientSet removes a value from the client side value store.
func (s *Session) ClientDelete(key string) {
	// Send the request to the client.
	s.SendCommand(`Bulldozer.data.delete('` + utils.EscapeJS(key) + `');`)

}

//###############//
//### Private ###//
//###############//

func onClientDataResponse(s *Session, data map[string]string) error {
	// Try to obtain the data key.
	key, ok := data[requestKeyDataKey]
	if !ok {
		return fmt.Errorf("client data: missing data key in request!")
	}

	// Try to obtain the random key.
	randomKey, ok := data[requestKeyDataRandomKey]
	if !ok {
		return fmt.Errorf("client data: missing random key in request!")
	}

	// Try to obtain the data value.
	d, ok := data[requestKeyData]
	if !ok {
		return fmt.Errorf("client data: missing data value in request!")
	}

	// Add the prefix to the key.
	key = cacheValueKeyCallbackPrefix + key + randomKey

	// Try to get the callback.
	i, ok := s.CachePull(key)
	if !ok {
		return fmt.Errorf("client data: no callback found for key '%s'", key)
	}

	// Assertion.
	cb, ok := i.(ClientDataCallback)
	if !ok {
		return fmt.Errorf("client data: failed to assert callback with key '%s'", key)
	}

	// Call the callback.
	cb(d)

	return nil
}
