/*
 *  Bulldozer Framework
 *  Copyright (C) DesertBit
 */

package sessions

import (
	"code.desertbit.com/bulldozer/bulldozer/utils"
	"container/list"
	"sync"
)

const (
	randomTokenLength = 8
)

// TODO: Implement the diffie hellman algorithm to generate a secret password and
// use this to generate a checksum with the password and the send message.

// This is a random token struct to generete and hold a random string.
// A new token generated with the new() method is used as soon as the client sends messages
// with the new token. All old tokens are immediately removed.
type randomToken struct {
	tokens *list.List
	mutex  sync.Mutex
}

func newRandomToken() (t *randomToken) {
	// Create a new token
	t = &randomToken{tokens: list.New()}

	// Create an initial random token
	t.new()

	return
}

func (t *randomToken) get() string {
	// Lock the mutex
	t.mutex.Lock()
	defer t.mutex.Unlock()

	// Return the current token value
	e := t.tokens.Front()

	if e != nil {
		if str, ok := e.Value.(string); ok {
			return str
		}
	}

	return ""
}

func (t *randomToken) new() (newToken string, ok bool) {
	// Lock the mutex
	t.mutex.Lock()
	defer t.mutex.Unlock()

	// Check if the client tries to flood the session, by sending always the old token...
	if t.tokens.Len() > 50 {
		ok = false
		return
	}

	// Create a new token
	newToken = utils.RandomString(randomTokenLength)

	// Add the new token to the list
	t.tokens.PushBack(newToken)

	ok = true
	return
}

func (t *randomToken) isTokenValid(token string) bool {
	// Input check
	if len(token) <= 0 {
		return false
	}

	// Lock the mutex
	t.mutex.Lock()
	defer t.mutex.Unlock()

	// Go through all available tokens and check if there is a match
	for e := t.tokens.Front(); e != nil; e = e.Next() {
		// Try to get the token string value
		str, ok := e.Value.(string)
		if !ok {
			return false
		}

		// Check if the tokens match
		if token == str {
			// Remove all tokens in front of the current found token
			var ne *list.Element
			for re := t.tokens.Front(); re != e; re = ne {
				ne = re.Next()
				t.tokens.Remove(re)
			}

			return true
		}
	}

	// The passed token is invalid
	return false
}
