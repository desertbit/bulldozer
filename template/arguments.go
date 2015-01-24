/*
 *  Bulldozer Framework
 *  Copyright (C) DesertBit
 */

package template

import (
	"code.desertbit.com/bulldozer/bulldozer/utils"
	"fmt"
	"strings"
)

type Args map[string]string

// getArgs splits an argument string to a map and skips empty spaces delimiters in quotes.
// If the passed string s is emtpy, then a nil map is returned.
func getArgs(s string) (Args, error) {
	// First split the string into a list, but skip delimiters in quotes
	l, err := utils.Fields(s)
	if err != nil {
		return nil, fmt.Errorf("argument split: '%s': %v", s, err)
	} else if l == nil {
		return nil, nil
	}

	// Fill the map
	m := make(Args)

	for _, str := range l {
		// Find the '=' delimiter
		pos := strings.Index(str, "=")

		// If not found, throw an error and exit
		if pos == -1 {
			return nil, fmt.Errorf("argument split: '%s': missing '=' delimiter!", s)
		}

		// Get the key and value
		p1 := str[0:pos]
		p2 := str[pos+1:]

		// Check if quotes are present.
		l := len(p2)
		if l < 2 || p2[0] != '"' || p2[l-1] != '"' {
			return nil, fmt.Errorf("argument split: '%s': quotes are missing around the data value!", s)
		}

		// Remove the quotes.
		p2 = p2[1 : l-1]

		if len(p1) == 0 || len(p2) == 0 {
			return nil, fmt.Errorf("argument split: '%s': empty data key or value!", s)
		}

		// Add it to the map
		m[p1] = p2
	}

	return m, nil
}
