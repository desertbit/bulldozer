/*
 *  Bulldozer Framework
 *  Copyright (C) DesertBit
 */

package auth

import (
	tr "code.desertbit.com/bulldozer/bulldozer/translate"

	"code.desertbit.com/bulldozer/bulldozer/log"
	"code.desertbit.com/bulldozer/bulldozer/mux"
	"code.desertbit.com/bulldozer/bulldozer/sessions"
	"code.desertbit.com/bulldozer/bulldozer/settings"
	"code.desertbit.com/bulldozer/bulldozer/template"
	"code.desertbit.com/bulldozer/bulldozer/templates"
	"code.desertbit.com/bulldozer/bulldozer/ui/messagebox"
	"code.desertbit.com/bulldozer/bulldozer/utils"
	"code.desertbit.com/bulldozer/bulldozer/utils/mail"

	"fmt"
	"strings"
	"time"
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
		showRegisterErrorMsgBox(s, tr.S("bud.auth.register.errorMsgBoxTextRegistrationDisabled"))
		return
	}

	// Prepare the inputs.
	name = strings.TrimSpace(name)
	loginName = strings.ToLower(strings.TrimSpace(loginName))
	email = strings.TrimSpace(email)

	// Validate...
	if len(name) == 0 || len(loginName) == 0 || len(email) == 0 || !strings.Contains(email, "@") ||
		len(name) > maxLength || len(loginName) > maxLength || len(email) > maxLength {
		showRegisterErrorMsgBox(s, tr.S("bud.auth.register.error.general"))
		return
	}

	// Check if user already exists.
	exist, err := dbUserExists(loginName)
	if err != nil {
		log.L.Error("failed to check if user '%s' exists: %v", loginName, err)
		showRegisterErrorMsgBox(s, tr.S("bud.auth.register.error.generalShort"))
		return
	} else if exist {
		showRegisterErrorMsgBox(s, tr.S("bud.auth.register.error.userAlreadyExists", loginName))
		return
	}

	// Generate a random new password.
	password := utils.RandomString(randomPasswordLength)

	// Add the user to the database
	u, err := dbAddUser(loginName, name, email, password)
	if err != nil {
		log.L.Error("failed to add user '%s' to database: %v", loginName, err)
		showRegisterErrorMsgBox(s, tr.S("bud.auth.register.error.generalShort"))
		return
	}

	// Send the registration e-mail.
	err = sendRegistrationEMail(u, password)
	if err != nil {
		log.L.Error("%v", err)
		showRegisterErrorMsgBox(s, tr.S("bud.auth.register.error.generalShort"))
		return
	}

	// Redirect to the login page.
	NavigateToLoginPage(s)

	// Just timeout for a short period, because the navigation call is
	// process in a separate goroutine. Otherwise the messagebox
	// would be shown before the page is changed.
	time.Sleep(350 * time.Millisecond)

	// Show a success message box.
	messagebox.New().
		SetTitle(tr.S("bud.auth.register.success.title")).
		SetText(tr.S("bud.auth.register.success.text")).
		SetType(messagebox.TypeSuccess).
		Show(s)
}

//###############//
//### Private ###//
//###############//

func routeRegisterPage(s *sessions.Session, req *mux.Request) {
	// If already authenticated, then redirect to the default page.
	if IsAuth(s) {
		s.NavigateHome()
		return
	}

	// If the registration is disabled, then redirect to the login page.
	if settings.Settings.RegistrationDisabled {
		NavigateToLoginPage(s)
		return
	}

	// Execute the login template.
	o, _, _, err := templates.Templates.ExecuteTemplateToString(s, registerTemplate)
	if err != nil {
		req.Error(fmt.Errorf("failed to execute register template: %v", err))
		return
	}

	// Set the body and title
	req.Body = o
	req.Title = tr.S("bud.auth.register.pageTitle")
	return
}

func showRegisterErrorMsgBox(s *sessions.Session, msg string) {
	// Show a messagebox
	messagebox.New().
		SetTitle(tr.S("bud.auth.register.errorMsgBoxTitle")).
		SetText(msg).
		SetType(messagebox.TypeAlert).
		Show(s)
}

func sendRegistrationEMail(u *dbUser, password string) error {
	// Create a new mail message.
	m := mail.Message{
		To:      []string{u.EMail},
		Subject: "Ihre Registrierung beim Gesundheitnetz",
	}

	// Create the login url.
	loginURL := settings.Settings.SiteUrl + LoginPageUrl

	// TODO: Translate this!
	m.Body = "Sie haben sich erfolgreich beim Ganzheitichen Gesundheitsnetz registriert." +
		"<br>Bitte melden Sie sich unter folgender Addresse an: <a href=\"" + loginURL + "\">" + loginURL + "</a>" +
		"<br><br>Ihr generiertes Passwort ist: " + password

	// Send the e-mail
	err := mail.Send(&m)
	if err != nil {
		return fmt.Errorf("failed to send registration e-mail: %v", err)
	}

	return nil
}
