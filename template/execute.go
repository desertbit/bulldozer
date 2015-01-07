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

//###############################//
//### Template struct methods ###//
//###############################//

// Execute applies a parsed template to the specified data object,
// writing the output to wr.
// If an error occurs executing the template or writing its output,
// execution stops, but partial results may already have been written to
// the output writer.
// A template may be executed safely in parallel.
// First optional string is an ID string, which is added to the unique context ID.
// All further optional strings are additional template style classes.
func (t *Template) Execute(s *sessions.Session, wr io.Writer, data interface{}, vars ...string) error {
	return execute(t, s, wr, data, vars...)
}

// ExecuteTemplate applies the template associated with t that has the given
// name to the specified data object and writes the output to wr.
// If an error occurs executing the template or writing its output,
// execution stops, but partial results may already have been written to
// the output writer.
// A template may be executed safely in parallel.
// A boolean is returned, defining if the template exists...
// First optional string is an ID string, which is added to the unique context ID.
// All further optional strings are additional template style classes.
func (t *Template) ExecuteTemplate(s *sessions.Session, wr io.Writer, name string, data interface{}, vars ...string) (bool, error) {
	tt := t.Lookup(name)
	if tt == nil {
		return false, fmt.Errorf("failed to execute template: template not found with name '%s'", name)
	}

	return true, execute(tt, s, wr, data, vars...)
}

// ExecuteToString does the same as Execute, but instead writes the output to a string.
func (t *Template) ExecuteToString(s *sessions.Session, data interface{}, vars ...string) (string, error) {
	var b bytes.Buffer
	err := t.Execute(s, &b, data)
	if err != nil {
		return "", err
	}

	return b.String(), err
}

// ExecuteTemplateToString does the same as ExecuteTemplate, but instead writes the output to a string.
func (t *Template) ExecuteTemplateToString(s *sessions.Session, name string, data interface{}, vars ...string) (string, bool, error) {
	var b bytes.Buffer
	found, err := t.ExecuteTemplate(s, &b, name, data)
	if err != nil {
		return "", found, err
	}

	return b.String(), found, err
}

//##############//
//### Public ###//
//##############//

// ExecuteContext executes the template context.
func ExecuteContext(c *Context, wr io.Writer, data interface{}) error {
	// Get the template pointer.
	t := c.t

	// Remove all previously registered session events for the current DOM ID.
	// They will be registered by the following template execution.
	releaseSessionTemplateEvents(c.s, c.domID)

	// Call the must function.
	action := t.callMustFuncs(c)
	if action != nil && action.stopped {
		// TODO: Finish this
		return fmt.Errorf("TODO: Finish this action!")
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
func execute(t *Template, s *sessions.Session, wr io.Writer, data interface{}, vars ...string) error {
	varsLen := len(vars)

	// Create the ID
	id := t.Name()
	if varsLen > 0 {
		id += "@" + vars[0]
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

	// Add additional style classes if present
	if varsLen > 1 {
		c.styleClasses = append(c.styleClasses, vars[1:]...)
	}

	return ExecuteContext(c, wr, data)
}
