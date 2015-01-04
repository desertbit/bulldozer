/*
 *  Bulldozer Framework
 *  Copyright (C) DesertBit
 */

package template

import (
	"fmt"
	"code.desertbit.com/bulldozer/bulldozer/log"
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
func parseTemplate(typeStr string, token string, d *parseData) error {
	// Split the token between spaces
	fields := strings.Fields(token)
	fieldsLen := len(fields)

	if fieldsLen == 0 {
		return fmt.Errorf("no template name specified!")
	}

	*d.final += d.leftDelim + "template " + fields[0] + " tmplC " + fields[0] + " $ "
	for i := 1; i < fieldsLen; i++ {
		*d.final += fields[i] + " "
	}
	*d.final += d.rightDelim

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
