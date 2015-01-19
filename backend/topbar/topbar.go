/*
 *  Bulldozer Framework
 *  Copyright (C) DesertBit
 */

package topbar

import (
	"code.desertbit.com/bulldozer/bulldozer/auth"
	"code.desertbit.com/bulldozer/bulldozer/backend"
	"code.desertbit.com/bulldozer/bulldozer/sessions"
)

const (
	// Template names:
	topbarTemplate = "topbar.tmpl"
)

//##############//
//### Public ###//
//##############//

func ExecTopBar(s *sessions.Session) (string, error) {
	// TODO: Only show topbar if the user has the specifc group set.
	if !auth.IsAuth(s) {
		return "", nil
	}

	// Execute the topbar template.
	body, _, _, err := backend.TemplatesStore.Templates.ExecuteTemplateToString(s, topbarTemplate, nil)
	if err != nil {
		return "", err
	}

	return body, nil
}
