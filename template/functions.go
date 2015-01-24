/*
 *  Bulldozer Framework
 *  Copyright (C) DesertBit
 */

package template

import (
	tr "code.desertbit.com/bulldozer/bulldozer/translate"
	"fmt"
)

var (
	bulldozerFuncMap FuncMap = FuncMap{
		"tr":         tr.S,
		"plugin":     renderPlugin,
		"eventKey":   createEventAccessKey,
		"passValues": passValues,
		"tmplC":      templateContext,
		"loadJS":     loadJavaScript,
		"loadStyle":  loadStyleSheet,
	}
)

//##################################//
//### Private template functions ###//
//##################################//

func templateContext(templateName string, id string, r *renderData, values ...interface{}) (*renderData, error) {
	// Get the context.
	c := r.Context

	// Get the right sub template.
	t := c.t.Lookup(templateName)
	if t == nil {
		return nil, fmt.Errorf("no template found with name: '%s'", templateName)
	}

	// Create the unique sub template ID if present.
	// Otherwise use the previous ID.
	if len(id) > 0 {
		id = c.id + "_" + id
	} else {
		id = c.id
	}

	// Create the template context.
	if len(t.globalContextID) == 0 {
		c = NewContext(c.s, t, id, c.parentID)
	} else {
		c = NewContext(c.s, t, t.globalContextID, t.globalContextID)
	}

	// Create a new render data for the sub template.
	subR := &renderData{
		Context: c,
		Data:    r.Data,
		Pkg:     r.Pkg,
	}

	return passValues(subR, values...)
}

// passValues passes multiple values to a pipe. This requires as first argument the template render data.
func passValues(r *renderData, values ...interface{}) (*renderData, error) {
	// Create a new render data value
	data := &renderData{
		Context: r.Context,
	}

	// Get the length of the values
	valuesLen := len(values)

	// If only one value is set, then set it as the root data value.
	if valuesLen == 1 {
		data.Data = values[0]
		return data, nil
	}

	// Values have to be passed with keys.
	if valuesLen%2 != 0 {
		return nil, fmt.Errorf("invalid passValues call: values must have a key")
	}

	// Create a new values map
	valuesMap := make(map[string]interface{}, valuesLen/2)

	for i := 0; i < valuesLen; i += 2 {
		key, ok := values[i].(string)
		if !ok {
			return nil, fmt.Errorf("passValues keys must be of type string")
		}
		valuesMap[key] = values[i+1]
	}

	// Set the map as data interface
	data.Data = valuesMap

	return data, nil
}

func loadJavaScript(c *Context, url string) string {
	// Load the javascript
	c.s.LoadJavaScript(url)

	return ""
}

func loadStyleSheet(c *Context, url string) string {
	// Load the javascript
	c.s.LoadStyleSheet(url)

	return ""
}
