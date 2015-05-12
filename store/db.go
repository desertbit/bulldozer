/*
 *  Bulldozer Framework
 *  Copyright (C) DesertBit
 */

package store

import (
	r "github.com/dancannon/gorethink"
	db "github.com/desertbit/bulldozer/database"

	"fmt"
	"time"
)

const (
	dbStoreTable     = "store"
	dbStoreInfoTable = "store_info"
	dbTmpStoreTable  = "store_tmp"
	dbLockStoreTable = "store_lock"
)

func init() {
	db.OnSetup(setupDB)
}

//#######################//
//### Database Struct ###//
//#######################//

type dbStoreInfo struct {
	Store      string `gorethink:"id"`
	LastChange int64
}

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

func setupDB() error {
	// Create the tables.
	err := db.CreateTables(dbStoreTable, dbStoreInfoTable, dbTmpStoreTable, dbLockStoreTable)
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

func dbHasTemporaryChanges() (bool, error) {
	// Get all changes.
	rows, err := r.Table(dbTmpStoreTable).Run(db.Session)
	if err != nil {
		return false, fmt.Errorf("failed to get temporary store state: %v", err)
	}

	return !rows.IsNil(), nil
}

func dbSaveTemporaryChanges() error {
	// TODO: Merge this into one ReQL command if possible.

	// Get all changes.
	rows, err := r.Table(dbTmpStoreTable).Run(db.Session)
	if err != nil {
		return fmt.Errorf("failed to get temporary store changes: %v", err)
	}

	// Return if nothing was found.
	if rows.IsNil() {
		return nil
	}

	// Assert
	var stores []*dbStore
	err = rows.All(&stores)
	if err != nil {
		return fmt.Errorf("failed to get temporary store changes: %v", err)
	}

	// Insert all temporary stores to the production table.
	_, err = r.Table(dbStoreTable).Insert(&stores, r.InsertOpts{
		Conflict: "update",
	}).RunWrite(db.Session)
	if err != nil {
		return fmt.Errorf("failed to save temporary store changes: %v", err)
	}

	// Remove all temporary changes on success.
	_, err = r.Table(dbTmpStoreTable).Delete().Run(db.Session)
	if err != nil {
		return fmt.Errorf("failed to remove temporary store changes: %v", err)
	}

	return nil
}

// dbGetStoreInfo retrieves the store info from the database.
func dbGetStoreInfo(storeID string) (*dbStoreInfo, error) {
	rows, err := r.Table(dbStoreInfoTable).Get(storeID).Run(db.Session)
	if err != nil {
		return nil, fmt.Errorf("failed to get store info from database: %v", err)
	}

	// Check if nothing was found.
	if rows.IsNil() {
		return nil, nil
	}

	var s dbStoreInfo
	err = rows.One(&s)
	if err != nil {
		return nil, fmt.Errorf("failed to get store info from database: %v", err)
	}

	return &s, nil
}

func dbUpdateStoreInfoLastChanged(storeID string) (int64, error) {
	// Create a new timestamp.
	timestamp := time.Now().Unix()

	// Create the store info value.
	info := dbStoreInfo{
		Store:      storeID,
		LastChange: timestamp,
	}

	_, err := r.Table(dbStoreInfoTable).
		Insert(&info, r.InsertOpts{Conflict: "update"}).
		RunWrite(db.Session)
	if err != nil {
		return -1, err
	}

	return timestamp, nil
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

	err = rows.All(&l)
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
