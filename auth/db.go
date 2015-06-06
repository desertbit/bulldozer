/*
 *  Bulldozer Framework
 *  Copyright (C) DesertBit
 */

package auth

import (
	r "github.com/dancannon/gorethink"
	db "github.com/desertbit/bulldozer/database"

	"fmt"
	"github.com/desertbit/bulldozer/log"
	"github.com/desertbit/bulldozer/settings"
	"github.com/desertbit/bulldozer/utils"
	"strings"
	"time"
)

const (
	DBUserTable      = "users"
	DBUserTableIndex = "LoginName"

	maxLength         = 100
	minPasswordLength = 8

	// A simple addition to the goji.Config.PasswordKey.
	// This might be useful, if the password key is stolen from the config,
	// however it isn't the final password encryption key.
	additionalPasswordKey = "bpw"

	cleanupLoopTimeout = 1 * time.Hour // Each one hour.
)

var (
	stopCleanupLoop chan struct{} = make(chan struct{})
)

func init() {
	db.OnSetup(setupDB)
	db.OnCreateIndexes(createIndexes)
}

//########################//
//### Database Structs ###//
//########################//

type dbUser struct {
	ID           string `gorethink:"id"`
	LoginName    string
	Name         string
	EMail        string
	PasswordHash string
	Enabled      bool
	LastLogin    int64
	Created      int64
	Groups       []string
}

//#######################//
//### Private Methods ###//
//#######################//

func setupDB() error {
	// Create the users table.
	err := db.CreateTable(DBUserTable)
	if err != nil {
		return err
	}

	return nil
}

func createIndexes() error {
	// Create a secondary index on the LoginName attribute.
	_, err := r.Table(DBUserTable).IndexCreate(DBUserTableIndex).Run(db.Session)
	if err != nil {
		return err
	}

	// Wait for the index to be ready to use.
	_, err = r.Table(DBUserTable).IndexWait(DBUserTableIndex).Run(db.Session)
	if err != nil {
		return err
	}

	return nil
}

func initDB() {
	// Start the cleanup loop in a new goroutine.
	go cleanupLoop()
}

func releaseDB() {
	// Stop the loop by triggering the quit trigger.
	close(stopCleanupLoop)
}

func dbUserExists(loginName string) (bool, error) {
	u, err := dbGetUser(loginName)
	if err != nil {
		return false, err
	}

	return u != nil, nil
}

func dbGetUser(loginName string) (*dbUser, error) {
	if len(loginName) == 0 {
		return nil, fmt.Errorf("failed to get database user: login name is empty!")
	}

	rows, err := r.Table(DBUserTable).GetAllByIndex(DBUserTableIndex, loginName).Run(db.Session)
	if err != nil {
		return nil, fmt.Errorf("failed to get database user '%s': %v", loginName, err)
	}

	// Check if nothing was found.
	if rows.IsNil() {
		return nil, nil
	}

	var u dbUser
	err = rows.One(&u)
	if err != nil {
		// Check if nothing was found.
		if err == r.ErrEmptyResult {
			return nil, nil
		}

		return nil, fmt.Errorf("failed to get database user '%s': %v", loginName, err)
	}

	return &u, nil
}

func dbGetUserByID(id string) (*dbUser, error) {
	if len(id) == 0 {
		return nil, fmt.Errorf("failed to get database user: ID is empty!")
	}

	rows, err := r.Table(DBUserTable).Get(id).Run(db.Session)
	if err != nil {
		return nil, fmt.Errorf("failed to get database user by ID '%s': %v", id, err)
	}

	// Check if nothing was found.
	if rows.IsNil() {
		return nil, nil
	}

	var u dbUser
	err = rows.One(&u)
	if err != nil {
		// Check if nothing was found.
		if err == r.ErrEmptyResult {
			return nil, nil
		}

		return nil, fmt.Errorf("failed to get database user by ID '%s': %v", id, err)
	}

	return &u, nil
}

func dbAddUser(loginName string, name string, email string, password string, removeOnExpire bool, groups ...string) (u *dbUser, err error) {
	// Prepare the inputs.
	loginName = strings.TrimSpace(loginName)
	name = strings.TrimSpace(name)
	email = strings.TrimSpace(email)

	// Validate the inputs.
	if len(loginName) == 0 || len(loginName) > maxLength ||
		len(name) == 0 || len(name) > maxLength ||
		len(email) == 0 || len(email) > maxLength ||
		len(password) == 0 || len(password) > maxLength {
		if len(loginName) > maxLength {
			loginName = loginName[:maxLength]
		}
		return nil, fmt.Errorf("failed to add user '%s': input string sizes are invalid!", loginName)
	} else if len(password) < minPasswordLength {
		return nil, fmt.Errorf("failed to add user '%s': new passord is to short!", loginName)
	}

	// Check if the user already exists.
	exist, err := dbUserExists(loginName)
	if err != nil {
		return nil, err
	} else if exist {
		return nil, fmt.Errorf("failed to add user: user '%s' already exists!", loginName)
	}

	// Hash and encrypt the password.
	password = hashPassword(password)

	// Create a new unique User ID.
	id, err := db.UUID()
	if err != nil {
		return nil, err
	}

	// Check if the groups exists.
	if len(groups) > 0 {
		for _, g := range groups {
			if !groupExists(g) {
				return nil, fmt.Errorf("failed to add user '%s': the group '%s' does not exists!", loginName, g)
			}
		}
	}

	// Create a new user.
	u = &dbUser{
		ID:           id,
		LoginName:    loginName,
		Name:         name,
		EMail:        email,
		PasswordHash: password,
		Enabled:      true,
		LastLogin:    0,
		Created:      time.Now().Unix(),
		Groups:       groups,
	}

	if removeOnExpire {
		u.LastLogin = -1
	}

	// Insert it to the database.
	_, err = r.Table(DBUserTable).Insert(u).RunWrite(db.Session)
	if err != nil {
		return nil, fmt.Errorf("failed to insert new user '%s' to database table: %v", loginName, err)
	}

	return u, nil
}

func dbUpdateUser(u *dbUser) error {
	// Check if the groups exists.
	// They might have changed.
	if len(u.Groups) > 0 {
		for _, g := range u.Groups {
			if !groupExists(g) {
				return fmt.Errorf("failed to update user '%s': the group '%s' does not exists!", u.LoginName, g)
			}
		}
	}

	_, err := r.Table(DBUserTable).Update(u).RunWrite(db.Session)
	if err != nil {
		return err
	}

	return nil
}

func dbRemoveUsers(ids ...string) error {
	if len(ids) == 0 {
		return nil
	}

	idsI := make([]interface{}, len(ids))
	for i, id := range ids {
		idsI[i] = id
	}

	// Remove the passed users with the given IDs.
	_, err := r.Table(DBUserTable).GetAll(idsI...).
		Delete().RunWrite(db.Session)
	if err != nil {
		return fmt.Errorf("failed to remove users by IDs '%+v': %v", ids, err)
	}

	// Trigger the event.
	for _, id := range ids {
		triggerOnRemovedUser(id)
	}

	return nil
}

func dbUpdateLastLogin(u *dbUser) error {
	// Set the last login time
	u.LastLogin = time.Now().Unix()

	return dbUpdateUser(u)
}

func dbChangePassword(u *dbUser, newPassword string) error {
	// Validate input.
	if len(newPassword) < minPasswordLength {
		return fmt.Errorf("failed to change password for user '%s': the new passord is to short", u.LoginName)
	}

	// Hash and encrypt the password.
	u.PasswordHash = hashPassword(newPassword)

	return dbUpdateUser(u)
}

// TODO: Add an option to retrieve batched users. Don't return all at once!
func dbGetUsersInGroup(group string) ([]*dbUser, error) {
	// Execute the query.
	rows, err := r.Table(DBUserTable).Filter(r.Row.Field("Groups").Contains(group)).Run(db.Session)
	if err != nil {
		return nil, fmt.Errorf("failed to get all database users: %v", err)
	}

	// Get the users from the query.
	var users []*dbUser
	err = rows.All(&users)
	if err != nil {
		return nil, fmt.Errorf("failed to get all database users: %v", err)
	}

	return users, nil
}

//########################//
//### Password methods ###//
//########################//

func hashPassword(password string) string {
	// Hash and encrypt the password
	return utils.EncryptXorBase64(additionalPasswordKey+settings.Settings.PasswordEncryptionKey, utils.Sha256Sum(password))

}

func decryptPasswordHash(hash string) (password string, err error) {
	// Decrypt and generate the temporary SHA256 hash with the session ID and random token.
	password, err = utils.DecryptXorBase64(additionalPasswordKey+settings.Settings.PasswordEncryptionKey, hash)
	return
}

//###############//
//### Cleanup ###//
//###############//

func cleanupLoop() {
	// Create a new ticker
	ticker := time.NewTicker(cleanupLoopTimeout)

	defer func() {
		// Stop the ticker
		ticker.Stop()
	}()

	for {
		select {
		case <-ticker.C:
			// Cleanup some expired database data.
			cleanupExpiredData()
		case <-stopCleanupLoop:
			// Just exit the loop
			return
		}
	}
}

func cleanupExpiredData() {
	// Create the expire timestamp.
	expires := time.Now().Unix() - int64(settings.Settings.RemoveNotConfirmedUsersTimeout)

	// Get all expired users.
	rows, err := r.Table(DBUserTable).Filter(
		r.Row.Field("LastLogin").Eq(-1).
			And(r.Row.Field("Created").Sub(expires).Le(0))).
		Run(db.Session)

	if err != nil {
		log.L.Error("failed to get all expired database users: %v", err)
		return
	}

	// Get the users from the query.
	var users []*dbUser
	err = rows.All(&users)
	if err != nil {
		log.L.Error("failed to get all expired database users: %v", err)
		return
	}

	if len(users) == 0 {
		return
	}

	// Create the slice of IDs.
	ids := make([]string, len(users))
	for i, u := range users {
		ids[i] = u.ID
	}

	// Remove the users.
	err = dbRemoveUsers(ids...)
	if err != nil {
		log.L.Error("failed to remove all expired database users: %v", err)
		return
	}

	return
}
