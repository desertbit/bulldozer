/*
 *  Bulldozer Framework
 *  Copyright (C) DesertBit
 */

package store

import (
	"code.desertbit.com/bulldozer/bulldozer/template"
)

//##################//
//### Store type ###//
//##################//

type Store struct {
	data *dbStore
}

//##############//
//### Public ###//
//##############//

func Init() error {
	// Initialize the database.
	return initDB()
}

// GetStore returns the current page store of the context.
func GetStore(c *template.Context) (*Store, error) {
	// The page ID is the context's root ID.
	pageID := c.RootID()

	// TODO: Cache stores in memory!

	data, err := dbGetStore(pageID)
	if err != nil {
		return nil, err
	}

	// Create a new store value.
	store := &Store{
		data: data,
	}

	return store, nil
}
