/*
 *  Bulldozer Framework
 *  Copyright (C) DesertBit
 */

package auth

import (
	tr "code.desertbit.com/bulldozer/bulldozer/translate"

	"code.desertbit.com/bulldozer/bulldozer/log"
	"code.desertbit.com/bulldozer/bulldozer/router"
	"code.desertbit.com/bulldozer/bulldozer/sessions"
	"code.desertbit.com/bulldozer/bulldozer/settings"
	"code.desertbit.com/bulldozer/bulldozer/template"
	"code.desertbit.com/bulldozer/bulldozer/ui/messagebox"
	"code.desertbit.com/bulldozer/bulldozer/utils"
	"fmt"
	"strings"
)

const (
	randomPasswordLength = 15
)

//#######################//
//### Register Events ###//
//#######################//

type registerEvents struct{}

func (e *registerEvents) EventRegister(c *template.Context, name string, loginName string, email string) {
	// Get the session pointer.
	s := c.Session()

	// Always hide the loading indicator on exit
	defer s.HideLoadingIndicator()

	// Just be sure...
	if settings.Settings.RegistrationDisabled {
		showRegisterErrorMsgBox(s, tr.S("blz.auth.register.errorMsgBoxTextRegistrationDisabled"))
		return
	}

	// Prepare the inputs.
	name = strings.TrimSpace(name)
	loginName = strings.ToLower(strings.TrimSpace(loginName))
	email = strings.TrimSpace(email)

	// Validate...
	if len(name) == 0 || len(loginName) == 0 || len(email) == 0 || !strings.Contains(email, "@") ||
		len(name) > maxLength || len(loginName) > maxLength || len(email) > maxLength {
		showRegisterErrorMsgBox(s, tr.S("blz.auth.register.error.general"))
		return
	}

	// Check if user already exists.
	exist, err := dbUserExists(loginName)
	if err != nil {
		log.L.Error("failed to check if user '%s' exists: %v", loginName, err)
		showRegisterErrorMsgBox(s, tr.S("blz.auth.register.error.generalShort"))
		return
	} else if exist {
		showRegisterErrorMsgBox(s, tr.S("blz.auth.register.error.userAlreadyExists", loginName))
		return
	}

	// Generate a random new password.
	password := utils.RandomString(randomPasswordLength)

	// Add the user to the database
	_, err = dbAddUser(loginName, name, email, password)
	if err != nil {
		log.L.Error("failed to add user '%s' to database: %v", loginName, err)
		showRegisterErrorMsgBox(s, tr.S("blz.auth.register.error.generalShort"))
		return
	}

	// TODO: Send password to the e-mail!
	fmt.Printf("TODO: send password '%s' to e-mail '%s'", password, email)

	// Redirect to the login page.
	backend.NavigateToPath(s, LoginPageUrl)

	// Show a success message box.
	messagebox.New().
		SetTitle(tr.S("blz.auth.register.success.title")).
		SetText(tr.S("blz.auth.register.success.text")).
		SetType(messagebox.TypeSuccess).
		Show(s)
}

//###############//
//### Private ###//
//###############//

func routeRegisterPage(s *sessions.Session, routeData *router.Data) (string, string, error) {
	// If already authenticated, then redirect to the default page.
	if IsAuth(s) {
		backend.NavigateToPath(s, "/")
		return "", "", nil
	}

	// If the registration is disabled, then redirect to the login page.
	if settings.Settings.RegistrationDisabled {
		backend.NavigateToPath(s, LoginPageUrl)
		return "", "", nil
	}

	// Execute the login template.
	o, _, _, err := templates.ExecuteTemplateToString(s, registerTemplate, nil)
	if err != nil {
		return "", "", fmt.Errorf("failed to execute register template: %v", err)
	}

	return o, tr.S("blz.auth.register.pageTitle"), nil
}

func showRegisterErrorMsgBox(s *sessions.Session, msg string) {
	// Show a messagebox
	messagebox.New().
		SetTitle(tr.S("blz.auth.register.errorMsgBoxTitle")).
		SetText(msg).
		SetType(messagebox.TypeAlert).
		Show(s)
}
