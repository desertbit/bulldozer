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

// IsInGroup returns true if the user is in one of the passed groups.
// True is returned if no groups are passed.
func (u *User) IsInGroup(groups ...string) bool {
	if len(groups) == 0 {
		return true
	}

	userGroups := u.u.Groups

	var found bool
	for _, group := range groups {
		found = false
		for _, userGroup := range userGroups {
			if group == userGroup {
				found = true
				break
			}
		}

		if !found {
			return false
		}
	}

	return true
}

// IsInGroupSlice accepts a slice instead of variadic arguments.
// This might be useful if called from templates directly.
func (u *User) IsInGroupSlice(groups []string) bool {
	return u.IsInGroup(groups...)
}

// Update the user data, by retreiving the data from the database.
func (u *User) Update() error {
	// Obtain the user value from the database with the user ID.
	dbUser, err := dbGetUserByID(u.u.ID)
	if err != nil {
		return err
	} else if dbUser == nil {
		return fmt.Errorf("failed to update user data: user does not exists with ID: '%s'", u.u.ID)
	}

	// Set the new value.
	u.u = dbUser

	return nil
}
