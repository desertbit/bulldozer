/*
 *  Bulldozer Framework
 *  Copyright (C) DesertBit
 */

package bulldozer

import (
	"code.desertbit.com/bulldozer/bulldozer/router"
	"code.desertbit.com/bulldozer/bulldozer/sessions"
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

func (b *bulldozerBackend) NavigateToPath(s *sessions.Session, path string) {
	Navigate(s, path)
}

func (b *bulldozerBackend) ReloadPage(s *sessions.Session) {
	ReloadPage(s)
}

func (b *bulldozerBackend) ExecErrorTemplate(s *sessions.Session, errorMessage string, vars ...bool) (int, string, string) {
	return execErrorTemplate(s, errorMessage, vars...)
}

func (b *bulldozerBackend) Route(path string, f func(*sessions.Session, *router.Data) (string, string, error)) {
	Route(path, f)
}
