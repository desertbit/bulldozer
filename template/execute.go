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

//#############//
//### Types ###//
//#############//

// renderData holds the template context and the execution data.
type renderData struct {
	Context *Context
	Pkg     map[string]interface{}
	Data    interface{}
}

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
	// Create a new context.
	c := newContext(s, t, optArgs...)

	// Execute the context.
	err := ExecuteContext(c, wr, data)
	if err != nil {
		return nil, err
	}

	return c, nil
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

	// Create a new context.
	c := newContext(s, tt, optArgs...)

	// Execute the context.
	err := ExecuteContext(c, wr, data)
	if err != nil {
		return nil, true, err
	}

	return c, true, nil
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
	c.Release()

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
			_, out, _ := backend.ExecErrorTemplate(c.ns.s, action.data, false)
			wr.Write([]byte(out))
			return nil
		} else if action.action == actionRedirect {
			// Navigate to the path.
			backend.NavigateToPath(c.ns.s, action.data)
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
