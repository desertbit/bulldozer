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
	"time"
)

const (
	expireAccessSocketTimeout = 25 * time.Second

	sessionIDLength         = 15
	socketAccessTokenLength = 40

	// Value keys
	keyCookieToken = "bzrCookieToken"

	// Cache value keys
	cacheKeyCookieToken = "bzrCookieToken"
	cacheKeySocketType  = "bzrSocketType"
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
	socket.OnNewSocketConnection(onNewSocketConnection)
}

//####################################//
//### Socket Access Gateway Struct ###//
//####################################//

type socketAccessGateway struct {
	Token      string
	RemoteAddr string
	UserAgent  string
}

//######################//
//### Session Struct ###//
//######################//

// Hint: All session methods are thread-safe.

type Sessions map[string]*Session

type Session struct {
	sessionID string
	path      string

	socketAccess *socketAccessGateway
	emitter      *emission.Emitter
	storeSession *store.Session
	stream       *stream.Stream
	socket       socket.Socket

	stopExpireAccessSocketTimeout chan struct{}

	isClosed   bool
	closeMutex sync.Mutex
}

// SessionID returns the session ID
func (s *Session) SessionID() string {
	return s.sessionID
}

// Path returns the current URL path
func (s *Session) Path() string {
	return s.path
}

// SendCommand sends a javascript command to the client
func (s *Session) SendCommand(cmd string) {
	s.stream.Write(cmd)
}

// IsClosed returns a boolean indicating if the session socket connection is closed.
func (s *Session) IsClosed() bool {
	return s.isClosed
}

// Close closes the socket connection and removes the session
func (s *Session) Close() {
	// Call the socket close method.
	// This will trigger the removeSession function.
	s.socket.Close()
}

// SocketType returns the session's socket type
func (s *Session) SocketType() socket.SocketType {
	return s.socket.Type()
}

// RemoteAddr returns the client remote address
func (s *Session) RemoteAddr() string {
	return s.socket.RemoteAddr()
}

// Get returns the session value for the given key.
// A single variadic argument is accepted, and it is optional:
// if a function is set, this function will be called if no value
// exists for the given key.
func (s *Session) Get(key string, vars ...func() interface{}) (interface{}, bool) {
	return s.storeSession.Get(key, vars...)
}

// Set sets the value with the given key.
func (s *Session) Set(key interface{}, value interface{}) {
	s.storeSession.Set(key, value)
}

// Delete removes the value with the given key.
func (s *Session) Delete(key interface{}) {
	s.storeSession.Delete(key)
}

// Dirty sets the session values to an unsaved state,
// which will trigger the save trigger handler.
// Use this method, if you don't want to always call the
// Set() method for pointer values.
func (s *Session) Dirty() {
	s.storeSession.Dirty()
}

// CacheGet obtains the cache value.
// A single variadic argument is accepted, and it is optional:
// if a function is set, this function will be called if no value
// exists for the given key.
func (s *Session) CacheGet(key interface{}, vars ...func() interface{}) (interface{}, bool) {
	return s.storeSession.CacheGet(key, vars...)
}

// CacheSet sets the cache value with the given key.
func (s *Session) CacheSet(key interface{}, value interface{}) {
	s.storeSession.CacheSet(key, value)
}

// CacheDelete removes the cache value with the given key.
func (s *Session) CacheDelete(key interface{}) {
	s.storeSession.CacheDelete(key)
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
	// Get the store session
	var err error
	var storeSession *store.Session
	var newStoreSessionCreated bool
	for {
		storeSession, newStoreSessionCreated, err = getStoreSession(rw, req)
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

	// Hint: If any error return is added here, don't forget to unlock the store session!

	// TODO: Check if block for different socket types is working!
	// !!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!

	// Create a new session with a random socket token
	s := &Session{
		path:                          utils.ToPath(req.URL.Path),
		stream:                        stream.New(),
		storeSession:                  storeSession,
		stopExpireAccessSocketTimeout: make(chan struct{}),
		isClosed:                      false,
	}

	// Create a new emitter and set the recover function
	s.emitter = emission.NewEmitter().
		RecoverWith(recoverEmitter)

	// Get the remote address and user agent
	remoteAddr, _ := utils.RemoteAddress(req)
	userAgent := req.Header.Get("User-Agent")

	// Set the session socket to a dummy socket.
	s.socket = socket.NewSocketDummy(remoteAddr, userAgent)

	// Add a custom event function to remove the session on close.
	s.socket.OnClose(func() {
		removeSession(s)
	})

	// Create a new socket access gateway
	s.socketAccess = &socketAccessGateway{
		Token:      utils.RandomString(socketAccessTokenLength),
		RemoteAddr: remoteAddr,
		UserAgent:  userAgent,
	}

	// Lock the mutex
	sessionsMutex.Lock()
	defer sessionsMutex.Unlock()

	// Obtain a new unique session Id
	var sid string
	for {
		// Get a new session ID
		sid = utils.RandomString(sessionIDLength)

		// Check if the session Id is already used.
		// This is very unlikely, but we have to check this!
		_, ok := sessions[sid]
		if !ok {
			// Break the loop if the ID is unique
			break
		}
	}

	// Set the session ID
	s.sessionID = sid

	// Add the session to the map
	sessions[sid] = s

	// Trigger the new session hook after the mutex is unlocked again
	defer triggerOnNewSession(s)

	// Remove the session if no socket connected to this session
	// after a specific timeout.
	go func() {
		// Create a new timer
		timer := time.NewTimer(expireAccessSocketTimeout)

		defer func() {
			// Stop the timer
			timer.Stop()
		}()

		select {
		case <-timer.C:
			// Remove the session
			removeSession(s)

			// Delete the store session from the store, if it is a newly created one.
			if newStoreSessionCreated {
				store.Remove(storeSession.ID())
			}
		case <-s.stopExpireAccessSocketTimeout:
			// Just exit the loop
			return
		}
	}()

	// Return the new created session
	return s, s.socketAccess.Token, nil
}

// GetSession returns a session with the given session ID.
// ok is false, if the session was not found.
func GetSession(sessionID string) (s *Session, ok bool) {
	// Lock the mutex
	sessionsMutex.Lock()
	defer sessionsMutex.Unlock()

	s, ok = sessions[sessionID]
	return
}

// GetSessions calls the passed function with all current active session.
// This is done with a function call, because a mutex has to be locked to access the sessions map.
func GetSessions(f func(sessions Sessions)) {
	// Lock the mutex
	sessionsMutex.Lock()
	defer sessionsMutex.Unlock()

	f(sessions)
}

//###############//
//### Private ###//
//###############//

// removeSession removes the session from the active session map
func removeSession(s *Session) {
	// Check if already closed.
	// This method should be called only once for each session.
	s.closeMutex.Lock()
	if s.isClosed {
		s.closeMutex.Unlock()
		return
	}
	s.isClosed = true
	s.closeMutex.Unlock()

	// Trigger the end session events
	triggerOnCloseSession(s)
	s.triggerOnClose()

	// Remove the lock for this store session
	s.storeSession.Unlock()

	// Lock the mutex
	sessionsMutex.Lock()
	defer sessionsMutex.Unlock()

	// Remove the session from the map
	delete(sessions, s.sessionID)
}
