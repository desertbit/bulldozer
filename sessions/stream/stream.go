/*
 *  Bulldozer Framework
 *  Copyright (C) DesertBit
 */

package stream

import (
	"sync"
	"time"
)

//######################//
//### Stream struct ###//
//######################//

// This stream struct is used as buffer to send streams to the the socket
// implementation (Ajax socket or Websocket)
type Stream struct {
	HasData chan bool

	data  string
	mutex sync.Mutex
}

func New() *Stream {
	return &Stream{
		HasData: make(chan bool, 1),
	}
}

// Write appends the data string to the data stream buffer
func (m *Stream) Write(data string) {
	// Lock mutex
	m.mutex.Lock()
	defer m.mutex.Unlock()

	// Append data to the data string
	m.data += data

	// Only trigger the channel if not already.
	// This way, this channel will never block.
	select {
	case m.HasData <- true:
	default:
	}
}

// Read gets the data in the stream buffer
func (m *Stream) Read() (data string) {
	// Wait for 1 millisecond, so that other data might
	// be added to the stream data. This way, two calls to stream.write
	// don't send two messages to the server.
	// Instead only one combined message is send.
	time.Sleep(1e6)

	// Lock mutex
	m.mutex.Lock()
	defer m.mutex.Unlock()

	// Set the return data
	data = m.data

	// Clear the original data
	m.data = ""

	return
}
