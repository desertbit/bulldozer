/*
 *  Bulldozer Framework
 *  Copyright (C) DesertBit
 */

package store

import (
	db "code.desertbit.com/bulldozer/bulldozer/database"
	r "github.com/dancannon/gorethink"

	"fmt"
	"sync"
)

const (
	dbStoreTable = "store"
)

//#######################//
//### Database Struct ###//
//#######################//

type dbStore struct {
	ID    string `gorethink:"id"`
	Data  map[string]*dbStoreData
	mutex sync.Mutex
}

type dbStoreData struct {
	Data  map[string]interface{}
	mutex sync.Mutex
}

//###############//
//### Private ###//
//###############//

func initDB() error {
	// Create the store table.
	err := db.CreateTableIfNotExists(dbStoreTable)
	if err != nil {
		return err
	}

	return nil
}

func dbGetStore(id string) (*dbStore, error) {
	if len(id) == 0 {
		return nil, fmt.Errorf("failed to get database page store: ID is empty!")
	}

	rows, err := r.Table(dbStoreTable).Get(id).Run(db.Session)
	if err != nil {
		return nil, fmt.Errorf("failed to get database page store by ID '%s': %v", id, err)
	}

	// Check if nothing was found.
	if rows.IsNil() {
		return nil, fmt.Errorf("failed to get database page store by ID '%s': no entry was found!", id)
	}

	var s dbStore
	err = rows.One(&s)
	if err != nil {
		return nil, fmt.Errorf("failed to get database page store by ID '%s': %v", id, err)
	}

	return &s, nil
}
