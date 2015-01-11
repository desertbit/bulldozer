/*
 *  Bulldozer Framework
 *  Copyright (C) DesertBit
 */

package bulldozer

import (
	"code.desertbit.com/bulldozer/bulldozer/global"
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

func (b *bulldozerBackend) ExecErrorTemplate(s *sessions.Session, errorMessage string, vars ...bool) (int, string, string) {
	return global.ExecErrorTemplate(s, errorMessage, vars...)
}
