/*
 *  Bulldozer Framework
 *  Copyright (C) DesertBit
 */

package sessions

import (
	"code.desertbit.com/bulldozer/bulldozer/sessions/store"
	"code.desertbit.com/bulldozer/bulldozer/settings"
	"code.desertbit.com/bulldozer/bulldozer/utils"
	"encoding/gob"
	"github.com/golang/glog"
	"net/http"
	"time"
)

const (
	cookieName        = "id"
	cookieTokenLength = 15

	cookieValueTimeout = 20 * time.Second
)

func init() {
	// Register the custom struct to gob
	gob.Register(&sessionCookie{})
}

//#######################//
//### Private Structs ###//
//#######################//

type sessionCookie struct {
	ID    string
	Token string
}

type cookieValue struct {
	LastToken    string
	SetNewCookie bool
}

//###############//
//### Private ###//
//###############//

// getStoreSession returns the store session fitting to the cookie or creates a new one.
// A boolean is returned indicating if a new session was created.
func getStoreSession(rw http.ResponseWriter, req *http.Request) (*store.Session, bool, error) {
	var newStoreSessionCreated bool = false
	var sCookie sessionCookie
	var storeSession *store.Session

	// Try to obtain the bulldozer session cookie.
	// If no cookie is found, then the new session is created automatically.
	cookie, err := req.Cookie(cookieName)
	if err == nil {
		err = secureCookie.Decode(cookieName, cookie.Value, &sCookie)
		if err != nil {
			// This is not a fatal error. Just log it and create a new session.
			// The new session is created automatically, if cookie session ID is emtpy or invalid.
			glog.Errorf("failed to decode session cookie: %v", err)
		}
	} else if err != http.ErrNoCookie {
		// Return the error if this is not the not found cookie error
		return nil, false, err
	}

	// Try to obtain the store session with the cookie session ID
	if len(sCookie.ID) > 0 {
		storeSession, err = store.Get(sCookie.ID)
		if err != nil && err != store.ErrNotFound {
			return nil, false, err
		}

		if storeSession != nil {
			// Check if the cookie token is valid
			cookieTokenI, ok := storeSession.Get(keyCookieToken)
			if !ok {
				// No cookie token session value found.
				// Remove the store session, because it is invalid.
				store.Remove(storeSession.ID())

				// Reset the storeSession pointer to nil, so a new session is createdd.
				storeSession = nil
			} else {
				cookieToken, ok := cookieTokenI.(string)
				if !ok || cookieToken != sCookie.Token {
					// Obtain the cached cookie value from the session
					cValue := getCachedCookieValue(storeSession)

					// Check if the cookie token was the previous token
					if cValue.LastToken == "" || cValue.LastToken != sCookie.Token {
						// Log
						addr, _ := utils.RemoteAddress(req)
						glog.Warningf("invalid client session cookie token '%s' from client: %s", sCookie.Token, addr)

						// Reset the storeSession pointer to nil, so a new session is createdd.
						storeSession = nil
					}
				}
			}
		}
	}

	// Create a new session if no store session was found
	if storeSession == nil {
		// Create a new session
		storeSession, err = store.New()
		if err != nil {
			return nil, false, err
		}

		// Set the flag
		newStoreSessionCreated = true

		// Set the new store session ID to the session cookie
		sCookie.ID = storeSession.ID()
	}

	// Obtain the cached cookie value from the session
	cValue := getCachedCookieValue(storeSession)

	// If the flag is set, then create a new cookie with a new token.
	if cValue.SetNewCookie {
		// Save the last token. This way, parallel accesses to obtain the session won't fail
		// because they still have the old token. This old token is valid for a short timeout.
		cValue.LastToken = sCookie.Token
		cValue.SetNewCookie = false

		// Start a new goroutine to reset the last token string and the flag
		go func() {
			// Sleep
			time.Sleep(cookieValueTimeout)

			// Reset the values
			cValue.LastToken = ""
			cValue.SetNewCookie = true
		}()

		// Add the cookie value to the cached session values
		storeSession.CacheSet(cacheKeyCookieToken, cValue)

		// Set a new random cookie token. This is a security improvement.
		sCookie.Token = utils.RandomString(cookieTokenLength)

		// Encode the session cookie
		encoded, err := secureCookie.Encode(cookieName, sCookie)
		if err != nil {
			// Return the encoding error
			return nil, false, err
		}

		// TODO: Set cookie max age to settings.Settings.SessionMaxAge if authenticated and if remeber login is set

		// Create a new session cookie
		cookie = &http.Cookie{
			Name:     cookieName,
			Value:    encoded,
			Path:     "/",
			MaxAge:   0,
			HttpOnly: true,                                // Don't allow scripts to manipulate the cookie
			Secure:   settings.Settings.SecureHttpsAccess, // Only send this cookie over a secure https connection if provided
		}

		// Set the new session cookie
		http.SetCookie(rw, cookie)

		// Set the cookie token to the session store
		storeSession.Set(keyCookieToken, sCookie.Token)
	}

	return storeSession, newStoreSessionCreated, nil
}

func getCachedCookieValue(storeSession *store.Session) (value *cookieValue) {
	// Obtain the cached cookie value from the session
	i, ok := storeSession.CacheGet(cacheKeyCookieToken)
	if ok {
		value, ok = i.(*cookieValue)
	}
	if !ok {
		// If not found, then create a new one
		value = &cookieValue{
			SetNewCookie: true,
		}
	}

	return
}
