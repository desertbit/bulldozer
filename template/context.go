/*
 *  Bulldozer Framework
 *  Copyright (C) DesertBit
 */

package template

import (
	"code.desertbit.com/bulldozer/bulldozer/sessions"
)

//######################//
//### Context struct ###//
//######################//

type Context struct {
	//Private
	s *sessions.Session
	t *Template
}

// Session resturns the current context session
func (c *Context) Session() *sessions.Session {
	return c.s
}

// Template returns the current context template
func (c *Context) Template() *Template {
	return c.t
}
