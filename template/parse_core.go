/*
 *  Bulldozer Framework
 *  Copyright (C) DesertBit
 */

package template

import (
	"fmt"
	"strings"
)

func init() {
	// Register the template parse functions
	registerParseFunc("template", parseTemplate)
	registerParseFunc("js", parseJS)
	registerParseFunc("end", parseAddEndTag)
}

//###############//
//### Private ###//
//###############//

// parseTemplate passes the templates context to the template pipeline.
func parseTemplate(token string, d *parseData) error {
	// Split the token between spaces
	fields := strings.Fields(token)
	fieldsLen := len(fields)

	if fieldsLen == 0 {
		return fmt.Errorf("no template name specified!")
	} else if fieldsLen == 1 {
		*d.final += d.leftDelim + "template " + fields[0] + " $" + d.rightDelim
	} else {
		*d.final += d.leftDelim + "template " + fields[0] + " passValues $ "
		for i := 1; i < fieldsLen; i++ {
			*d.final += fields[i] + " "
		}
		*d.final += d.rightDelim
	}

	return nil
}

// Javascript section which is executed as soon as everything is loaded.
// JS Event syntax:
//		{{{js load}}} ... {{{end js}}}
//		{{{js unload}}} ... {{{end js}}}
func parseJS(token string, d *parseData) error {
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

func parseAddEndTag(token string, d *parseData) error {
	if token == "js" {
		// Add the javascript end section
		*d.final += "});</script>"
	} else {
		// Nothing to do. Just add the tag as it is.
		if len(token) > 0 {
			token = " " + token
		}

		*d.final += d.leftDelim + "end" + token + d.rightDelim
	}

	return nil
}
