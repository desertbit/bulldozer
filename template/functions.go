/*
 *  Bulldozer Framework
 *  Copyright (C) DesertBit
 */

package template

import (
	tr "code.desertbit.com/bulldozer/bulldozer/translate"
	ht "html/template"

	"bytes"
	"fmt"
)

var (
	bulldozerFuncMap FuncMap = FuncMap{
		"tr":        tr.S,
		"plugin":    renderPlugin,
		"eventKey":  createEventAccessKey,
		"tmplR":     renderTemplate,
		"loadJS":    loadJavaScript,
		"loadStyle": loadStyleSheet,
	}
)

//##################################//
//### Private template functions ###//
//##################################//

func renderTemplate(templateName string, id string, r *renderData, values ...interface{}) (ht.HTML, error) {
	// Get the context.
	c := r.Context

	// Get the right sub template.
	t := c.t.Lookup(templateName)
	if t == nil {
		return ht.HTML(""), fmt.Errorf("no template found with name: '%s'", templateName)
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

	// Create a new render data interface.
	var data interface{}

	// Get the length of the values
	valuesLen := len(values)

	// If only one value is set, then set it as the root data value.
	if valuesLen == 1 {
		data = values[0]
	} else {
		// Values have to be passed with keys.
		if valuesLen%2 != 0 {
			return ht.HTML(""), fmt.Errorf("invalid template call: values must have a key")
		}

		// Create a new values map
		valuesMap := make(map[string]interface{}, valuesLen/2)

		for i := 0; i < valuesLen; i += 2 {
			key, ok := values[i].(string)
			if !ok {
				return ht.HTML(""), fmt.Errorf("invalid template call: keys must be of type string")
			}
			valuesMap[key] = values[i+1]
		}

		// Set the map as data interface
		data = valuesMap
	}

	// Execute the sub template context.
	var b bytes.Buffer
	err := ExecuteContext(c, &b, data)
	if err != nil {
		return ht.HTML(""), fmt.Errorf("failed to render sub template '%s': %v", templateName, err)
	}

	return ht.HTML(b.String()), nil
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
