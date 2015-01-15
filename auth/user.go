/*
 *  Bulldozer Framework
 *  Copyright (C) DesertBit
 */

package auth

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
