/*
 *  Bulldozer Framework
 *  Copyright (C) DesertBit
 */

package store

import (
	"github.com/desertbit/bulldozer/template"
)

const (
	templatePackageName = "store"
)

func init() {
	// Register the authentication template package.
	template.RegisterPackage(templatePackageName, new(templatePackage))
}

//########################//
//### Template Package ###//
//########################//

type templatePackage struct{}

func (p *templatePackage) IsLocked(c *template.Context) bool {
	return IsLocked(c)
}

func (p *templatePackage) IsBlocked(c *template.Context) bool {
	return IsLockedByAnotherSession(c)
}
