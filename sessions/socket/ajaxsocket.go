/*
 *  Bulldozer Framework
 *  Copyright (C) DesertBit
 */

package socket

import (
	"code.desertbit.com/bulldozer/bulldozer/utils"
	"github.com/golang/glog"
	"io"
	"io/ioutil"
	"net/http"
	"strings"
	"sync"
	"time"
)

const (
	ajaxPollTimeout = 35 * time.Second

	ajaxSocketDataInit = "init"
)

var (
	ajaxSockets map[string]*AjaxSocket = make(map[string]*AjaxSocket)
	ajaxMutex   sync.Mutex
)

//#######################//
//### AjaxSocket Layer ###//
//#######################//

type AjaxSocket struct {
	uid       string
	pollToken string

	isClosed  bool
	isClosing chan struct{}
	mutex     sync.Mutex

	writeChannel chan string

	onClose func()
	onRead  func(string)

	userAgent  string
	remoteAddr string
}

func NewAjaxSocket() *AjaxSocket {
	// Create a new websocket struct
	return &AjaxSocket{
		onRead:       nil,
		isClosed:     false,
		onClose:      nil,
		isClosing:    make(chan struct{}),
		writeChannel: make(chan string, 3),
	}
}

func (a *AjaxSocket) Type() SocketType {
	return TypeAjaxSocket
}

func (a *AjaxSocket) RemoteAddr() string {
	return a.remoteAddr
}

func (a *AjaxSocket) UserAgent() string {
	return a.userAgent
}

func (a *AjaxSocket) Close() {
	// Lock the mutex
	a.mutex.Lock()

	// Just return if the socket is already closed
	if a.isClosed {
		// Unlock the mutex again
		a.mutex.Unlock()
		return
	}

	// Update the flag
	a.isClosed = true

	// Unlock the mutex again
	a.mutex.Unlock()

	// Stop the polling goroutine if running
	close(a.isClosing)

	// Remove the ajax socket from the map
	ajaxMutex.Lock()
	delete(ajaxSockets, a.uid)
	ajaxMutex.Unlock()

	// Trigger the onClose function if defined
	if a.onClose != nil {
		a.onClose()
	}
}

func (a *AjaxSocket) OnClose(f func()) {
	a.onClose = f
}

func (a *AjaxSocket) IsClosed() bool {
	return a.isClosed
}

func (a *AjaxSocket) Write(data string) {
	a.writeChannel <- data
}

func (a *AjaxSocket) OnRead(f func(string)) {
	a.onRead = f
}

//####################//
//### HTTP Handler ###//
//####################//

func handleAjaxSocket(w http.ResponseWriter, req *http.Request) {
	// Get the body data
	body, err := ioutil.ReadAll(req.Body)

	// Check for bad requests
	if err != nil || req.Method != "POST" {
		glog.Warningf("client tried to access the ajax interface with an invalid http method: %s", req.Method)
		http.Error(w, "Bad Request", 400)
		return
	}

	data := string(body)

	// Check if the client requests an ajax initialization
	if data == ajaxSocketDataInit {
		var uid string

		// Create a new ajax socket struct
		a := NewAjaxSocket()

		// Lock the mutex
		ajaxMutex.Lock()

		// Obtain a new unique Id
		for {
			// Get a new Id
			uid = utils.RandomString(10)

			// Check if the new Id is already used.
			// This is very unlikely, but we have to check this!
			_, ok := ajaxSockets[uid]
			if !ok {
				// Break the loop if the Id is unique
				break
			}
		}

		// Set the uid and client information
		a.uid = uid
		a.remoteAddr, _ = utils.RemoteAddress(req)
		a.userAgent = req.Header.Get("User-Agent")

		// Create a new poll token
		a.pollToken = utils.RandomString(7)

		// Add the new ajax socket to the map
		ajaxSockets[uid] = a

		// Unlock the mutex again
		ajaxMutex.Unlock()

		// Tell the client the unique Id and poll token
		io.WriteString(w, uid+"&"+a.pollToken)

		// Trigger the new socket connection function
		triggerOnNewSocketConnection(a)

		return
	}

	// Get the uid from the data string
	i := strings.Index(data, "&")
	if i <= 0 {
		glog.Warningf("client didn't send the ajax uid: data: %s", data)
		http.Error(w, "Bad Request", 400)
		return
	}

	uid := data[:i]

	// Remove the uid from the data string
	data = data[i+1:]
	if len(data) == 0 {
		glog.Warningf("client send empty data")
		http.Error(w, "Bad Request", 400)
		return
	}

	// Lock the mutex
	ajaxMutex.Lock()

	// Obtain the ajax socket with the uid
	a, ok := ajaxSockets[uid]
	if !ok {
		// Unlock the mutex again
		ajaxMutex.Unlock()

		glog.Warningf("client requested an invalid ajax socket: uid is invalid")
		http.Error(w, "Bad Request", 400)
		return
	}

	// Unlock the mutex again
	ajaxMutex.Unlock()

	// Update the remote address
	a.remoteAddr, _ = utils.RemoteAddress(req)

	// Trigger the onRead function if defined
	if a.onRead != nil {
		a.onRead(data)
	}
}

func handleAjaxSocketPoll(w http.ResponseWriter, req *http.Request) {
	// Get the body data
	body, err := ioutil.ReadAll(req.Body)

	// Check for bad requests
	if err != nil || req.Method != "POST" {
		glog.Warningf("client tried to access the ajax interface with an invalid http method: %s", req.Method)
		http.Error(w, "Bad Request", 400)
		return
	}

	data := string(body)

	// Get the uid from the data string
	i := strings.Index(data, "&")
	if i <= 0 {
		glog.Warningf("client didn't send the ajax uid: data: %s", data)
		http.Error(w, "Bad Request", 400)
		return
	}

	uid := data[:i]

	// Remove the uid from the data string
	data = data[i+1:]
	if len(data) == 0 {
		glog.Warningf("client send empty token")
		http.Error(w, "Bad Request", 400)
		return
	}

	// Lock the mutex
	ajaxMutex.Lock()

	// Obtain the ajax socket with the uid
	a, ok := ajaxSockets[uid]
	if !ok {
		// Unlock the mutex again
		ajaxMutex.Unlock()

		glog.Warningf("client requested an invalid ajax socket: uid is invalid")
		http.Error(w, "Bad Request", 400)
		return
	}

	// Unlock the mutex again
	ajaxMutex.Unlock()

	// Check if the poll token matches
	if a.pollToken != data {
		glog.Warningf("client has send an invalid poll token!")
		http.Error(w, "Bad Request", 400)
		return
	}

	// Create a new poll token
	a.pollToken = utils.RandomString(7)

	// Create a timeout timer for the poll
	timeout := time.NewTimer(ajaxPollTimeout)

	defer func() {
		// Stop the timeout timer
		timeout.Stop()
	}()

	// Send messages as soon as there are some
	select {
	case data := <-a.writeChannel:
		// Send the new poll token and message to the client
		io.WriteString(w, a.pollToken+"&"+data)
	case <-timeout.C:
		// Do nothing on timeout
		// Just release this go routine
		return
	case <-a.isClosing:
		// Do nothing on timeout
		// Just release this go routine
		return
	}
}
