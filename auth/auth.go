/*
 *  Bulldozer Framework
 *  Copyright (C) DesertBit
 */

package auth

import (
	"code.desertbit.com/bulldozer/bulldozer/settings"
	"code.desertbit.com/bulldozer/bulldozer/template/store"
)

const (
	authTemplatesDir = "auth"
)

var (
	// Template Stores
	templatesStore *store.Store
)

//##############//
//### Public ###//
//##############//

func Init() error {
	// Create a new store and parse it.
	s, err := store.New(settings.Settings.BulldozerCoreTemplatesPath + "/" + authTemplatesDir)
	if err != nil {
		return err
	}
	s.Parse()

	// Set the templates store.
	templatesStore = s

	return nil
}
