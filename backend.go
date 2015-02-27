/*
 *  Bulldozer Framework
 *  Copyright (C) DesertBit
 */

package bulldozer

import (
	"code.desertbit.com/bulldozer/bulldozer/router"
	"code.desertbit.com/bulldozer/bulldozer/sessions"
	"code.desertbit.com/bulldozer/bulldozer/template"
	"io/ioutil"
)

var (
	backend *bulldozerBackend = newBulldozerBackend()
)

//#################################################//
//### Private bulldozer backend for sub packages ###//
//##################################################//

type bulldozerBackend struct{}

func newBulldozerBackend() *bulldozerBackend {
	return &bulldozerBackend{}
}

func (b *bulldozerBackend) ExecErrorTemplate(s *sessions.Session, errorMessage string, vars ...bool) (int, string, string) {
	return execErrorTemplate(s, errorMessage, vars...)
}

func (b *bulldozerBackend) Route(path string, f func(*sessions.Session, *router.Data) (string, string, error)) {
	Route(path, f)
}

func (b *bulldozerBackend) RoutePage(path string, pageTitle string, pageTemplate string, UID string) {
	RoutePage(path, pageTitle, pageTemplate, UID)
}

func (b *bulldozerBackend) ParsePageTemplate(templateName string, path string) (*template.Template, error) {
	// Read the file.
	buf, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}

	// Parse the template.
	return Templates.New(templateName).Parse(string(buf))
}
