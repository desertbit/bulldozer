/*
 *  Bulldozer Framework
 *  Copyright (C) DesertBit
 */

package database

import (
	"code.desertbit.com/bulldozer/bulldozer/settings"
	r "github.com/dancannon/gorethink"
)

var (
	session *r.Session
)

//##############//
//### Public ###//
//##############//

func Connect() (err error) {
	// Connext to the database server.
	session, err = r.Connect(r.ConnectOpts{
		Address:     settings.Settings.DatabaseAddress,
		Database:    settings.Settings.DatabaseName,
		MaxIdle:     settings.Settings.DatabaseMaxIdle,
		MaxActive:   settings.Settings.DatabaseMaxActive,
		IdleTimeout: settings.Settings.DatabaseIdleTimeout,
	})
	if err != nil {
		return
	}

	return
}

func Close() {
	if session == nil {
		return
	}

	session.Close()
}
