/*
 *  Bulldozer Framework
 *  Copyright (C) DesertBit
 */

package auth

import (
	db "code.desertbit.com/bulldozer/bulldozer/database"
	r "github.com/dancannon/gorethink"

	"code.desertbit.com/bulldozer/bulldozer/log"
	"code.desertbit.com/bulldozer/bulldozer/settings"
	"code.desertbit.com/bulldozer/bulldozer/utils"
	"fmt"
	"strings"
	"time"
)

const (
	dbUserTable      = "users"
	dbUserTableIndex = "LoginName"

	dbGroupTable = "groups"

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

type dbGroup struct {
	Name        string `gorethink:"id"`
	Description string
}

//#######################//
//### Private Methods ###//
//#######################//

func initDB() error {
	// Create the users table.
	err := db.CreateTableIfNotExists(dbUserTable, func() {
		// Create a secondary index on the LoginName attribute.
		_, err := r.Table(dbUserTable).IndexCreate(dbUserTableIndex).Run(db.Session)
		if err != nil {
			panic(err)
		}

		// Wait for the index to be ready to use.
		_, err = r.Table(dbUserTable).IndexWait(dbUserTableIndex).Run(db.Session)
		if err != nil {
			panic(err)
		}
	})
	if err != nil {
		return err
	}

	// Create the groups table.
	err = db.CreateTableIfNotExists(dbGroupTable)
	if err != nil {
		return err
	}

	// Start the cleanup loop in a new goroutine.
	go cleanupLoop()

	return nil
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

	rows, err := r.Table(dbUserTable).GetAllByIndex(dbUserTableIndex, loginName).Run(db.Session)
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
		return nil, fmt.Errorf("failed to get database user '%s': %v", loginName, err)
	}

	return &u, nil
}

func dbGetUserByID(id string) (*dbUser, error) {
	if len(id) == 0 {
		return nil, fmt.Errorf("failed to get database user: ID is empty!")
	}

	rows, err := r.Table(dbUserTable).Get(id).Run(db.Session)
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
		return nil, fmt.Errorf("failed to get database user by ID '%s': %v", id, err)
	}

	return &u, nil
}

// TODO: Add an option to retrieve batched users. Don't return all at once!
func dbGetUsers() ([]*dbUser, error) {
	rows, err := r.Table(dbUserTable).Run(db.Session)
	if err != nil {
		return nil, fmt.Errorf("failed to get all database users: %v", err)
	}

	var users []*dbUser
	err = rows.All(&users)
	if err != nil {
		return nil, fmt.Errorf("failed to get all database users: %v", err)
	}

	return users, nil
}

func dbAddUser(loginName string, name string, email string, password string, groups ...string) (u *dbUser, err error) {
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
		exists, err := dbGroupsExists(groups)
		if err != nil {
			return nil, err
		} else if !exists {
			return nil, fmt.Errorf("failed to add user '%s': one of the groups '%v' does not exists!", loginName, groups)
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
		LastLogin:    -1,
		Created:      time.Now().Unix(),
		Groups:       groups,
	}

	// Insert it to the database.
	_, err = r.Table(dbUserTable).Insert(u).RunWrite(db.Session)
	if err != nil {
		return nil, fmt.Errorf("failed to insert new user '%s' to database table: %v", loginName, err)
	}

	return u, nil
}

func dbUpdateUser(u *dbUser) error {
	_, err := r.Table(dbUserTable).Update(u).RunWrite(db.Session)
	if err != nil {
		return err
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

func dbAddGroup(name string, description string) (*dbGroup, error) {
	if len(name) == 0 {
		return nil, fmt.Errorf("failed to add group: group name is empty!")
	}

	// Check if the group already exists.
	exists, err := dbGroupExists(name)
	if err != nil {
		return nil, fmt.Errorf("failed to add group '%s': %v", name, err)
	} else if exists {
		return nil, fmt.Errorf("failed to add group '%s': group already exists!", name)
	}

	// Create a new group value.
	g := &dbGroup{
		Name:        name,
		Description: description,
	}

	// Insert it to the database.
	_, err = r.Table(dbGroupTable).Insert(g).RunWrite(db.Session)
	if err != nil {
		return nil, fmt.Errorf("failed to insert new group '%s' to database table: %v", name, err)
	}

	return g, nil
}

func dbGroupExists(name string) (bool, error) {
	g, err := dbGetGroup(name)
	if err != nil {
		return false, err
	}

	return g != nil, nil
}

func dbGroupsExists(names []string) (bool, error) {
	groups, err := dbGetGroups()
	if err != nil {
		return false, err
	}

	var found bool
	for _, name := range names {
		found = false
		for _, group := range groups {
			if name == group.Name {
				found = true
				break
			}
		}

		if !found {
			return false, nil
		}
	}

	return true, nil
}

func dbGetGroup(name string) (*dbGroup, error) {
	rows, err := r.Table(dbGroupTable).Get(name).Run(db.Session)
	if err != nil {
		return nil, fmt.Errorf("failed to get database group '%s': %v", name, err)
	}

	// Check if nothing was found.
	if rows.IsNil() {
		return nil, nil
	}

	var g dbGroup
	err = rows.One(&g)
	if err != nil {
		return nil, fmt.Errorf("failed to get database group '%s': %v", name, err)
	}

	return &g, nil
}

func dbGetGroups() ([]*dbGroup, error) {
	rows, err := r.Table(dbGroupTable).Run(db.Session)
	if err != nil {
		return nil, fmt.Errorf("failed to get all database groups: %v", err)
	}

	var groups []*dbGroup
	err = rows.All(&groups)
	if err != nil {
		return nil, fmt.Errorf("failed to get all database groups: %v", err)
	}

	return groups, nil
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

	// Remove all expired users.
	_, err := r.Table(dbUserTable).Filter(
		r.Row.Field("LastLogin").Eq(-1).
			And(r.Row.Field("Created").Sub(expires).Le(0))).
		Delete().RunWrite(db.Session)

	if err != nil {
		log.L.Error("failed to remove expired database users: %v", err)
		return
	}
}
