/*
 *  Bulldozer Framework
 *  Copyright (C) DesertBit
 */

package template

import (
	"code.desertbit.com/bulldozer/bulldozer/log"
	"code.desertbit.com/bulldozer/bulldozer/utils"
	"fmt"
	"strings"
)

func init() {
	// Register the template parse functions
	registerParseFunc("template", parseTemplate)
	registerParseFunc("js", parseJS)
	registerParseFunc("end", parseAddEndTag)
	registerParseFunc("id", parseDomID)
	registerParseFunc("script", parseScript)
	registerParseFunc("style", parseStyle)
	registerParseFunc("must", parseMust)
	registerParseFunc("global", parseGlobal)
}

//###############//
//### Private ###//
//###############//

// parseTemplate passes the templates context to the template pipeline.
// An unique template context id can be set as second argument with id="%ID".
// To pass only one argument to the template render context, just append it.
// Multiple arguments can be passed with key=%Value.
func parseTemplate(typeStr string, token string, d *parseData) error {
	// Split the string into a slice, but skip delimiters in quotes.
	fields, err := utils.Fields(token)
	if err != nil {
		return fmt.Errorf("include template: arguments: '%s': %v", token, err)
	}

	fieldsLen := len(fields)
	argIndex := 1

	if fieldsLen == 0 {
		return fmt.Errorf("no template name specified!")
	}

	// Get the template name.
	templateName := fields[0]

	// Get the ID if present.
	id := `""`
	if fieldsLen >= 2 && strings.HasPrefix(fields[1], "id=") {
		id = strings.TrimPrefix(fields[1], "id=")
		argIndex = 2
	}

	// Prepare the arguments.
	var args string
	if argIndex+1 == fieldsLen {
		args = fields[argIndex]
	} else {
		var str, key, data string
		var pos int
		for ; argIndex < fieldsLen; argIndex++ {
			str = fields[argIndex]

			// Find the '=' delimiter
			pos = strings.Index(str, "=")

			// If not found, return an error.
			if pos == -1 {
				return fmt.Errorf("include template: arguments: '%s': missing '=' delimiter!", str)
			}

			// Get the values.
			key = str[0:pos]
			data = str[pos+1:]

			if len(key) == 0 || len(data) == 0 {
				return fmt.Errorf("include template: arguments: '%s': empty key or value!", str)
			}

			// Add the arguments to the final string.
			args += "\"" + key + "\" " + data + " "
		}
	}

	*d.final += d.leftDelim + "tmplR " + templateName + " " + id + " $ " + args + d.rightDelim

	return nil
}

// Javascript section which is executed as soon as everything is loaded.
// JS Event syntax:
//		{{{js load}}} ... {{{end js}}}
//		{{{js unload}}} ... {{{end js}}}
func parseJS(typeStr string, token string, d *parseData) error {
	// Create a copy of the data string.
	// Otherwise the following method would remove the section...
	src := *d.src

	// Check if the end tag for the javascript section is present
	_, err := getSection("js", &src, d)
	if err != nil {
		return fmt.Errorf("invalid javascript syntax! Missing end tag {{end js}}")
	}

	// Add the javascript starting section
	*d.final += `<script>Bulldozer.core.`

	if token == "load" {
		*d.final += "onJsLoad"
	} else if token == "unload" {
		*d.final += "onJsUnload"
	} else {
		return fmt.Errorf("invalid js argument '%s': valid arguments are: 'load' and 'unload'", token)
	}

	*d.final += `("{{$.Context.DomID}}",function(){`

	return nil
}

func parseAddEndTag(typeStr string, token string, d *parseData) error {
	if token == "js" {
		// Add the javascript end section
		*d.final += "});</script>"
	} else if token == "event" {
		// Add the event end section
		*d.final += "});"
	} else {
		// Nothing to do. Just add the tag as it is.
		if len(token) > 0 {
			token = " " + token
		}

		*d.final += d.leftDelim + "end" + token + d.rightDelim
	}

	return nil
}

// This is the template equivalent function to context.GenDomID(...).
func parseDomID(typeStr string, token string, d *parseData) error {
	// Check the length of the arguments
	if len(token) == 0 {
		return fmt.Errorf("DOM ID: no ID passed to the template id function.\nSyntax: {{id \"$ID\"}}")
	}

	*d.final += `{{$.Context.GenDomID ` + token + `}}`

	return nil
}

func parseScript(typeStr string, token string, d *parseData) error {
	// Check if the javascript url is set.
	if len(token) == 0 {
		return fmt.Errorf("no javascript URL set!\nSyntax: {{script \"$URL\"}}")
	}

	*d.final += `{{loadJS $.Context ` + token + `}}`

	return nil
}

func parseStyle(typeStr string, token string, d *parseData) error {
	// Check if the stylesheet url is set.
	if len(token) == 0 {
		return fmt.Errorf("no stylesheet URL set!\nSyntax: {{style \"$URL\"}}")
	}

	*d.final += `{{loadStyle $.Context ` + token + `}}`

	return nil
}

func parseMust(typeStr string, token string, d *parseData) error {
	if len(token) == 0 {
		return fmt.Errorf("invalid must call: must function name is empty!")
	}

	// Try to obtain the must function
	m, ok := mustFuncs[token]
	if !ok {
		return fmt.Errorf("invalid must call: must function with name '%s' does not exists!", token)
	}

	// Add the must function to the template.
	d.t.mustFuncs = append(d.t.mustFuncs, m)

	// Don't add anything to the template text...
	return nil
}

func parseGlobal(typeStr string, token string, d *parseData) error {
	// Remove leadiing and following quotes.
	token = strings.TrimPrefix(strings.TrimSuffix(token, "\""), "\"")

	if len(token) == 0 {
		return fmt.Errorf("invalid global call: global ID is empty!")
	}

	if len(d.t.globalContextID) > 0 {
		log.L.Warning("template '%s': overwriting global ID: '%s'", d.t.Name(), d.t.globalContextID)
	}

	// Set the template global context ID.
	d.t.globalContextID = token

	// Don't add anything to the template text...
	return nil
}
