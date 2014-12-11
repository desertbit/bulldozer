/*
 *  Bulldozer Framework
 *  Copyright (C) DesertBit
 */

package sessions

import (
	"code.desertbit.com/bulldozer/bulldozer/sessions/socket"
	"code.desertbit.com/bulldozer/bulldozer/sessions/store"
	"code.desertbit.com/bulldozer/bulldozer/sessions/stream"
	"code.desertbit.com/bulldozer/bulldozer/settings"
	"code.desertbit.com/bulldozer/bulldozer/utils"
	"github.com/chuckpreslar/emission"
	"github.com/golang/glog"
	"github.com/gorilla/securecookie"
	"net/http"
	"sync"
)

const (
	cookieName = "id"
)

var (
	// We don't use a RWMutex here, because the map access is fast enough and
	// a RWMutex would create more overhead.
	sessionsMutex sync.Mutex
	sessions      Sessions = make(Sessions)

	// The secure cookie
	secureCookie *securecookie.SecureCookie
)

func init() {
	// Initialize the socket http handlers
	socket.InitHttpHandlers()

	// Set the function
	//socket.OnNewSocketConnection(onNewSocketConnection)
}

//######################//
//### Session Struct ###//
//######################//

type Sessions map[string]*Session

type Session struct {
	sessionId string

	stream  *stream.Stream
	emitter *emission.Emitter
}

// SessionID returns the session ID
func (s *Session) SessionID() string {
	return s.sessionId
}

// SendCommand sends a javascript command to the client
func (s *Session) SendCommand(cmd string) {
	s.stream.Write(cmd)
}

//##############//
//### Public ###//
//##############//

// Init initializes the sessions packages.
// This is called and handled by default by the bulldozer main package.
func Init() {
	// Create a new secure cookie object with the cookie keys
	secureCookie = securecookie.New(settings.Settings.CookieHashKey, settings.Settings.CookieBlockKey)

	// Set the max age in seconds
	secureCookie.MaxAge(settings.Settings.SessionMaxAge)

	// Initialize the store package
	store.Init()
}

// Release releases this session package.
// This is handled by the main bulldozer package.
func Release() {
	// Release the store package
	store.Release()
}

// New creates and registers a new session, by adding it to the
// active session map. The session cookie is extracted from the request
// and the the new session is assigned to the server session.
// If no cookie is set, a new one will be assigned.
// A unique socket access token is returned.
// Use this token to connect to the session socket.
func New(rw http.ResponseWriter, req *http.Request) (*Session, string, error) {
	var sessionId string

	// Try to obtain the bulldozer session cookie.
	// If no cookie is found, then the new session is created automatically,
	// if sessionId is emtpy or invalid.
	cookie, err := req.Cookie(cookieName)
	if err == nil {
		err = secureCookie.Decode(cookieName, cookie.Value, &sessionId)
		if err != nil {
			// This is not a fatal error. Just log it, reset sessionId and create a new session.
			// The new session is created automatically, if sessionId is emtpy or invalid.
			glog.Errorf("failed to decode session cookie: %v", err)
			sessionId = ""
		}
	} else if err != http.ErrNoCookie {
		// Return the error if this is not the not found cookie error
		return nil, "", err
	}

	// TODO: Should we renew the cookie max age as in store?

	// TODO: Check if invalid!!!
	// Check if no session ID was found
	if len(sessionId) == 0 {
		// TODO: create new one
		_, sessionId, err = store.New()

		// Encode the session ID
		encoded, err := secureCookie.Encode(cookieName, sessionId)
		if err != nil {
			// Return the encoding error
			return nil, "", err
		}

		// TODO: Set cookie max age

		// Create a new session cookie
		cookie = &http.Cookie{
			Name:     cookieName,
			Value:    encoded,
			Path:     "/",
			MaxAge:   0,                                   //settings.Settings.SessionMaxAge,
			HttpOnly: true,                                // Don't allow scripts to manipulate the cookie
			Secure:   settings.Settings.SecureHttpsAccess, // Only send this cookie over a secure https connection if provided
		}

		// Set the session cookie
		http.SetCookie(rw, cookie)
	}

	// TODO: Rename this!!!
	var sid string
	socketToken := "TODO"

	// TODO: Delete this session if no socket connection is connected!!!!!!!!!!!!!! after timeout
	// ANd also delete the store session from the db

	// Create a new session
	s := &Session{
		stream: stream.New(),
	}

	// Create a new emitter and set the recover function
	s.emitter = emission.NewEmitter().
		RecoverWith(recoverEmitter)

	// Lock the mutex
	sessionsMutex.Lock()
	defer sessionsMutex.Unlock()

	// Obtain a new unique session Id
	for {
		// Get a new session ID
		sid = utils.RandomString(15)

		// Check if the session Id is already used.
		// This is very unlikely, but we have to check this!
		_, ok := sessions[sid]
		if !ok {
			// Break the loop if the ID is unique
			break
		}
	}

	// Set the session ID
	s.sessionId = sid

	// Add the session to the map
	sessions[sid] = s

	// Trigger the new session hook after the mutex is unlocked again
	defer triggerOnNewSession(s)

	// Return the new created session
	return s, socketToken, nil
}

//###############//
//### Private ###//
//###############//

// removeSession removes the session from the active session map
func removeSession(s *Session) {
	// Trigger the end session events
	triggerOnCloseSession(s)
	s.triggerOnClose()

	// Lock the mutex
	sessionsMutex.Lock()
	defer sessionsMutex.Unlock()

	// Remove the session from the map
	delete(sessions, s.sessionId)
}
