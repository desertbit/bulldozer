/*
 *  Bulldozer Framework
 *  Copyright (C) DesertBit
 */

package auth

import (
	"fmt"
)

//###################//
//### User Struct ###//
//###################//

type User struct {
	u *dbUser
}

func newUser(u *dbUser) *User {
	return &User{
		u: u,
	}
}

func (u *User) ID() string {
	return u.u.ID
}

func (u *User) LoginName() string {
	return u.u.LoginName
}

func (u *User) Name() string {
	return u.u.Name
}

func (u *User) EMail() string {
	return u.u.EMail
}

func (u *User) Enabled() bool {
	return u.u.Enabled
}

func (u *User) LastLogin() int64 {
	return u.u.LastLogin
}

func (u *User) Created() int64 {
	return u.u.Created
}

func (u *User) Groups() []string {
	return u.u.Groups
}

// Update the user data, by retreiving the data from the database.
func (u *User) Update() error {
	// Obtain the user value from the cache or database with the user ID.
	dbUser, err := cacheGetDBUser(u.u.ID)
	if err != nil {
		return err
	} else if dbUser == nil {
		return fmt.Errorf("failed to update user data: user does not exists with ID: '%s'", u.u.ID)
	}

	// Set the new value.
	u.u = dbUser

	return nil
}
