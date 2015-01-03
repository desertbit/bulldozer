/*
 *  Bulldozer Framework
 *  Copyright (C) DesertBit
 */

package template

import (
	"errors"
)

var (
	bulldozerFuncMap FuncMap = FuncMap{
		"plugin":     renderPlugin,
		"eventKey":   createEventAccessKey,
		"passValues": passValues,
		"loadJS":     loadJavaScript,
		"loadStyle":  loadStyleSheet,
	}
)

//##################################//
//### Private template functions ###//
//##################################//

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
		return nil, errors.New("invalid passValues call: values must have a key")
	}

	// Create a new values map
	valuesMap := make(map[string]interface{}, valuesLen/2)

	for i := 0; i < valuesLen; i += 2 {
		key, ok := values[i].(string)
		if !ok {
			return nil, errors.New("passValues keys must be of type string")
		}
		valuesMap[key] = values[i+1]
	}

	// Set the map as data interface
	data.Data = valuesMap

	return data, nil
}

func loadJavaScript(c *Context, url string) error {
	// Load the javascript
	c.s.LoadJavaScript(url)

	return nil
}

func loadStyleSheet(c *Context, url string) error {
	// Load the javascript
	c.s.LoadStyleSheet(url)

	return nil
}
