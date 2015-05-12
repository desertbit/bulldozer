/*
 *  Bulldozer Framework
 *  Copyright (C) DesertBit
 */

package template

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/desertbit/bulldozer/log"
	"github.com/desertbit/bulldozer/sessions"
	"github.com/desertbit/bulldozer/utils"
)

func init() {
	// Register the template parse function.
	registerParseFunc("event", parseEvent)
}

//##############//
//### Public ###//
//##############//

// TriggerGlobalEvent triggers the global event on the client side defined with the template event syntax.
func TriggerGlobalEvent(s *sessions.Session, eventName string, params ...interface{}) {
	cmd := `Bulldozer.core.emitGlobalServerEvent('` + utils.EscapeJS(eventName) + `'`

	// Append all the parameters
	for i, param := range params {
		// type assertion
		switch v := param.(type) {
		case int:
			cmd += "," + strconv.Itoa(v)
		case bool:
			cmd += "," + strconv.FormatBool(v)
		case string:
			cmd += ",'" + utils.EscapeJS(v) + "'"
		default:
			log.L.Error("context: trigger global event: invalid type of function event parameter: %v : parameters: %v", i+1, params)
			return
		}
	}

	cmd += ");"

	// Send the command to the client
	s.SendCommand(cmd)
}

//###############//
//### Private ###//
//###############//

// Event section which is triggered from the server side.
// Server event syntax:
//    {{event FuncName($arg1, $arg2, $arg3 ...)}} ... {{end event}}
// Global server event syntax:
//    {{event global FuncName($arg1, $arg2, $arg3 ...)}} ... {{end event}}
func parseEvent(typeStr string, token string, d *parseData) error {
	// Create a copy of the data string.
	// Otherwise the following method would remove the section...
	src := *d.src

	// Check if the end tag for the event section is present
	_, err := getSection("event", &src, d)
	if err != nil {
		return fmt.Errorf("invalid event syntax! Missing end tag {{end event}}")
	}

	// Try to find the '(' symbol
	pos := strings.Index(token, "(")

	// If not found, throw and error and exit
	if pos == -1 {
		return fmt.Errorf("invalid event syntax! Missing event function bracket '('!")
	}

	// Extract the function name and remove it from the original string
	funcName := strings.TrimSpace(token[0:pos])
	token = strings.TrimSpace(token[pos+1:])

	// Split the function name
	split := strings.Fields(funcName)

	// Check if the lenght of the slice is valid
	l := len(split)
	if l <= 0 || l > 2 {
		return fmt.Errorf("invalid event syntax! Invalid event function name: '%s'", funcName)
	}

	// If two parameters are passed, then check if the global parameter is set
	global := false
	if l == 2 {
		if split[0] == "global" {
			global = true
			funcName = split[1]
		} else {
			return fmt.Errorf("invalid event syntax! Invalid event function name: '%s'", funcName)
		}
	}

	// Check if the function name is valid
	if funcName == "" {
		return fmt.Errorf("invalid event syntax! No event function name defined!")
	}

	// The last symbol has to be the ending bracket
	if token[len(token)-1] != ')' {
		return fmt.Errorf("invalid event syntax! Missing event function ending bracket ')'!")
	}

	// Remove the last ending bracket and trim spaces
	token = strings.TrimSpace(token[:len(token)-1])

	// Check if the event is global and create the javascript call
	*d.final += `Bulldozer.core.`

	if global {
		*d.final += `addGlobalServerEvent(`
	} else {
		*d.final += `addServerEvent("{{$.Context.DomID}}",`
	}

	*d.final += "'" + utils.EscapeJS(funcName) + "',function(" + token + "){"

	return nil
}
