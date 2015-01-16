/*
 *  Bulldozer Framework
 *  Copyright (C) DesertBit
 */

package auth

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
