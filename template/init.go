/*
 *  Bulldozer Framework
 *  Copyright (C) DesertBit
 */

package template

import (
	"code.desertbit.com/bulldozer/bulldozer/sessions"
)

var (
	backend bulldozerBackend
)

//##############//
//### Types ###//
//##############//

type bulldozerBackend interface {
	NavigateToPath(s *sessions.Session, path string)
	ExecErrorTemplate(s *sessions.Session, errorMessage string, vars ...bool) (int, string, string)
}

//##############//
//### Public ###//
//##############//

// Initialize the template package. This is handled by the bulldozer main package.
func Init(b bulldozerBackend) {
	// Set the backend interface.
	backend = b
}
