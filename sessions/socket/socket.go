/*
 *  Bulldozer Framework
 *  Copyright (C) DesertBit
 */

package socket

import (
	"net/http"
)

const (
	// The socket types
	TypeDummySocket SocketType = 1 << iota
	TypeAjaxSocket  SocketType = 1 << iota
	TypeWebSocket   SocketType = 1 << iota

	// Socket keys and sessions
	keySessionID        = "sid"
	valueSessionInit    = "init"
	valueInvalidSession = "invalid_session"
	valueInvalidRequest = "invalid_request"
	keyToken            = "tok"
)

var (
	onNewSocketConnectionFunc func(Socket)
)

type SocketType int

//##############//
//### Public ###//
//##############//

func InitHttpHandlers() {
	// Create the ajax handlers
	http.HandleFunc("/bulldozer/ajax", handleAjaxSocket)
	http.HandleFunc("/bulldozer/ajax/poll", handleAjaxSocketPoll)

	// Create the websocket handler
	http.HandleFunc("/bulldozer/ws", handleWebSocket)
}

func OnNewSocketConnection(f func(Socket)) {
	onNewSocketConnectionFunc = f
}

//###############//
//### Private ###//
//###############//

func triggerOnNewSocketConnection(s Socket) {
	if onNewSocketConnectionFunc != nil {
		onNewSocketConnectionFunc(s)
	}
}

//########################//
//### Socket interface ###//
//########################//

type Socket interface {
	Type() SocketType
	RemoteAddr() string
	UserAgent() string

	Close()
	IsClosed() bool
	OnClose(func())

	Write(data string)
	OnRead(func(data string))
}
