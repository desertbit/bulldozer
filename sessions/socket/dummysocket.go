/*
 *  Bulldozer Framework
 *  Copyright (C) DesertBit
 */

package socket

import (
	"sync"
)

//###########################//
//### Dummy Socket struct ###//
//###########################//

type DummySocket struct {
	remoteAddr string
	userAgent  string
	isClosed   bool
	mutex      sync.Mutex
	onClose    func()
}

func NewSocketDummy(remoteAddr string, userAgent string) *DummySocket {
	// Create a new dummy socket struct
	return &DummySocket{
		remoteAddr: remoteAddr,
		userAgent:  userAgent,
		isClosed:   false,
		onClose:    nil,
	}
}

func (s *DummySocket) Type() SocketType {
	return TypeDummySocket
}

func (s *DummySocket) RemoteAddr() string {
	return s.remoteAddr
}

func (s *DummySocket) UserAgent() string {
	return s.userAgent
}

func (s *DummySocket) IsClosed() bool {
	return s.isClosed
}

func (s *DummySocket) Close() {
	// Lock the mutex
	s.mutex.Lock()

	// Just return if the socket is already closed
	if s.isClosed {
		// Unlock the mutex again
		s.mutex.Unlock()
		return
	}

	// Update the flag
	s.isClosed = true

	// Unlock the mutex again
	s.mutex.Unlock()

	// Trigger the onClose function if defined
	if s.onClose != nil {
		s.onClose()
	}
}

func (s *DummySocket) OnClose(f func()) {
	s.onClose = f
}

func (s *DummySocket) Write(string) {}

func (s *DummySocket) OnRead(func(string)) {}
