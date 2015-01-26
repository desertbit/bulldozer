/*
 *  Bulldozer Framework
 *  Copyright (C) DesertBit
 */

package store

import (
	db "code.desertbit.com/bulldozer/bulldozer/database"
	r "github.com/dancannon/gorethink"

	"fmt"
)

const (
	dbStoreTable    = "store"
	dbTmpStoreTable = "store_tmp"
)

//#######################//
//### Database Struct ###//
//#######################//

type dbStore struct {
	ID     string `gorethink:"id"`
	Values map[string]*dbStoreData
}

type dbStoreData struct {
	Data interface{}

	// Unexported.
	isDecoded bool
}

func newDBStoreData(data interface{}) *dbStoreData {
	return &dbStoreData{
		Data:      data,
		isDecoded: true,
	}
}

func newDBStore(id string) *dbStore {
	return &dbStore{
		ID: id,
	}
}

func (s *dbStore) createMapIfNil() {
	if s.Values == nil {
		s.Values = make(map[string]*dbStoreData)
	}
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

	// Create the temporary store table.
	err = db.CreateTableIfNotExists(dbTmpStoreTable)
	if err != nil {
		return err
	}

	return nil
}

// dbGetStore retrieves the store for the given ID from the database.
// One variadic boolean can be passed. If true and temporary store data
// exists, then this temporary data is returned instead.
func dbGetStore(id string, vars ...bool) (*dbStore, error) {
	if len(id) == 0 {
		return nil, fmt.Errorf("failed to get database page store: ID is empty!")
	}

	var dbTable = dbStoreTable
	useTemporaryData := false

	if len(vars) > 0 && vars[0] {
		dbTable = dbTmpStoreTable
		useTemporaryData = true
	}

	rows, err := r.Table(dbTable).Get(id).Run(db.Session)
	if err != nil {
		return nil, fmt.Errorf("failed to get database page store by ID '%s': %v", id, err)
	}

	// Check if nothing was found.
	if rows.IsNil() {
		// If nothing was found in the temporary table,
		// then try to get the data from the default table.
		if useTemporaryData {
			return dbGetStore(id)
		}

		return nil, nil
	}

	var s dbStore
	err = rows.One(&s)
	if err != nil {
		return nil, fmt.Errorf("failed to get database page store by ID '%s': %v", id, err)
	}

	return &s, nil
}

// dbInsertStore inserts the store to the database.
// One variadic boolean can be passed.
// If true, then the store is saved to the temporary table.
func dbInsertStore(s *dbStore, vars ...bool) error {
	var dbTable = dbStoreTable
	if len(vars) > 0 && vars[0] {
		dbTable = dbTmpStoreTable
	}

	_, err := r.Table(dbTable).Insert(s).RunWrite(db.Session)
	if err != nil {
		return err
	}

	return nil
}

// dbUpdateStore updates the store in the database.
// One variadic boolean can be passed.
// If true, then the store is saved to the temporary table.
func dbUpdateStore(s *dbStore, vars ...bool) error {
	var dbTable = dbStoreTable
	if len(vars) > 0 && vars[0] {
		dbTable = dbTmpStoreTable
	}

	_, err := r.Table(dbTable).Update(s).RunWrite(db.Session)
	if err != nil {
		return err
	}

	return nil
}
