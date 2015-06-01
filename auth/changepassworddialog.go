/*
 *  Bulldozer Framework
 *  Copyright (C) DesertBit
 */

package auth

import (
	"encoding/gob"
	"fmt"
	"strings"

	"github.com/desertbit/bulldozer/callback"
	"github.com/desertbit/bulldozer/log"
	"github.com/desertbit/bulldozer/sessions"
	"github.com/desertbit/bulldozer/template"
	"github.com/desertbit/bulldozer/translate"
	"github.com/desertbit/bulldozer/ui/dialog"
	"github.com/desertbit/bulldozer/ui/messagebox"
)

const (
	sessionValueKeyChangePasswordUserID = "budChangePasswordUID"
	sessionValueKeyChangePasswordOpts   = "budChangePasswordOpts"
)

var (
	changePasswordDialog *dialog.Dialog
)

func init() {
	// Register the custom type.
	gob.Register(new(ChangePasswordDialogOpts))

	// Create the dialog and set the options.
	changePasswordDialog = dialog.New().
		SetSize(dialog.SizeSmall).
		SetClosable(false)
}

//##############//
//### Public ###//
//##############//

type ChangePasswordDialogOpts struct {
	// If true, a success message box is shown to the user.
	ShowSuccessMsgBox bool

	// If not empty, the callback specified by the name for
	// the callback package is executed on success.
	CallbackName string
}

// ShowChangePasswordDialog shows a change password dialog for the given user.
// One optional parameter can be passed, defining a callback name for the callback package.
// This callback is executed on success.
func ShowChangePasswordDialog(s *sessions.Session, u *User, opts ...ChangePasswordDialogOpts) error {
	// Create the template render data.
	data := struct {
		Username string
	}{
		Username: u.u.LoginName,
	}

	// Show the dialog
	_, err := changePasswordDialog.Show(s, data)
	if err != nil {
		return fmt.Errorf("failed to show the change password dialog: %v", err)
	}

	// Save the user ID to the session values.
	s.InstanceSet(sessionValueKeyChangePasswordUserID, u.u.ID)

	// Save the options to the session values if present.
	if len(opts) > 0 {
		s.InstanceSet(sessionValueKeyChangePasswordOpts, &opts[0])
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

	// Get the options.
	opts := func() *ChangePasswordDialogOpts {
		// Get the options.
		i, ok := s.InstanceGet(sessionValueKeyChangePasswordOpts)
		if !ok {
			log.L.Error("change password dialog: failed to get options from session: no session value found!")
			return nil
		}

		// Assertion.
		opts, ok := i.(*ChangePasswordDialogOpts)
		if !ok {
			log.L.Error("change password dialog: failed to assert options!")
			return nil
		}

		return opts
	}()
	if opts == nil {
		// Show a messagebox
		messagebox.New().
			SetTitle(tr.S("bud.auth.changePassword.error.changePasswordTitle")).
			SetText(tr.S("bud.auth.changePassword.error.changePassword")).
			SetType(messagebox.TypeAlert).
			Show(s)
		return
	}

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

	// Call the callback if defined.
	if len(opts.CallbackName) > 0 {
		callback.Call(opts.CallbackName, s, newUser(user))
	}

	// Show a success message box if defined to.
	if opts.ShowSuccessMsgBox {
		messagebox.New().
			SetTitle(tr.S("bud.auth.changePassword.success.title")).
			SetText(tr.S("bud.auth.changePassword.success.text")).
			SetType(messagebox.TypeSuccess).
			Show(s)
	}
}

func (e *changePasswordDialogEvents) EventCancel(c *template.Context) {
	// Just remove the unneeded ID again.
	c.Session().InstanceDelete(sessionValueKeyChangePasswordUserID)

	// Close the dialog
	changePasswordDialog.Close(c)
}
