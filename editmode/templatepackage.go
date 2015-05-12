/*
 *  Bulldozer Framework
 *  Copyright (C) DesertBit
 */

package editmode

import (
	"github.com/desertbit/bulldozer/template"
)

const (
	templatePackageName = "editmode"
)

func init() {
	// Register the editmode template package.
	template.RegisterPackage(templatePackageName, new(templatePackage))
}

//########################//
//### Template Package ###//
//########################//

type templatePackage struct{}

func (p *templatePackage) IsActive(c *template.Context) bool {
	return IsActive(c.Session())
}
