/*
 *  Bulldozer Framework
 *  Copyright (C) DesertBit
 */

package auth

import (
	db "code.desertbit.com/bulldozer/bulldozer/database"
	r "github.com/dancannon/gorethink"

	"code.desertbit.com/bulldozer/bulldozer/settings"
	"code.desertbit.com/bulldozer/bulldozer/utils"
	"fmt"
	"strings"
	"time"
)

const (
	dbUserTable      = "users"
	dbUserTableIndex = "LoginName"

	maxLength         = 100
	minPasswordLength = 8

	// A simple addition to the goji.Config.PasswordKey.
	// This might be useful, if the password key is stolen from the config,
	// however it isn't the final password encryption key.
	additionalPasswordKey = "bpw"
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
}

//#######################//
//### Private Methods ###//
//#######################//

func initDB() error {
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

	return err
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

	rows, _ := r.Table(dbUserTable).GetAllByIndex(dbUserTableIndex, loginName).Run(db.Session)
	if rows.IsNil() {
		return nil, nil
	}

	var u dbUser
	err := rows.One(&u)
	if err != nil {
		return nil, fmt.Errorf("failed to get database user '%s': %v", loginName, err)
	}

	return &u, nil
}

func dbGetUserByID(id string) (*dbUser, error) {
	if len(id) == 0 {
		return nil, fmt.Errorf("failed to get database user: ID is empty!")
	}

	rows, _ := r.Table(dbUserTable).Get(id).Run(db.Session)
	if rows.IsNil() {
		return nil, nil
	}

	var u dbUser
	err := rows.One(&u)
	if err != nil {
		return nil, fmt.Errorf("failed to get database user by ID '%s': %v", id, err)
	}

	return &u, nil
}

func dbGetUsers() (users []*dbUser, err error) {
	rows, _ := r.Table(dbUserTable).Run(db.Session)
	err = rows.All(users)
	if err != nil {
		return nil, fmt.Errorf("failed to get all database users: %v", err)
	}

	return
}

func dbAddUser(loginName string, name string, email string, password string) (u *dbUser, err error) {
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

	// Create a new user.
	u = &dbUser{
		ID:           id,
		LoginName:    loginName,
		Name:         name,
		EMail:        email,
		PasswordHash: password,
		Enabled:      true,
		LastLogin:    -1,
	}

	// Insert it to the database.
	_, err = r.Table(dbUserTable).Insert(u).RunWrite(db.Session)
	if err != nil {
		return nil, fmt.Errorf("failed to insert new user '%s' to database table: %v", loginName, err)
	}

	return u, nil
}

func dbUpdateLastLogin(u *dbUser) error {
	// Set the last login time
	u.LastLogin = time.Now().Unix()

	_, err := r.Table(dbUserTable).Update(u).RunWrite(db.Session)
	return err
}

func dbChangePassword(u *dbUser, newPassword string) error {
	// Validate input.
	if len(newPassword) < minPasswordLength {
		return fmt.Errorf("failed to change password for user '%s': the new passord is to short", u.LoginName)
	}

	// Hash and encrypt the password.
	u.PasswordHash = hashPassword(newPassword)

	// Update the data in the database
	_, err := r.Table(dbUserTable).Update(u).RunWrite(db.Session)
	return err
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
