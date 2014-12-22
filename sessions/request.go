/*
 *  Bulldozer Framework
 *  Copyright (C) DesertBit
 */

package sessions

import (
	"fmt"
)

var (
	requests map[string]RequestFunc = make(map[string]RequestFunc)
)

type RequestFunc func(s *Session, data map[string]string) error

// Request registers the new request function.
// A previously registeres request with the same type will be overridden.
// This function has to be called during initialization!
// This is not thread safe!
func Request(typeStr string, r RequestFunc) error {
	// Check for valid input data
	if len(typeStr) == 0 || r == nil {
		return fmt.Errorf("register request: invalid or empty data passed!")
	}

	// Add the request to the map
	requests[typeStr] = r
	return nil
}
