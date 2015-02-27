/*
 *  Bulldozer Framework
 *  Copyright (C) DesertBit
 */

package auth

import (
	"code.desertbit.com/bulldozer/bulldozer/callback"
	"code.desertbit.com/bulldozer/bulldozer/log"
	"code.desertbit.com/bulldozer/bulldozer/sessions"
	"code.desertbit.com/bulldozer/bulldozer/template"
	"code.desertbit.com/bulldozer/bulldozer/translate"
	"code.desertbit.com/bulldozer/bulldozer/ui/dialog"
	"code.desertbit.com/bulldozer/bulldozer/ui/messagebox"
	"fmt"
	"strings"
)

const (
	sessionValueKeyChangePasswordUserID   = "budChangePasswordUID"
	sessionValueKeyChangePasswordCallback = "budChangePasswordCB"
)

var (
	changePasswordDialog *dialog.Dialog
)

func init() {
	// Create the dialog and set the options.
	changePasswordDialog = dialog.New().
		SetSize(dialog.SizeSmall).
		SetClosable(false)
}

//##############//
//### Public ###//
//##############//

// showChangePasswordDialog shows a change password dialog for the given user.
// One optional parameter can be passed, defining a callback name for the callback package.
// This callback is executed on success.
func showChangePasswordDialog(s *sessions.Session, u *dbUser, vars ...string) error {
	// Create the template render data.
	data := struct {
		Username string
	}{
		Username: u.LoginName,
	}

	// Show the dialog
	_, err := changePasswordDialog.Show(s, data)
	if err != nil {
		return fmt.Errorf("failed to show the change password dialog: %v", err)
	}

	// Save the user ID to the session values.
	s.InstanceSet(sessionValueKeyChangePasswordUserID, u.ID)

	// Save the callback name to the session values if present.
	if len(vars) > 0 {
		s.InstanceSet(sessionValueKeyChangePasswordCallback, vars[0])
	}

	return nil
}

//####################//
//### Login Events ###//
//####################//

type changePasswordDialogEvents struct{}

func (e *changePasswordDialogEvents) EventSubmit(c *template.Context, newPassword string) {
	// Get the session pointer.
	s := c.Session()

	// Get the user value.
	user := func() *dbUser {
		// Get the user ID.
		i, ok := s.InstanceGet(sessionValueKeyChangePasswordUserID)
		if !ok {
			log.L.Error("change password dialog: failed to get session user ID: no session value found!")
			return nil
		}

		// Assertion.
		userID, ok := i.(string)
		if !ok {
			log.L.Error("change password dialog: failed to assert user ID to string!")
			return nil
		}

		u, err := dbGetUserByID(userID)
		if err != nil {
			log.L.Error("change password dialog: failed to get session user by ID: '%s': %v", userID, err)
			return nil
		} else if u == nil {
			log.L.Error("change password dialog: failed to get session user by ID: '%s': user is nil!", userID)
			return nil
		}

		return u
	}()
	if user == nil {
		// Show a messagebox
		messagebox.New().
			SetTitle(tr.S("bud.auth.changePassword.error.changePasswordTitle")).
			SetText(tr.S("bud.auth.changePassword.error.changePassword")).
			SetType(messagebox.TypeAlert).
			Show(s)
		return
	}

	// Prepare and validate the password.
	if strings.TrimSpace(newPassword) == "" || len(newPassword) < minPasswordLength {
		// Show a messagebox
		messagebox.New().
			SetTitle(tr.S("bud.auth.changePassword.error.shortPasswordTitle")).
			SetText(tr.S("bud.auth.changePassword.error.shortPassword")).
			SetType(messagebox.TypeWarning).
			Show(s)
		return
	}

	// Change the password
	err := dbChangePassword(user, newPassword)
	if err != nil {
		// Log the error
		log.L.Error("failed to change user password: %v", err)

		// Show a messagebox
		messagebox.New().
			SetTitle(tr.S("bud.auth.changePassword.error.changePasswordTitle")).
			SetText(tr.S("bud.auth.changePassword.error.changePassword")).
			SetType(messagebox.TypeAlert).
			Show(s)
		return
	}

	// Just remove the unneeded ID again.
	s.InstanceDelete(sessionValueKeyChangePasswordUserID)

	// Close the dialog
	changePasswordDialog.Close(c)

	// Get and call the callback if defined.
	i, ok := s.InstancePull(sessionValueKeyChangePasswordCallback)
	if ok {
		name, ok := i.(string)
		if ok {
			callback.Call(name, s, user)
		}
	}
}

func (e *changePasswordDialogEvents) EventCancel(c *template.Context) {
	// Just remove the unneeded ID again.
	c.Session().InstanceDelete(sessionValueKeyChangePasswordUserID)

	// Close the dialog
	changePasswordDialog.Close(c)
}
