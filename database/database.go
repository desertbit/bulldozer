/*
 *  Bulldozer Framework
 *  Copyright (C) DesertBit
 */

package database

import (
	r "github.com/dancannon/gorethink"

	"code.desertbit.com/bulldozer/bulldozer/settings"
	"fmt"
)

var (
	Session *r.Session
)

//##############//
//### Public ###//
//##############//

func Connect() (err error) {
	// Create the database address string.
	addr := settings.Settings.DatabaseAddr + ":" + settings.Settings.DatabasePort

	// Connext to the database server.
	Session, err = r.Connect(r.ConnectOpts{
		Address:     addr,
		Database:    settings.Settings.DatabaseName,
		MaxIdle:     settings.Settings.DatabaseMaxIdle,
		MaxActive:   settings.Settings.DatabaseMaxActive,
		IdleTimeout: settings.Settings.DatabaseIdleTimeout,
	})
	if err != nil {
		return fmt.Errorf("failed to connect to database: %v", err)
	}

	return
}

func Close() {
	if Session == nil {
		return
	}

	Session.Close()
}

// UUID creates a new unique ID, which can be used as database access ID.
func UUID() (string, error) {
	// Create a new unique ID.
	r, err := r.UUID().Run(Session)
	if err != nil {
		return "", fmt.Errorf("failed to obtain a new unique ID: %v", err)
	}

	// Get the value.
	var id string
	err = r.One(&id)
	if err != nil {
		return "", fmt.Errorf("failed to obtain a new unique ID: %v", err)
	}

	if len(id) == 0 {
		return "", fmt.Errorf("failed to obtain a new unique ID: %v", err)
	}

	return id, nil
}

func CreateTables(tableNames ...string) (err error) {
	for _, t := range tableNames {
		err = CreateTable(t)
		if err != nil {
			return err
		}
	}

	return nil
}

func CreateTable(tableName string) (err error) {
	// Create the table.
	_, err = r.Db(settings.Settings.DatabaseName).TableCreate(tableName).RunWrite(Session)
	return
}

// CreateTableIfNotExists creates the table if it does not exists
// and calls the function f if passed.
func CreateTableIfNotExists(tableName string, f ...func() error) error {
	// Get a table list.
	res, err := r.Db(settings.Settings.DatabaseName).TableList().Run(Session)
	if err != nil {
		return err
	}

	var tableList []string
	err = res.All(&tableList)
	if err != nil {
		return err
	}

	// Check if the table exists.
	for _, table := range tableList {
		if table == tableName {
			return nil
		}
	}

	// Create the table.
	err = CreateTable(tableName)
	if err != nil {
		return err
	}

	// Call the callback if present
	if len(f) > 0 {
		err = f[0]()
		if err != nil {
			return err
		}
	}

	return nil
}
