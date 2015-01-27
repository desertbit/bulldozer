/*
 *  Bulldozer Framework
 *  Copyright (C) DesertBit
 */

package template

import (
	"bytes"
	"code.desertbit.com/bulldozer/bulldozer/log"
	"code.desertbit.com/bulldozer/bulldozer/sessions"
	"code.desertbit.com/bulldozer/bulldozer/utils"
	"encoding/gob"
	"fmt"
	"strconv"
	"sync"
)

const (
	GlobalID = "global"
)

func init() {
	// Register the custom types.
	gob.Register(&contextData{})
}

//#########################//
//### Context Data type ###//
//#########################//

// This data can be stored in the session store (ex.: template events).
type contextData struct {
	ID           string
	RootID       string
	DomID        string
	TemplateUID  string
	TemplateName string
	StyleClasses []string

	// Optional values which are stored in the session store.
	// They survive application restarts.
	// Values have to be encodable by gob. Don't store pointers!
	Values      map[interface{}]interface{}
	valuesMutex sync.Mutex
}

func (c *contextData) createValuesMapIfNil() {
	if c.Values == nil {
		c.Values = make(map[interface{}]interface{})
	}
}

//#########################//
//### Context Namespace ###//
//#########################//

type contextNamespace struct {
	s     *sessions.Session
	store *contextStore

	// The execution values lifetime is one complete template
	// execution with all rendered sub templates.
	values map[interface{}]interface{}
	mutex  sync.Mutex
}

func newContextNamespace(s *sessions.Session) *contextNamespace {
	return &contextNamespace{
		s:      s,
		store:  getContextStore(s), // Get the context store if present or nil.
		values: make(map[interface{}]interface{}),
	}
}

func (ns *contextNamespace) AddToStore(d *contextData) {
	if ns.store != nil {
		ns.store.Set(d)
	}
}

func (ns *contextNamespace) RemoveFromStore(d *contextData) {
	if ns.store != nil {
		ns.store.Remove(d)
	}
}

//####################//
//### Context type ###//
//####################//

type Context struct {
	data *contextData
	ns   *contextNamespace
	t    *Template
}

// newContext creates a new context.
func newContext(s *sessions.Session, t *Template, optArgs ...ExecOpts) *Context {
	// Create a new context data value.
	data := &contextData{
		ID:           GlobalID,      // Set the global ID as default value.
		DomID:        t.staticDomID, // Use the static DOM ID by default. If emtpy, a new DOM ID will be calculated...
		TemplateUID:  t.ns.uid,
		TemplateName: t.Name(),
		StyleClasses: t.styleClasses,
	}

	// Apply the optional options.
	if len(optArgs) > 0 {
		opts := &optArgs[0]

		// Set the custom ID if set.
		if len(opts.ID) > 0 {
			data.ID = opts.ID
		}

		// Set the custom DOM ID if set.
		if len(opts.DomID) > 0 {
			data.DomID = opts.DomID
		}

		// Append the additional style classes if present.
		if len(opts.StyleClasses) > 0 {
			data.StyleClasses = append(data.StyleClasses, opts.StyleClasses...)
		}
	}

	// If the global context ID is set, then use this as ID.
	if len(t.globalContextID) > 0 {
		data.ID = t.globalContextID
	}

	// Calculate and set the unique DOM ID with
	// the context ID and session encryption key if the DOM ID is emtpy.
	if len(data.DomID) == 0 {
		data.DomID = utils.EncryptDomId(s.DomEncryptionKey(), "c_"+data.ID)
	}

	// This is the root context. Set the root ID to the ID.
	data.RootID = data.ID

	// Create a new context value.
	c := &Context{
		data: data,
		t:    t,
		ns:   newContextNamespace(s),
	}

	// Add the context data to the store if present.
	c.ns.AddToStore(data)

	return c
}

// One variadiv boolean can be set. If false, the context won't be added to the context store.
func newContextFromData(s *sessions.Session, data *contextData, vars ...bool) (*Context, error) {
	// Get the template namespace with the template UID.
	ns, ok := getNameSpace(data.TemplateUID)
	if !ok {
		return nil, fmt.Errorf("no template namespace found '%s'!", data.TemplateUID)
	}

	// Get the template.
	t := ns.Get(data.TemplateName)
	if t == nil {
		return nil, fmt.Errorf("no template with name '%s' in namespace '%s'!", data.TemplateName, data.TemplateUID)
	}

	// Create a new context value.
	c := &Context{
		data: data,
		t:    t,
		ns:   newContextNamespace(s),
	}

	if len(vars) == 0 || vars[0] {
		// Add the context data to the store if present.
		c.ns.AddToStore(data)
	}

	return c, nil
}

// New creates a new sub context.
// One optional slice can be passed, which defines additional style classes.
func (c *Context) New(t *Template, id string, vars ...[]string) *Context {
	// Create a new context data value.
	data := &contextData{
		ID:           id,
		RootID:       c.data.RootID, // Use the root ID of the parent context.
		DomID:        t.staticDomID, // Use the static DOM ID by default. If emtpy, a new DOM ID will be calculated...
		TemplateUID:  t.ns.uid,
		TemplateName: t.Name(),
		StyleClasses: t.styleClasses,
	}

	// If the global context ID is set, then use this as new ID and Root ID.
	if len(t.globalContextID) > 0 {
		data.ID = t.globalContextID
		data.RootID = t.globalContextID
	}

	// Calculate and set the unique DOM ID with
	// the context ID and session encryption key if the DOM ID is emtpy.
	if len(data.DomID) == 0 {
		data.DomID = utils.EncryptDomId(c.ns.s.DomEncryptionKey(), "c_"+data.ID)
	}

	// Add the additional style classes if present.
	if len(vars) > 0 {
		data.StyleClasses = append(data.StyleClasses, vars[0]...)
	}

	// Create a new context value.
	subC := &Context{
		data: data,
		t:    t,
		ns:   c.ns, // Contexts share the same namespace.
	}

	// Add the context data to the store if present.
	c.ns.AddToStore(data)

	return subC
}

// ID returns the unique ID of this execution context.
// Use this for example as database access keys...
func (c *Context) ID() string {
	return c.data.ID
}

// RootID returns the unique ID of the root template.
func (c *Context) RootID() string {
	return c.data.RootID
}

// DomID returns the DOM ID of the current context.
func (c *Context) DomID() string {
	return c.data.DomID
}

// GenDomID generates the real DOM ID of id.
// This is equivalent to the following template call: {{id "YOUR_ID"}}
func (c *Context) GenDomID(id string) string {
	// Create the DOM ID
	domId := "c_" + c.data.ID + "+" + id

	// Calculate the unique DOM ID with the session encryption key.
	return utils.EncryptDomId(c.ns.s.DomEncryptionKey(), domId)
}

// Session resturns the current context session.
func (c *Context) Session() *sessions.Session {
	return c.ns.s
}

// Template returns the current context template.
func (c *Context) Template() *Template {
	return c.t
}

// Styles returns a slice of all template styles.
func (c *Context) Styles() []string {
	return c.data.StyleClasses
}

// StylesString returns a string of all template styles.
func (c *Context) StylesString() (str string) {
	// Get the slice.
	styles := c.data.StyleClasses

	// Add the styles to a string separated by one emtpy space.
	if len(styles) > 0 {
		for _, sc := range styles {
			str += sc + " "
		}

		// Remove the last emtpy space.
		str = str[:len(str)-1]
	}

	return
}

// Release removes all session template events
// and releases the current context.
func (c *Context) Release() {
	// Remove all registered session events for the current DOM ID.
	releaseSessionTemplateEvents(c.ns.s, c.data.DomID)

	// Remove context data from the store.
	c.ns.RemoveFromStore(c.data)
}

// Update executes the template and updates the new DOM content.
// One optional argument can be passed. This is the render data for the template.
// If omitted, the data is obtained through the template.OnGetData function.
func (c *Context) Update(vars ...interface{}) error {
	var data interface{}
	if len(vars) > 0 {
		data = vars[0]
	}

	// Execute the template
	var b bytes.Buffer
	err := ExecuteContext(c, &b, data)
	if err != nil {
		return err
	}

	// Update the current div wrapper of this template.
	c.ns.s.SendCommand(`Bulldozer.render.updateTemplate("` + c.data.DomID + `",'` + utils.EscapeJS(b.String()) + `');`)

	return nil
}

// TriggerEvent triggers the event on the client side defined with the template event syntax.
func (c *Context) TriggerEvent(eventName string, params ...interface{}) {
	cmd := `Bulldozer.core.emitServerEvent("` + c.data.DomID + `",'` + utils.EscapeJS(eventName) + `'`

	// Append all the parameters
	for i, param := range params {
		// type assertion
		switch v := param.(type) {
		case int:
			cmd += "," + strconv.Itoa(v)
		case bool:
			cmd += "," + strconv.FormatBool(v)
		case string:
			cmd += ",'" + utils.EscapeJS(v) + "'"
		default:
			log.L.Error("context: trigger event: invalid type of function event parameter: %v : parameters: %v", i+1, params)
			return
		}
	}

	cmd += ");"

	// Send the command to the client
	c.ns.s.SendCommand(cmd)
}

//############################//
//### Context Store Values ###//
//############################//

// StoreGet obtains the store value, which are stored in the session store.
// They survive application restarts.
// Values have to be encodable by gob. Don't store pointers!
// A single variadic argument is accepted, and it is optional:
// if a function is set, this function will be called if no value
// exists for the given key.
// This operation is thread-safe.
func (c *Context) StoreGet(key interface{}, vars ...func() interface{}) (value interface{}, ok bool) {
	// Get the data pointer.
	data := c.data

	// Lock the mutex.
	data.valuesMutex.Lock()
	defer data.valuesMutex.Unlock()

	// Get the value if the map is not nil.
	ok = false
	if data.Values != nil {
		value, ok = data.Values[key]
	}

	// If no value is found and the create function variable
	// is set, then call the function and set the new value.
	if !ok && len(vars) > 0 {
		value = vars[0]()

		data.createValuesMapIfNil()

		data.Values[key] = value
		ok = true

		// Tell the session, that data might have changed.
		c.ns.s.Dirty()
	}

	return
}

// StorePull does the same as Get(), but additionally removes the value from the store if present.
// Use this for Flash values...
func (c *Context) StorePull(key interface{}, vars ...func() interface{}) (interface{}, bool) {
	i, ok := c.StoreGet(key, vars...)
	if ok {
		c.StoreDelete(key)
	}

	return i, ok
}

// StoreSet sets the store value with the given key. This operation is thread-safe.
func (c *Context) StoreSet(key interface{}, value interface{}) {
	// Get the data pointer.
	data := c.data

	// Lock the mutex.
	data.valuesMutex.Lock()
	defer data.valuesMutex.Unlock()

	// Set the value.
	data.createValuesMapIfNil()
	data.Values[key] = value

	// Tell the session, that data might have changed.
	c.ns.s.Dirty()
}

// StoreDelete removes the store value with the given key.
// This operation is thread-safe.
func (c *Context) StoreDelete(key interface{}) {
	// Get the data pointer.
	data := c.data

	// Lock the mutex.
	data.valuesMutex.Lock()
	defer data.valuesMutex.Unlock()

	if data.Values == nil {
		return
	}

	delete(data.Values, key)

	// Tell the session, that data might have changed.
	c.ns.s.Dirty()
}

//########################//
//### Execution Values ###//
//########################//

// Get obtains the execution value. Execution values exist for one complete execution cycle.
// A single variadic argument is accepted, and it is optional:
// if a function is set, this function will be called if no value
// exists for the given key.
// This operation is thread-safe.
func (c *Context) Get(key interface{}, vars ...func() interface{}) (value interface{}, ok bool) {
	// Get the namespace.
	ns := c.ns

	// Lock the mutex.
	ns.mutex.Lock()
	defer ns.mutex.Unlock()

	value, ok = ns.values[key]

	// If no value is found and the create function variable
	// is set, then call the function and set the new value.
	if !ok && len(vars) > 0 {
		value = vars[0]()
		ns.values[key] = value
		ok = true
	}

	return
}

// Pull does the same as Get(), but additionally removes the value from the map if present.
// Use this for Flash values...
func (c *Context) Pull(key interface{}, vars ...func() interface{}) (interface{}, bool) {
	i, ok := c.Get(key, vars...)
	if ok {
		c.Delete(key)
	}

	return i, ok
}

// Set sets the execution value with the given key. This operation is thread-safe.
func (c *Context) Set(key interface{}, value interface{}) {
	// Get the namespace.
	ns := c.ns

	// Lock the mutex.
	ns.mutex.Lock()
	defer ns.mutex.Unlock()

	ns.values[key] = value
}

// Delete removes the execution value with the given key. This operation is thread-safe.
func (c *Context) Delete(key interface{}) {
	// Get the namespace.
	ns := c.ns

	// Lock the mutex.
	ns.mutex.Lock()
	defer ns.mutex.Unlock()

	delete(ns.values, key)
}
