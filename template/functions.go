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
		"tr":        tr.S,
		"plugin":    renderPlugin,
		"eventKey":  createEventAccessKey,
		"tmplC":     templateContext,
		"loadJS":    loadJavaScript,
		"loadStyle": loadStyleSheet,
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
	// Otherwise use the template name as ID.
	if len(id) > 0 {
		id = c.data.ID + "_" + id
	} else {
		id = c.data.ID + "^" + templateName
	}

	// Create the new sub template context.
	c = c.New(t, id)

	// Create a new render data for the sub template.
	data := &renderData{
		Context: c,
		Pkg:     r.Pkg,
	}

	//### Pass the values to the new sub render data.

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
	c.ns.s.LoadJavaScript(url)

	return ""
}

func loadStyleSheet(c *Context, url string) string {
	// Load the javascript
	c.ns.s.LoadStyleSheet(url)

	return ""
}
