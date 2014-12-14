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
	"github.com/gorilla/securecookie"
	"net/http"
	"sync"
)

const (
	// Value keys
	keyCookieToken = "bzrCookieToken"

	// Cache value keys
	cacheKeyCookieToken = "bzrCookieToken"
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

	storeSession *store.Session
}

// SessionID returns the session ID
func (s *Session) SessionID() string {
	return s.sessionId
}

// SendCommand sends a javascript command to the client
func (s *Session) SendCommand(cmd string) {
	s.stream.Write(cmd)
}

// Value returns the session value for the given key
func (s *Session) Value(key string) {
	// Don't allow to access some important private session values
	if key == keyCookieToken {
		// return nil
	}

}

// TODO: add cachevalues....

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
	// Get the store session
	var err error
	var storeSession *store.Session
	for {
		storeSession, err = getStoreSession(rw, req)
		if err != nil {
			return nil, "", err
		}

		// Add a lock for this new session
		if !storeSession.Lock() {
			// If this fails, then the current storeSession pointer has been
			// released from memory, by another parallel Unlock request.
			// This might never happen, but we have to handle this, by just
			// making another call to store.Get...
			continue
		}

		break
	}

	// TODO: Onclose unlock store session

	// TODO: Flash messages

	// TODO: Rename this!!!
	var sid string
	socketToken := "TODO"

	// TODO: Delete this session if no socket connection is connected!!!!!!!!!!!!!! after timeout
	// ANd also delete the store session from the db

	// Create a new session
	s := &Session{
		stream:       stream.New(),
		storeSession: storeSession,
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
