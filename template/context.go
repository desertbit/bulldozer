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
	"strconv"
)

//######################//
//### Context struct ###//
//######################//

type Context struct {
	id           string
	parentID     string
	domID        string
	styleClasses []string

	s *sessions.Session
	t *Template
}

// NewContext creates a new context.
// One optional argument can be passed, which defines additional style classes.
func NewContext(s *sessions.Session, t *Template, id string, parentID string, vars ...[]string) *Context {
	// Create a new context
	c := &Context{
		id:       id,
		parentID: parentID,
		s:        s,
		t:        t,
	}

	// Set the DOM ID
	if len(t.staticDomID) == 0 {
		// Calculate and set the unique DOM ID with the session encryption key
		c.domID = utils.EncryptDomId(c.s.DomEncryptionKey(), "i_"+c.id)
	} else {
		// Use the static DOM ID.
		c.domID = t.staticDomID
	}

	// Add the additional style classes if present.
	if len(vars) >= 1 {
		c.styleClasses = vars[0]
	}

	return c
}

// ID returns the unique ID of this execution context.
// Use this for example as database access keys...
func (c *Context) ID() string {
	return c.id
}

// ParentID returns the main template executing unique ID.
func (c *Context) ParentID() string {
	return c.parentID
}

// DomID returns the DOM ID of the current context
func (c *Context) DomID() string {
	return c.domID
}

// GenDomID generates the real DOM ID of id.
// This is equivalent to the following template call: {{id "YOUR_ID"}}
func (c *Context) GenDomID(id string) string {
	// Create the DOM ID
	domId := "i_" + c.id + "+" + id

	// Calculate the unique DOM ID with the session encryption key
	return utils.EncryptDomId(c.s.DomEncryptionKey(), domId)
}

// Session resturns the current context session
func (c *Context) Session() *sessions.Session {
	return c.s
}

// Template returns the current context template
func (c *Context) Template() *Template {
	return c.t
}

// Styles returns a slice of all template styles
func (c *Context) Styles() []string {
	// Return a merged slice of template styles and context styles.
	return append(append([]string(nil), c.t.styleClasses...), c.styleClasses...)
}

// StylesString returns a string of all template styles
func (c *Context) StylesString() (str string) {
	// Get the styles as slice
	styles := c.Styles()

	// Add the styles to a string separated by one emtpy space.
	if styles != nil && len(styles) > 0 {
		for _, sc := range styles {
			str += sc + " "
		}

		// Remove the last emtpy space
		str = str[:len(str)-1]
	}

	return
}

// Release removes all session template events
// and releases the current context.
func (c *Context) Release() {
	// Remove all registered session events for the current DOM ID.
	releaseSessionTemplateEvents(c.s, c.domID)
}

// Update executes the template and updates the new DOM content
func (c *Context) Update(data interface{}) error {
	// Execute the template
	var b bytes.Buffer
	err := ExecuteContext(c, &b, data)
	if err != nil {
		return err
	}

	// Update the current div wrapper of this template.
	c.s.SendCommand(`Bulldozer.render.updateTemplate("` + c.domID + `",'` + utils.EscapeJS(b.String()) + `');`)

	return nil
}

// TriggerEvent triggers the event on the client side defined with the template event syntax.
func (c *Context) TriggerEvent(eventName string, params ...interface{}) {
	cmd := `Bulldozer.core.emitServerEvent("` + c.domID + `",'` + utils.EscapeJS(eventName) + `'`

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
	c.s.SendCommand(cmd)
}
