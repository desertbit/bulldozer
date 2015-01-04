/*
 *  Bulldozer Framework
 *  Copyright (C) DesertBit
 */

package sessions

import (
	"code.desertbit.com/bulldozer/bulldozer/sessions/socket"
	"code.desertbit.com/bulldozer/bulldozer/sessions/stream"
	"code.desertbit.com/bulldozer/bulldozer/log"
	"time"
)

const (
	socketKeyPing = "ping"
	socketKeyPong = "pong"

	socketKeySessionID = "sid"
	socketKeyToken     = "tok"
	socketKeyTask      = "tsk"

	socketValueInvalidRequest = "invalid_request"

	initTokenDelimiter = "&"

	// Send pings to peer with this period
	pingPeriod = 30 * time.Second
)

//#############################//
//### Socket Session Struct ###//
//#############################//

type socketSession struct {
	socketConn socket.Socket
	session    *Session
	token      *randomToken
	stream     *stream.Stream

	pingCount int
	pingTimer *time.Timer

	stopWriteLoop chan struct{}
}

//###############//
//### Private ###//
//###############//

func onNewSocketConnection(s socket.Socket) {
	// Create a new socket session
	ss := &socketSession{
		socketConn: s,
		session:    nil,
		token:      newRandomToken(),
		stream:     stream.New(),

		pingCount: 0,
		pingTimer: time.NewTimer(pingPeriod),

		stopWriteLoop: make(chan struct{}),
	}

	// Set the socket event functions
	s.OnClose(ss.onClose)
	s.OnRead(ss.onRead)

	// Start the goroutine for writing messages to the client
	go ss.writeLoop()
}

func (ss *socketSession) resetPingTimer() {
	// Reset the timer again
	ss.pingTimer.Reset(pingPeriod)

	// Reset the ping count
	ss.pingCount = 0
}

// writeLoop writes messages from the stream to the socket connection
func (ss *socketSession) writeLoop() {
	defer func() {
		// Stop the timer
		ss.pingTimer.Stop()
	}()

	for {
		select {
		case <-ss.stream.HasData:
			// Create a new token
			t, ok := ss.token.new()
			if !ok {
				log.L.Warning("Closing session %s with remote address '%s' due to flooding attack!", ss.session.SessionID(), ss.socketConn.RemoteAddr())
				// Immediately close the session. The client tries to flood the server...
				ss.socketConn.Close()
				return
			}

			// Send the new token and the message
			ss.socketConn.Write(t + "&" + ss.stream.Read())
		case <-ss.pingTimer.C:
			// Check if the client didn't respond since the last ping request.
			if ss.pingCount >= 1 {
				// Close the socket
				ss.socketConn.Close()
				return
			}

			// Increment the ping count
			ss.pingCount += 1

			// Create a new token
			t, ok := ss.token.new()
			if !ok {
				log.L.Warning("Closing session %s with remote address '%s' due to flooding attack!", ss.session.SessionID(), ss.socketConn.RemoteAddr())
				// Immediately close the session. The client tries to flood the server...
				ss.socketConn.Close()
				return
			}

			// Send the new token and a ping request
			ss.socketConn.Write(t + "&" + socketKeyPing)

			// Reset the timer again
			ss.pingTimer.Reset(pingPeriod)
		case <-ss.stopWriteLoop:
			// Just exit the loop
			return
		}
	}
}

// Send an invalid request to the client and close the socket connection
func (ss *socketSession) receivedInvalidRequest(closeConn bool) {
	ss.socketConn.Write(socketValueInvalidRequest)

	if closeConn {
		ss.socketConn.Close()
	}
}

func (ss *socketSession) onClose() {
	// Stop the write messages loop by triggering the quit trigger
	close(ss.stopWriteLoop)

	// Remove the session if defined
	if ss.session != nil {
		removeSession(ss.session)
	}
}

func (ss *socketSession) onRead(data string) {
	// Create a data map from the received message
	m := getDataMap(data)

	// Try to obtain the session Id
	sid, ok := m[socketKeySessionID]
	if !ok {
		log.L.Warning("received an invalid session ID from the client")
		ss.receivedInvalidRequest(true)
		return
	}

	// Try to obtain the temporary token
	token, ok := m[socketKeyToken]
	if !ok {
		log.L.Warning("missing temporary token in client request")
		ss.receivedInvalidRequest(true)
		return
	}

	// If no session is set, then try to initialize it
	if ss.session == nil {
		ss.initSocketSession(m, sid, token)
		return
	}

	// Check if the session matches and if the token is valid
	if ss.session.sessionID != sid || !ss.token.isTokenValid(token) {
		log.L.Warning("socket session: the session ID or session token is invalid!")
		ss.receivedInvalidRequest(true)
		return
	}

	// Reset the ping timer
	ss.resetPingTimer()

	// Try to obtain the task
	task, ok := m[socketKeyTask]
	if !ok {
		log.L.Warning("missing task in client request")
		ss.receivedInvalidRequest(true)
		return
	}

	// If this is a pong answer, then just return.
	// The timeout timer is already reset
	if task == socketKeyPong {
		return
	}

	// Get the request with the task string as type
	request, ok := requests[task]
	if !ok {
		log.L.Warning("session request for task type '%s' not found!", task)
		ss.receivedInvalidRequest(false)
		return
	}

	// Call the request function
	err := request(ss.session, m)
	if err != nil {
		log.L.Warning("session request '%s': error: %v", task, err)
		ss.receivedInvalidRequest(false)
		return
	}
}

func (ss *socketSession) initSocketSession(m map[string]string, sid string, accessToken string) {
	if sid == "" || accessToken == "" {
		log.L.Warning("invalid session ID '%s' or access token '%s' in client request!", sid, accessToken)
		ss.receivedInvalidRequest(true)
		return
	}

	// Try to get the session with the session ID
	s, ok := GetSession(sid)
	if !ok {
		log.L.Warning("invalid session ID in client request: session with ID '%s' not found!", sid)
		ss.receivedInvalidRequest(true)
		return
	}

	// Check if the socket access token, remote address and user agent is valid
	if s.socketAccess == nil ||
		s.socketAccess.Token != accessToken ||
		s.socketAccess.RemoteAddr != ss.socketConn.RemoteAddr() ||
		s.socketAccess.UserAgent != ss.socketConn.UserAgent() {
		log.L.Warning("invalid socket access: token, remote address or user agent don't match!")
		ss.receivedInvalidRequest(true)
		return
	}

	// Invalidate the socket access gateway of this session.
	// A connection was successfully established.
	s.socketAccess = nil

	// Check if the new socket connection has the same socket type
	// than other active socket connections in the same store session.
	socketTypeI, _ := s.storeSession.CacheGet(cacheKeySocketType, func() interface{} {
		return ss.socketConn.Type()
	})
	socketType, ok := socketTypeI.(socket.SocketType)
	if !ok || socketType != ss.socketConn.Type() {
		log.L.Error("session socket connected with a different socket type than the other active socket sessions: remote address: %s", ss.socketConn.RemoteAddr())
		ss.receivedInvalidRequest(true)
		return
	}

	// Set the session pointer
	ss.session = s

	// Stop the expire timeout
	close(s.stopExpireAccessSocketTimeout)

	// Set the new socket stream to the session
	pStream := s.stream
	s.stream = ss.stream

	// Set the socket to the session
	s.socket = ss.socketConn

	// Tell the client the instance ID and the new token
	ss.socketConn.Write(s.instanceID + initTokenDelimiter + ss.token.get())

	// Write all previous buffered stream data to the new stream
	data := pStream.Read()
	if len(data) > 0 {
		s.stream.Write(data)
	}
}

// getDataMap creates a data map out of a string.
// The input string format should be: "key1=data1&key2=data2&"
// Escaped '\\' and '\&' are replaced with '\' and '&'
func getDataMap(s string) map[string]string {
	m := make(map[string]string)
	var key, data []rune
	var pp rune
	isKey := true

	for _, p := range s {
		if isKey {
			if p == '=' {
				isKey = false
			} else {
				key = append(key, p)
			}
		} else {
			if p == '\\' && pp == '\\' {
				// Do nothing here, because '\\' should be tranformed in '\'.
				// Change the p rune to any character, so that pp won't hold '\'.
				// This is important for the '&'...
				p = ' '
			} else if p == '&' {
				// Skip escaped '&' characters
				if pp == '\\' {
					// Remove the last '\' and replace it by the '&' character
					data[len(data)-1] = '&'
				} else {
					// Return an emtpy map if the key is empty
					if len(key) == 0 {
						return make(map[string]string)
					}

					m[string(key)] = string(data)
					key = key[:0]
					data = data[:0]
					isKey = true
				}
			} else {
				data = append(data, p)
			}
		}

		// Save the current part to the previous part rune
		pp = p
	}

	return m
}
