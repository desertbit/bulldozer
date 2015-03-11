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

type Users []*User

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

// IsSysOp returns a boolean if the user is a system operator.
func (u *User) IsSysOp() bool {
	return u.IsInGroup(GroupSysOp)
}

// IsAdmin returns a boolean if the user is an administrator.
func (u *User) IsAdmin() bool {
	return u.IsInGroup(GroupAdmin)
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

	// Create a copy, if the groups slice of the user is changed by another process...
	userGroups := u.u.Groups

	for _, group := range groups {
		for _, userGroup := range userGroups {
			if group == userGroup {
				return true
			}
		}
	}

	return false
}

// IsInGroups accepts a slice instead of variadic arguments.
// This might be useful if called from templates directly.
func (u *User) IsInGroups(groups []string) bool {
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

//#################################//
//### User manipulation methods ###//
//#################################//

// SetName sets the user's name.
// You have to call the commit method to make this persistent.
func (u *User) SetName(name string) {
	u.u.Name = name
}

// SetEMail sets the user's e-mail.
// You have to call the commit method to make this persistent.
func (u *User) SetEMail(email string) {
	u.u.EMail = email
}

// Commit all changes to the database.
func (u *User) Commit() error {
	err := dbUpdateUser(u.u)
	if err != nil {
		return fmt.Errorf("auth: failed to commit user changes: %v", err)
	}

	return nil
}
