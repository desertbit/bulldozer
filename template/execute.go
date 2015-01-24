/*
 *  Bulldozer Framework
 *  Copyright (C) DesertBit
 */

// This template extends the golang template with a custom parse method.
// Some methods are copied from the original golang template to ensure compatibility.

package template

import (
	"bytes"
	"code.desertbit.com/bulldozer/bulldozer/sessions"
	"fmt"
	"io"
)

const (
	GlobalID = "global"
)

//##########################//
//### Optional Data type ###//
//##########################//

type ExecOpts struct {
	ID           string   // This is added to the unique context ID.
	DomID        string   // Set this, to set a custom DOM ID.
	StyleClasses []string // Additional style classes.
}

//###############################//
//### Template struct methods ###//
//###############################//

// Execute applies a parsed template to the specified data object,
// writing the output to wr.
// If an error occurs executing the template or writing its output,
// execution stops, but partial results may already have been written to
// the output writer.
// A template may be executed safely in parallel.
// Optional options can be passed.
func (t *Template) Execute(s *sessions.Session, wr io.Writer, data interface{}, optArgs ...ExecOpts) (*Context, error) {
	return execute(t, s, wr, data, optArgs...)
}

// ExecuteTemplate applies the template associated with t that has the given
// name to the specified data object and writes the output to wr.
// If an error occurs executing the template or writing its output,
// execution stops, but partial results may already have been written to
// the output writer.
// A template may be executed safely in parallel.
// A boolean is returned, defining if the template exists...
// Optional options can be passed.
func (t *Template) ExecuteTemplate(s *sessions.Session, wr io.Writer, name string, data interface{}, optArgs ...ExecOpts) (*Context, bool, error) {
	tt := t.Lookup(name)
	if tt == nil {
		return nil, false, fmt.Errorf("failed to execute template: template not found with name '%s'", name)
	}

	c, err := execute(tt, s, wr, data, optArgs...)
	return c, true, err
}

// ExecuteToString does the same as Execute, but instead writes the output to a string.
func (t *Template) ExecuteToString(s *sessions.Session, data interface{}, optArgs ...ExecOpts) (string, *Context, error) {
	var b bytes.Buffer
	c, err := t.Execute(s, &b, data, optArgs...)
	if err != nil {
		return "", nil, err
	}

	return b.String(), c, err
}

// ExecuteTemplateToString does the same as ExecuteTemplate, but instead writes the output to a string.
func (t *Template) ExecuteTemplateToString(s *sessions.Session, name string, data interface{}, optArgs ...ExecOpts) (string, *Context, bool, error) {
	var b bytes.Buffer
	c, found, err := t.ExecuteTemplate(s, &b, name, data, optArgs...)
	if err != nil {
		return "", nil, found, err
	}

	return b.String(), c, found, err
}

//##############//
//### Public ###//
//##############//

// ExecuteContext executes the template context.
func ExecuteContext(c *Context, wr io.Writer, data interface{}) (err error) {
	// Recover panics and return the error message.
	defer func() {
		if e := recover(); e != nil {
			err = fmt.Errorf("bulldozer execute template panic: %v", e)
		}
	}()

	// Get the template pointer.
	t := c.t

	// Remove all previously registered session events for the current DOM ID.
	// They will be registered by the following template execution.
	releaseSessionTemplateEvents(c.s, c.domID)

	// Return the last parse error if present.
	if t.hasParseError != nil {
		return t.hasParseError
	}

	// Call the must function.
	action := t.callMustFuncs(c)
	if action != nil && action.action != actionContinue {
		if action.action == actionError {
			// Execute the error template without logging the error
			// and write it to the io writer.
			_, out, _ := backend.ExecErrorTemplate(c.s, action.data, false)
			wr.Write([]byte(out))
			return nil
		} else if action.action == actionRedirect {
			// Navigate to the path.
			backend.NavigateToPath(c.s, action.data)
			return nil
		} else {
			return fmt.Errorf("invalid template action type: %v", action.action)
		}
	}

	// Trigger the template execution event.
	t.triggerOnTemplateExecution(c, data)

	// Trigger the template execution finished event on exit.
	defer t.triggerOnTemplateExecutionFinished(c, data)

	// Create the render data
	d := renderData{
		Context: c,
		Data:    data,
		Pkg:     packages,
	}

	return t.template.Execute(wr, &d)
}

//###############//
//### Private ###//
//###############//

// renderData holds the template context and the execution data.
type renderData struct {
	Context *Context
	Pkg     map[string]interface{}
	Data    interface{}
}

// Execute executes the passed template.
// First optional string is an ID string, which is added to the unique context ID.
// All further optional strings are additional template style classes.
func execute(t *Template, s *sessions.Session, wr io.Writer, data interface{}, optArgs ...ExecOpts) (*Context, error) {
	var id string

	// Apply the optional options.
	var opts *ExecOpts
	if len(optArgs) > 0 {
		opts = &optArgs[0]

		// Add the custom ID.
		if len(opts.ID) != 0 {
			id = opts.ID
		}
	}

	// Prepare the ID.
	if len(id) == 0 {
		id = GlobalID
	}

	// Create a new context with the unique ID. The parent ID is the current ID,
	// because this is the executing template.
	// If the global context ID is set, then use this as ID.
	var c *Context
	if len(t.globalContextID) == 0 {
		c = NewContext(s, t, id, id)
	} else {
		c = NewContext(s, t, t.globalContextID, t.globalContextID)
	}

	// Apply the optional options.
	if opts != nil {
		// Set the custom DOM ID if set.
		if len(opts.DomID) != 0 {
			c.domID = opts.DomID
		}

		// Add additional style classes if present.
		if len(opts.StyleClasses) > 0 {
			c.styleClasses = append(c.styleClasses, opts.StyleClasses...)
		}
	}

	return c, ExecuteContext(c, wr, data)
}
