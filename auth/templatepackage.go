/*
 *  Bulldozer Framework
 *  Copyright (C) DesertBit
 */

package auth

import (
	"code.desertbit.com/bulldozer/bulldozer/template"
)

const (
	templatePackageName = "auth"
)

func init() {
	// Register the authentication template package.
	template.RegisterPackage(templatePackageName, new(templatePackage))
}

//########################//
//### Template Package ###//
//########################//

type templatePackage struct{}

// TODO: Also add logout function.

func (p *templatePackage) MustIsAuth(a *template.Action, c *template.Context) {
	if !IsAuth(c.Session()) {
		// TODO: Translate this!
		a.Error("Access denied! Please authenticate to access this page.")
	}
}
