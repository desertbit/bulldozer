/*
 *  Bulldozer Framework
 *  Copyright (C) DesertBit
 */

package auth

import (
	tr "code.desertbit.com/bulldozer/bulldozer/translate"

	"code.desertbit.com/bulldozer/bulldozer/callback"
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
	passwordTokenLength          = 15
	sessionValueKeyPasswordToken = "blzAuthPassTok"

	finishLoginCallbackName = "blzAuthFinishLogin"
)

func init() {
	// Register the callback.
	callback.Register(finishLoginCallbackName, finishLogin)
}

//####################//
//### Login Events ###//
//####################//

type loginEvents struct{}

func (e *loginEvents) EventLogin(c *template.Context, loginName string, passwordHash string) {
	// Get the session pointer.
	s := c.Session()

	// Always hide the loading indicator on return.
	defer s.HideLoadingIndicator()

	// Get the password login random token.
	passwordTokenI, ok := s.InstanceGet(sessionValueKeyPasswordToken)
	if !ok {
		log.L.Warning("failed to obtain password token from session store for session with remote address: '%s'", s.RemoteAddr())
		showLoginErrorMsgBox(s)
		return
	}
	passwordToken, ok := passwordTokenI.(string)
	if !ok {
		log.L.Warning("failed to cast password token for session with remote address: '%s'", s.RemoteAddr())
		showLoginErrorMsgBox(s)
		return
	}

	// Trim the login name.
	loginName = strings.ToLower(strings.TrimSpace(loginName))

	// Check if inputs are valid.
	if len(loginName) == 0 || len(passwordHash) == 0 {
		showLoginErrorMsgBox(s)
		return
	}

	// Try to get the user.
	u, err := dbGetUser(loginName)
	if err != nil {
		log.L.Error("%v", err)
		showLoginErrorMsgBox(s)
		return
	} else if u == nil {
		showLoginErrorMsgBox(s)
		return
	}

	// Decrypt and generate the temporary SHA256 hash with the session ID and random token.
	hash, err := decryptPasswordHash(u.PasswordHash)
	if err != nil {
		log.L.Error("failed to decrypt password hash for user '%s': %v", loginName, err)
		showLoginErrorMsgBox(s)
		return
	}
	hash = utils.Sha256Sum(hash + s.SessionID() + passwordToken)

	// Check if the password is valid.
	if passwordHash != hash {
		showLoginErrorMsgBox(s)
		return
	}

	// If this is the first login, then request a new password.
	if u.LastLogin <= 0 {
		if err = showChangePasswordDialog(s, u, finishLoginCallbackName); err != nil {
			log.L.Error(err.Error())
		}
		return
	}

	// Finish the login
	finishLogin(s, u)
}

//###############//
//### Private ###//
//###############//

func onLoginTemplateGetData(c *template.Context) interface{} {
	// Generate a new random password token.
	passwordToken := utils.RandomString(passwordTokenLength)

	// Save the password token to the session.
	c.Session().InstanceSet(sessionValueKeyPasswordToken, passwordToken)

	// Create the template render data.
	data := struct {
		RegistrationDisabled bool
		PasswordToken        string
	}{
		RegistrationDisabled: settings.Settings.RegistrationDisabled,
		PasswordToken:        passwordToken,
	}

	return data
}

func routeLoginPage(s *sessions.Session, routeData *router.Data) (string, string, error) {
	// If already authenticated, then redirect to the default page.
	if IsAuth(s) {
		backend.NavigateToPath(s, "/")
		return "", "", nil
	}

	// Execute the login template.
	o, _, _, err := templates.ExecuteTemplateToString(s, loginTemplate)
	if err != nil {
		return "", "", fmt.Errorf("failed to execute login template: %v", err)
	}

	return o, tr.S("blz.auth.login.pageTitle"), nil
}

func showLoginErrorMsgBox(s *sessions.Session) {
	// Show a messagebox
	messagebox.New().
		SetTitle(tr.S("blz.auth.login.error.title")).
		SetText(tr.S("blz.auth.login.error.text")).
		SetType(messagebox.TypeAlert).
		Show(s)
}

func finishLogin(s *sessions.Session, u *dbUser) {
	// Update the last login time
	err := dbUpdateLastLogin(u)
	if err != nil {
		log.L.Error("failed to update last login time of user '%s': %v", u.LoginName, err)
		showLoginErrorMsgBox(s)
		return
	}

	// Create a new session authentication data value.
	d := &sessionAuthData{
		UserID: u.ID,
	}

	// Save the authentication data to the session values.
	// This makes the user login public to the complete application.
	s.Set(sessionValueKeyIsAuth, d)

	// Redirect to the default page.
	backend.NavigateToPath(s, "/")

	// Trigger the event
	triggerOnNewAuthenticatedSession(s)
}
