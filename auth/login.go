/*
 *  Bulldozer Framework
 *  Copyright (C) DesertBit
 */

package auth

import (
	"code.desertbit.com/bulldozer/bulldozer/router"
	"code.desertbit.com/bulldozer/bulldozer/sessions"
	"code.desertbit.com/bulldozer/bulldozer/tr"
	"fmt"
)

//####################//
//### Login Events ###//
//####################//

type loginEvents struct{}

//###############//
//### Private ###//
//###############//

// TODO: Login Page was not automatically reloaded by the store!!!!!!!!!!!!!!!!!!!!!!!!

func routeLoginPage(s *sessions.Session, routeData *router.Data) (string, string, error) {
	// TODO: Check if auhenticated and if already redirect...

	// Create the template render data.
	data := struct{}{}

	// Execute the login template.
	o, _, _, err := templatesStore.Templates.ExecuteTemplateToString(s, loginTemplate, data)
	if err != nil {
		return "", "", fmt.Errorf("failed to execute login template: %v", err)
	}

	return o, tr.S("blz.auth.login.pageTitle"), nil
}
