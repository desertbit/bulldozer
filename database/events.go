/*
 *  Bulldozer Framework
 *  Copyright (C) DesertBit
 */

package database

import (
	r "github.com/dancannon/gorethink"

	"code.desertbit.com/bulldozer/bulldozer/log"
	"code.desertbit.com/bulldozer/bulldozer/settings"
	"fmt"
)

var (
	setupFuncs         []SetupFunc
	createIndexesFuncs []CreateIndexesFunc
)

//#############//
//### Types ###//
//#############//

type SetupFunc func() error
type CreateIndexesFunc func() error

//##############//
//### Events ###//
//##############//

// OnSetup called if the database should be setup.
func OnSetup(f SetupFunc) {
	setupFuncs = append(setupFuncs, f)
}

// OnCreateIndexes is invoked if the primary and secondary indexes
// should be recreated.
// RethinkDB dump statement:
//   Secondary indexes cannot be exported.
//   You will have to manually recreate them.
func OnCreateIndexes(f CreateIndexesFunc) {
	createIndexesFuncs = append(createIndexesFuncs, f)
}

//##############//
//### Public ###//
//##############//

// Setup the database.
func Setup() (err error) {
	log.L.Info("Setting up database...")

	defer func() {
		if err == nil {
			log.L.Info("Finished with success.")
		} else {
			log.L.Warning("Errors occurred during the setup!")
		}
	}()

	var errStr string

	// Create the database.
	_, err = r.DbCreate(settings.Settings.DatabaseName).RunWrite(Session)
	if err != nil {
		errStr = err.Error()
	}

	// Call the hooks.
	for _, f := range setupFuncs {
		err = f()
		if err != nil {
			if len(errStr) > 0 {
				errStr += "\n"
			}
			errStr += err.Error()
		}
	}

	// Create the indexes.
	if err = CreateIndexes(); err != nil {
		if len(errStr) > 0 {
			errStr += "\n"
		}
		errStr += err.Error()
	}

	if len(errStr) > 0 {
		return fmt.Errorf(errStr)
	}

	return nil
}

// CreateIndexes creates the primary and secondary indexes.
func CreateIndexes() (err error) {
	log.L.Info("Creating database indexes...")

	// Call the hooks.
	var errStr string
	for _, f := range createIndexesFuncs {
		err = f()
		if err != nil {
			if len(errStr) > 0 {
				errStr += "\n"
			}
			errStr += err.Error()
		}
	}

	if len(errStr) > 0 {
		return fmt.Errorf(errStr)
	}

	return nil
}
