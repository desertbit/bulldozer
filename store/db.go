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
	dbStoreTable     = "store"
	dbTmpStoreTable  = "store_tmp"
	dbLockStoreTable = "store_lock"
)

//#######################//
//### Database Struct ###//
//#######################//

type dbLockData struct {
	ID    string `gorethink:"id"`
	Value string
}

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

	// Create the temporary store table.
	err = db.CreateTableIfNotExists(dbLockStoreTable)
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
		return nil, fmt.Errorf("failed to get database store: ID is empty!")
	}

	var dbTable = dbStoreTable
	useTemporaryData := false

	if len(vars) > 0 && vars[0] {
		dbTable = dbTmpStoreTable
		useTemporaryData = true
	}

	rows, err := r.Table(dbTable).Get(id).Run(db.Session)
	if err != nil {
		return nil, fmt.Errorf("failed to get database store by ID '%s': %v", id, err)
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
		return nil, fmt.Errorf("failed to get database store by ID '%s': %v", id, err)
	}

	return &s, nil
}

// dbUpdateStore inserts or updates the store in the database.
// One variadic boolean can be passed.
// If true, then the store is saved to the temporary table.
func dbUpdateStore(s *dbStore, vars ...bool) error {
	var dbTable = dbStoreTable
	if len(vars) > 0 && vars[0] {
		dbTable = dbTmpStoreTable
	}

	_, err := r.Table(dbTable).Insert(s, r.InsertOpts{
		Conflict: "update",
	}).RunWrite(db.Session)
	if err != nil {
		return err
	}

	return nil
}

// dbGetLocks retrieves all locked IDs from the database.
func dbGetLocks() (l []*dbLockData, err error) {
	rows, err := r.Table(dbLockStoreTable).Run(db.Session)
	if err != nil {
		return nil, fmt.Errorf("failed to get all locks from database: %v", err)
	}

	// Check if nothing was found.
	if rows.IsNil() {
		return nil, nil
	}

	err = rows.All(l)
	if err != nil {
		return nil, fmt.Errorf("failed to get all locks from database: %v", err)
	}

	return
}

func dbIsLocked(id string, value string) (bool, error) {
	rows, err := r.Table(dbLockStoreTable).
		Get(id).Run(db.Session)
	if err != nil {
		return false, err
	}

	// Check if nothing was found.
	if rows.IsNil() {
		return false, nil
	}

	var d dbLockData
	err = rows.One(&d)
	if err != nil {
		return false, fmt.Errorf("failed to get lock state from database: %v", err)
	}

	return d.Value == value, nil
}

func dbIsLockedByAnotherValue(id string, value string) (bool, error) {
	rows, err := r.Table(dbLockStoreTable).
		Get(id).Run(db.Session)
	if err != nil {
		return false, err
	}

	// Check if nothing was found.
	if rows.IsNil() {
		return false, nil
	}

	var d dbLockData
	err = rows.One(&d)
	if err != nil {
		return false, fmt.Errorf("failed to get lock state from database: %v", err)
	}

	return d.Value != value, nil
}

func dbLock(id string, value string) error {
	d := dbLockData{
		ID:    id,
		Value: value,
	}

	_, err := r.Table(dbLockStoreTable).
		Insert(d, r.InsertOpts{Conflict: "update"}).
		RunWrite(db.Session)
	if err != nil {
		return err
	}

	return nil
}

func dbUnlock(id string) error {
	_, err := r.Table(dbLockStoreTable).
		Get(id).Delete().RunWrite(db.Session)
	if err != nil {
		return err
	}

	return nil
}
