/*
 *  Bulldozer Framework
 *  Copyright (C) DesertBit
 */

package template

import (
	"code.desertbit.com/bulldozer/bulldozer/log"
	"code.desertbit.com/bulldozer/bulldozer/sessions"
	"code.desertbit.com/bulldozer/bulldozer/utils"
	"encoding/gob"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"sync"
)

const (
	EventMethodPrefix = "Event"

	globalEventsNameSpace = "global"
	eventsAccessKeyLength = 10

	// Session instance value keys:
	instanceKeyEvents = "blzEvents"

	requestTypeEmit = "emit"
	keyEmitDomID    = "did"
	keyEmitKey      = "key"
	keyEmitParam    = "arg"
)

func init() {
	// Register the custom types to gob.
	gob.Register(&sessionEvents{})
	gob.Register(&sessionEvent{})

	// Register the emit template parse function.
	registerParseFunc("emit", parseEmit)

	// Register the emit server request.
	err := sessions.Request(requestTypeEmit, sessionRequestEmit)
	if err != nil {
		log.L.Fatalf("failed to register session emit request: %v", err)
	}
}

//###############################//
//### Public Template methods ###//
//###############################//

// RegisterEvents registeres the event methods of the interface to all templates.
// Event method names have to start with the EventMethodPrefix.
// They are called from the client-side without this prefix.
// One optional parameter can be set, to define the events namespace.
// If no namespace is defined, then the event is registered in the global namespace.
func (t *Template) RegisterEvents(i interface{}, vars ...string) *Template {
	// Get the namespace
	var namespace string
	if len(vars) > 0 {
		namespace = vars[0]
	}
	if len(namespace) == 0 {
		namespace = globalEventsNameSpace
	}

	var events *events
	func() {
		// Lock the mutex
		t.ns.eventsMapMutex.Lock()
		defer t.ns.eventsMapMutex.Unlock()

		// Obtain the events value.
		// If it doesn't exists for the namespace, then create one.
		var ok bool
		events, ok = t.ns.eventsMap[namespace]
		if !ok {
			// Create a new events value and pass the template.
			events = newEvents(t)

			// Add the new events to the map
			t.ns.eventsMap[namespace] = events
		}
	}()

	// Register all methods of the interface which start
	// with the events prefix.
	noMethods := true
	iType := reflect.TypeOf(i)
	for x := 0; x < iType.NumMethod(); x++ {
		method := iType.Method(x)
		name := method.Name

		// Skip this method if it does not start with the method prefix.
		if !strings.HasPrefix(name, EventMethodPrefix) {
			continue
		}

		// Trim the prefix from the name
		name = strings.TrimPrefix(name, EventMethodPrefix)

		// Create a new event value.
		e := &event{
			receiver: i,
			method:   method.Func,
		}

		// Add the method to the events
		if overwritten := events.Set(name, e); overwritten {
			// Print a warning message.
			log.L.Error("template: RegisterEvents: overwritting event function: '%s.%s'!", namespace, name)
		}

		// Set the flag
		noMethods = false
	}

	if noMethods {
		log.L.Warning("template: RegisterEvents: registered interface event methods, but no event methods where found!")
	}

	return t
}

//###########################//
//### Private events type ###//
//###########################//

type event struct {
	receiver interface{}
	method   reflect.Value
}

type events struct {
	t *Template

	// Key:   event function name
	// Value: event function
	funcs      map[string]*event
	funcsMutex sync.Mutex
}

func newEvents(t *Template) *events {
	return &events{
		t:     t,
		funcs: make(map[string]*event),
	}
}

func (es *events) Exists(funcName string) (ok bool) {
	// Lock the mutex
	es.funcsMutex.Lock()
	defer es.funcsMutex.Unlock()

	// Check if the value with key exists.
	_, ok = es.funcs[funcName]
	return
}

func (es *events) Get(funcName string) (e *event, ok bool) {
	// Lock the mutex
	es.funcsMutex.Lock()
	defer es.funcsMutex.Unlock()

	// Get the value with the name as key
	e, ok = es.funcs[funcName]
	return
}

func (es *events) Set(funcName string, e *event) (overwritten bool) {
	// Set the default flag value.
	overwritten = false

	// Lock the mutex
	es.funcsMutex.Lock()
	defer es.funcsMutex.Unlock()

	// Set the boolean if the event is overwritten
	if _, ok := es.funcs[funcName]; ok {
		overwritten = true
	}

	// Set the value
	es.funcs[funcName] = e

	return
}

//###################################//
//### Private session value types ###//
//###################################//

type sessionEvents struct {
	// First key:    template DOM ID
	// Second key:   event access key
	Events map[string]map[string]*sessionEvent
	mutex  sync.Mutex
}

func newSessionEvents() *sessionEvents {
	return &sessionEvents{
		Events: make(map[string]map[string]*sessionEvent),
	}
}

type sessionEvent struct {
	FuncNameSpace string
	FuncName      string
	ContextData   *contextData
}

//###############//
//### Private ###//
//###############//

func releaseAllSessionEvents(s *sessions.Session) {
	s.InstanceDelete(instanceKeyEvents)
}

func releaseSessionTemplateEvents(s *sessions.Session, domID string) {
	// Get the session events.
	sEvents := getSessionEvents(s)

	// Lock the mutex.
	sEvents.mutex.Lock()
	defer sEvents.mutex.Unlock()

	// Remove the template session events from the map.
	delete(sEvents.Events, domID)

	// Mark the session as dirty, because template events were removed.
	s.Dirty()
}

func parseEmit(typeStr string, token string, d *parseData) error {
	// Try to find the '(' symbol
	pos := strings.Index(token, "(")

	// If not found, throw and error and exit
	if pos == -1 {
		return fmt.Errorf("missing emit function bracket '('!")
	}

	// Extract the function name and remove it from the original string
	funcName := strings.TrimSpace(token[0:pos])
	token = strings.TrimSpace(token[pos+1:])

	// Find the function namespace if present
	namespace := globalEventsNameSpace
	pos = strings.Index(funcName, ".")
	if pos != -1 {
		// Extract the function namespace and remove it from the function name.
		namespace = funcName[:pos]
		funcName = funcName[pos+1:]
	}

	// Check if the function name and namespace is valid.
	if funcName == "" {
		return fmt.Errorf("empty emit function name!")
	} else if namespace == "" {
		return fmt.Errorf("empty emit function namespace!")
	}

	// The last symbol has to be the ending bracket
	if token[len(token)-1] != ')' {
		return fmt.Errorf("missing emit function ending bracket ')'!")
	}

	// Remove the last ending bracket and trim spaces.
	token = strings.TrimSpace(token[:len(token)-1])

	// Generate the javascript code to call the function
	// with the DOM ID and the access key.
	cmd := `Bulldozer.core.emit("{{$.Context.DomID}}","{{eventKey $.Context "` + namespace + `" "` + funcName + `"}}"`

	// Add the function arguments if present
	if len(token) > 0 {
		cmd += "," + token
	}

	// Add the ending bracket
	cmd += ");"

	// Add the command to the final string
	*d.final += cmd

	return nil
}

func createEventAccessKey(c *Context, namespace string, funcName string) (string, error) {
	ns := c.t.ns
	s := c.ns.s

	// Check if the function exists.
	events, ok := func() (e *events, ok bool) {
		// Lock the mutex
		ns.eventsMapMutex.Lock()
		defer ns.eventsMapMutex.Unlock()

		// Obtain the events value.
		e, ok = ns.eventsMap[namespace]
		return
	}()

	// Throw an error if the namespace is invalid.
	if !ok {
		if namespace == globalEventsNameSpace {
			return "", fmt.Errorf("emit call: '%s': event function does not exists in the global namespace!", funcName)
		} else {
			return "", fmt.Errorf("emit call: '%s.%s': namespace does not exists '%s'", namespace, funcName, namespace)
		}
	}
	// Throw an error if the function does not exists.
	if !events.Exists(funcName) {
		return "", fmt.Errorf("emit call: '%s.%s': function does not exists '%s'", namespace, funcName, funcName)
	}

	// Create a new session event value.
	event := &sessionEvent{
		FuncName:      funcName,
		FuncNameSpace: namespace,
		ContextData:   c.data,
	}

	// Get the session events.
	sEvents := getSessionEvents(s)

	// Lock the mutex.
	sEvents.mutex.Lock()
	defer sEvents.mutex.Unlock()

	// Get the template events map.
	templateEvents, ok := sEvents.Events[c.data.DomID]
	if !ok {
		// Create a map for the current template.
		templateEvents = make(map[string]*sessionEvent)
		sEvents.Events[c.data.DomID] = templateEvents
	}

	// Create a unique event access key.
	var key string
	for {
		key = utils.RandomString(eventsAccessKeyLength)
		if _, ok := templateEvents[key]; !ok {
			// Break the map if the key is unique.
			break
		}
		// Continue, because the key already exists.
	}

	// Add the new event with the key to the template events map.
	templateEvents[key] = event

	// Mark the session as dirty, because template events
	// were registered to the session instance values.
	s.Dirty()

	// Return the new event access key.
	return key, nil
}

func getSessionEvents(s *sessions.Session) *sessionEvents {
	// Get the session events value. Create and add it, if not present.
	eventsI, _ := s.InstanceGet(instanceKeyEvents, func() interface{} {
		return newSessionEvents()
	})

	// Assertion
	events, ok := eventsI.(*sessionEvents)
	if !ok {
		// Log the error
		log.L.Error("template emit: failed to assert value to session events value!")

		// Just create a new one and set it to the instance values.
		events = newSessionEvents()
		s.InstanceSet(instanceKeyEvents, events)
	}

	return events
}

// sessionRequestEmit is triggered from the client side.
// Call the template function with the given key and pass the parameters...
func sessionRequestEmit(s *sessions.Session, data map[string]string) error {
	// Recover panics and log the error message.
	defer func() {
		if e := recover(); e != nil {
			log.L.Error("bulldozer template emit panic: %v", e)
		}
	}()

	// Try to obtain the DOM ID
	domID, ok := data[keyEmitDomID]
	if !ok {
		return fmt.Errorf("emit request: DOM ID is missing in the request: %v", data)
	}

	// Try to obtain the function key
	key, ok := data[keyEmitKey]
	if !ok {
		return fmt.Errorf("emit request: event access key is missing in the request: %v", data)
	}

	// Get all parameters (arg1, arg2, arg3, ...) and add it to a slice
	var params []string
	for i := 1; ; i++ {
		p, ok := data[keyEmitParam+strconv.Itoa(i)]
		if !ok {
			break
		}

		params = append(params, p)
	}

	// Get the event if present.
	sEvent, ok := func() (event *sessionEvent, ok bool) {
		// Get the session events.
		sEvents := getSessionEvents(s)

		// Lock the mutex.
		sEvents.mutex.Lock()
		defer sEvents.mutex.Unlock()

		// Get the template events map.
		templateEvents, ok := sEvents.Events[domID]
		if !ok {
			return nil, false
		}

		event, ok = templateEvents[key]
		return
	}()
	if !ok {
		return fmt.Errorf("invalid emit call from client: domID '%s' key '%s' parameters '%v': no session event registered.", domID, key, params)
	}

	// Get the context data.
	cData := sEvent.ContextData

	// Create the template context.
	c, err := newContextFromData(s, cData)
	if err != nil {
		return fmt.Errorf("invalid emit call from client: domID '%s' key '%s' parameters '%v': %v", domID, key, params, err)
	}

	// Get the template namespace pointer.
	ns := c.t.ns

	// Get the event functions of the given namespace.
	events, ok := func() (e *events, ok bool) {
		// Lock the mutex
		ns.eventsMapMutex.Lock()
		defer ns.eventsMapMutex.Unlock()

		// Obtain the events value.
		e, ok = ns.eventsMap[sEvent.FuncNameSpace]
		return
	}()
	if !ok {
		return fmt.Errorf("invalid emit call from client: domID '%s' key '%s' parameters '%v': invalid template event namespace '%s'", domID, key, params, sEvent.FuncNameSpace)
	}

	// Get the function reflect value.
	event, ok := events.Get(sEvent.FuncName)
	if !ok {
		return fmt.Errorf("invalid emit call from client: domID '%s' key '%s' parameters '%v': no event function defined '%s'", domID, key, params, sEvent.FuncName)
	}

	// Get the type of the function.
	t := event.method.Type()

	// Get the number of parameters.
	funcNumIn := t.NumIn()

	// Check if the first function parameter is of type *Context
	if funcNumIn < 2 || reflect.TypeOf(c) != t.In(1) {
		return fmt.Errorf("invalid emit call from client: domID '%s' key '%s' parameters '%v': the event function's first parameter has to be a *template.Context pointer!", domID, key, params)
	}

	// The receiver is the first in argument. They have to match!
	if reflect.TypeOf(event.receiver) != t.In(0) {
		return fmt.Errorf("invalid emit call from client: domID '%s' key '%s' parameters '%v': the event function's receiver is invalid!", domID, key, params)
	}

	// Check if the number of parameters are valid.
	if len(params)+2 != funcNumIn {
		return fmt.Errorf("invalid emit call from client: domID '%s' key '%s' parameters '%v': event parameters don't match!", domID, key, params)
	}

	// Create the parameters slice
	in := make([]reflect.Value, len(params)+2)

	// Add the template receiver and the context as first function parameters
	in[0] = reflect.ValueOf(event.receiver)
	in[1] = reflect.ValueOf(c)

	// Add all other parameters to the slice
	var funcIndex int
	for k, param := range params {
		funcIndex = k + 2

		// Convert the parameter if required and add it to the slice
		switch t.In(funcIndex).Kind() {
		case reflect.String:
			in[funcIndex] = reflect.ValueOf(param)
		case reflect.Int:
			// Convert the string to an integer
			i, err := strconv.Atoi(param)
			if err != nil {
				return fmt.Errorf("invalid emit call from client: domID '%s' key '%s' parameters '%v'", domID, key, params)
			}

			in[funcIndex] = reflect.ValueOf(i)
		case reflect.Int64:
			// Convert the string to an 64-bit integer
			i, err := strconv.ParseInt(param, 10, 64)
			if err != nil {
				return fmt.Errorf("invalid emit call from client: domID '%s' key '%s' parameters '%v'", domID, key, params)
			}

			in[funcIndex] = reflect.ValueOf(i)
		case reflect.Bool:
			// Convert the string to a boolean
			b, err := strconv.ParseBool(param)
			if err != nil {
				return fmt.Errorf("invalid emit call from client: domID '%s' key '%s' parameters '%v'", domID, key, params)
			}

			in[funcIndex] = reflect.ValueOf(b)
		default:
			// Can't convert the parameter
			return fmt.Errorf("invalid emit call from client: domID '%s' key '%s' parameters '%v'", domID, key, params)
		}
	}

	// Call the function
	event.method.Call(in)

	return nil
}
