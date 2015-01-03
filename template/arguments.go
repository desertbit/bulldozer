/*
 *  Bulldozer Framework
 *  Copyright (C) DesertBit
 */

package template

import (
	"fmt"
	"strings"
)

type Args map[string]string

// getArgs splits an argument string to a map and skips empty spaces delimiters in quotes.
// If the passed string s is emtpy, then a nil map is returned.
func getArgs(s string) (Args, error) {
	// First split the string into a list, but skip delimiters in quotes
	var l []string
	var data []rune
	var dataStr string
	skip := false

	// Trim all empty spaces
	s = strings.TrimSpace(s)

	// If s is empty, then return nil
	if len(s) == 0 {
		return nil, nil
	}

	// Append a delimiter to the end of the string, to ensure
	// that the last element is also added to the list.
	s += " "

	// Split the string
	for _, p := range s {
		if p == '"' {
			skip = !skip
		} else if !skip && p == ' ' {
			dataStr = strings.TrimSpace(string(data))
			if dataStr != "" {
				l = append(l, dataStr)
			}

			data = data[:0]
		} else {
			data = append(data, p)
		}
	}

	// Fill the map
	m := make(Args)

	for _, str := range l {
		// Find the '=' delimiter
		pos := strings.Index(str, "=")

		// If not found, throw and error and exit
		if pos == -1 {
			return nil, fmt.Errorf("argument split: '%s': missing '=' delimiter!", s)
		}

		// Get the key and value
		p1 := str[0:pos]
		p2 := str[pos+1:]

		if p1 == "" || p2 == "" {
			return nil, fmt.Errorf("argument split: '%s': empty data key or value!", s)
		}

		// Add it to the map
		m[p1] = p2
	}

	return m, nil
}
