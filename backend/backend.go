/*
 *  Bulldozer Framework
 *  Copyright (C) DesertBit
 */

package backend

import (
	"code.desertbit.com/bulldozer/bulldozer/settings"
	"code.desertbit.com/bulldozer/bulldozer/template/store"
	"code.desertbit.com/bulldozer/bulldozer/utils"
	"fmt"
)

const (
	// Backend template directory name.
	backendTemplatesDir = "backend"
)

var (
	TemplatesStore *store.Store
)

//##############//
//### Public ###//
//##############//

func Init() error {
	// Create a new store and parse it.
	s, err := store.New(utils.AddTrailingSlashToPath(settings.Settings.BulldozerCoreTemplatesPath) + backendTemplatesDir)
	if err != nil {
		return fmt.Errorf("failed to load backend templates: %v", err)
	}

	// Parse the templates.
	s.Parse()

	// Set the templates store.
	TemplatesStore = s

	return nil
}
