/*
 *  Bulldozer Framework
 *  Copyright (C) DesertBit
 */

package auth

import (
	tr "code.desertbit.com/bulldozer/bulldozer/translate"

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

func (p *templatePackage) MustIsAuth(a *template.Action, c *template.Context) {
	if !IsAuth(c) {
		a.Error(tr.S("bud.auth.pkg.mustAuthErrorMessage"))
	}
}

func (p *templatePackage) IsAuth(c *template.Context) bool {
	return IsAuth(c)
}

func (p *templatePackage) GetUser(c *template.Context) *User {
	return GetUser(c)
}

func (p *templatePackage) LoginName(c *template.Context) string {
	u := GetUser(c)
	if u == nil {
		return ""
	}

	return u.LoginName()
}

func (p *templatePackage) UserName(c *template.Context) string {
	u := GetUser(c)
	if u == nil {
		return ""
	}

	return u.Name()
}

func (p *templatePackage) Group(c *template.Context, groups ...string) bool {
	// Get the user.
	u := GetUser(c)
	if u == nil {
		return false
	}

	// Perform the actual action.
	return u.IsInGroup(groups...)
}
