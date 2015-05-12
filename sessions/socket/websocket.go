/*
 *  Bulldozer Framework
 *  Copyright (C) DesertBit
 */

package socket

import (
	"github.com/desertbit/bulldozer/log"
	"github.com/desertbit/bulldozer/utils"
	"github.com/gorilla/websocket"
	"io"
	"net/http"
	"sync"
	"time"
)

const (
	// Time allowed to write a message to the peer.
	writeWait = 10 * time.Second

	// Time allowed to read the next message from the peer.
	readWait = 60 * time.Second

	// Maximum message size allowed from peer.
	maxMessageSize = 0
)

var (
	// Websocket upgrader
	upgrader = websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
	}
)

//#######################//
//### WebSocket Layer ###//
//#######################//

type WebSocket struct {
	ws *websocket.Conn

	isClosed bool
	mutex    sync.Mutex

	onClose func()
	onRead  func(string)

	userAgent string

	remoteAddrFunc func() string
}

func NewWebSocket() *WebSocket {
	// Create a new websocket struct
	return &WebSocket{
		ws:       nil,
		onRead:   nil,
		isClosed: false,
		onClose:  nil,
	}
}

func (w *WebSocket) Type() SocketType {
	return TypeWebSocket
}

func (w *WebSocket) RemoteAddr() string {
	return w.remoteAddrFunc()
}

func (w *WebSocket) UserAgent() string {
	return w.userAgent
}

func (w *WebSocket) Close() {
	// Lock the mutex
	w.mutex.Lock()

	// Just return if the socket is already closed
	if w.isClosed {
		// Unlock the mutex again
		w.mutex.Unlock()
		return
	}

	// Update the flag
	w.isClosed = true

	// Unlock the mutex again
	w.mutex.Unlock()

	// Send a close message to the client
	w.write(websocket.CloseMessage, []byte{})

	// Close the socket
	w.ws.Close()

	// Trigger the onClose function if defined
	if w.onClose != nil {
		w.onClose()
	}
}

func (w *WebSocket) OnClose(f func()) {
	w.onClose = f
}

func (w *WebSocket) IsClosed() bool {
	return w.isClosed
}

// Write sends the data to the client
func (w *WebSocket) Write(data string) {
	err := w.write(websocket.TextMessage, []byte(data))
	if err != nil {
		log.L.Warning("failed to write to websocket with remote address %s: %s", w.RemoteAddr(), err.Error())

		// Close the websocket on error
		w.Close()
	}
}

func (w *WebSocket) OnRead(f func(string)) {
	w.onRead = f
}

//#################################//
//### WebSocket Layer - Private ###//
//#################################//

// write writes a message with the given message type and payload.
func (w *WebSocket) write(mt int, payload []byte) error {
	w.ws.SetWriteDeadline(time.Now().Add(writeWait))
	return w.ws.WriteMessage(mt, payload)
}

// readLoop reads messages from the websocket
func (w *WebSocket) readLoop() {
	defer func() {
		// Close the socket and the session
		w.Close()
	}()

	// Set the limits
	w.ws.SetReadLimit(maxMessageSize)

	// Set the pong handler
	w.ws.SetPongHandler(func(string) error {
		// Reset the read deadline
		w.ws.SetReadDeadline(time.Now().Add(readWait))
		return nil
	})

	for {
		// Reset the read deadline
		w.ws.SetReadDeadline(time.Now().Add(readWait))

		// Read from the websocket
		_, data, err := w.ws.ReadMessage()
		if err != nil {
			if err != io.EOF {
				log.L.Warning("failed to read data from websocket with remote address %s: %s", w.RemoteAddr(), err.Error())
			}
			return
		}

		// Trigger the onRead function if defined
		if w.onRead != nil {
			w.onRead(string(data))
		}
	}
}

//####################//
//### HTTP Handler ###//
//####################//

func handleWebSocket(rw http.ResponseWriter, req *http.Request) {
	// This has to be a GET request
	if req.Method != "GET" {
		http.Error(rw, "Method not allowed", 405)
		return
	}

	// Upgrade to websocket
	ws, err := upgrader.Upgrade(rw, req, nil)
	if err != nil {
		log.L.Info("failed to upgrade to websocket layer: %s", err.Error())
		http.Error(rw, "Bad Request", 400)
		return
	}

	// Create a new websocket connection struct
	w := NewWebSocket()

	// Set the websocket connection and the user agent
	w.ws = ws
	w.userAgent = req.Header.Get("User-Agent")

	// Get the remote address and set the remote address get function
	remoteAddr, requestMethodUsed := utils.RemoteAddress(req)
	if requestMethodUsed {
		// Obtain the remote address from the websocket
		w.remoteAddrFunc = func() string {
			return utils.RemovePortFromRemoteAddr(w.ws.RemoteAddr().String())
		}
	} else {
		// Obtain the remote address from the current string.
		// It was obtained using the request Headers. So don't use the
		// websocket RemoteAddr() method, because it does not return
		// the clients IP address.
		w.remoteAddrFunc = func() string {
			return remoteAddr
		}
	}

	// Trigger the new socket connection function
	triggerOnNewSocketConnection(w)

	// Start to read messages from the websocket in a new goroutine
	go w.readLoop()
}
