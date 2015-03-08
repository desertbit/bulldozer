/*
 *  Bulldozer Framework
 *  Copyright (C) DesertBit
 */

package topbar

import (
	"code.desertbit.com/bulldozer/bulldozer/auth"
	"code.desertbit.com/bulldozer/bulldozer/controlpanel"
	"code.desertbit.com/bulldozer/bulldozer/editmode"
	"code.desertbit.com/bulldozer/bulldozer/sessions"
	"code.desertbit.com/bulldozer/bulldozer/store"
	"code.desertbit.com/bulldozer/bulldozer/template"
	"code.desertbit.com/bulldozer/bulldozer/templates"
	"fmt"
)

const (
	topbarTemplateName   = "bud/topbar/topbar"
	topbarEventNamespace = "budTB"
	topbarTemplateID     = "budTB"
)

var (
	topbarTemplate *template.Template
)

func init() {
	auth.OnEndAuthenticatedSession(onEndAuthenticatedSession)
}

//##############//
//### Public ###//
//##############//

func Init() (err error) {
	// Obtain the topbar template.
	topbarTemplate = templates.Templates.Lookup(topbarTemplateName)
	if topbarTemplate == nil {
		return fmt.Errorf("failed to lookup topbar template!")
	}

	// Register the topbar events.
	topbarTemplate.RegisterEvents(new(events), topbarEventNamespace)

	return nil
}

// RegisterEvents passes the events to the underlying topbar template.
// One optional parameter can be set, to define the events namespace.
// If no namespace is defined, then the event is registered in the global namespace.
// Call this method not before the bulldozer initialization process!
func RegisterEvents(i interface{}, vars ...string) {
	topbarTemplate.RegisterEvents(i, vars...)
}

// ExecTopBar executes the topbar.
// You can pass a session or context value to this method.
// If a context value is available, then always pass it instead of the session.
// This will improve the performance and won't retrieve a user value multiple
// times from the database during one template execution cycle.
// The context is only used to retrieve the authenticated user information if present.
// The topbar is executed in it's own context anyway.
func ExecTopBar(i interface{}) (string, error) {
	var s *sessions.Session

	switch i.(type) {
	case *sessions.Session:
		// Set the session pointer.
		s = i.(*sessions.Session)
	case *template.Context:
		// Assert to context value.
		c := i.(*template.Context)

		// Set the session pointer.
		s = c.Session()
	default:
		return "", fmt.Errorf("invalid topbar.ExecTopBar call: called method with invalid interface type!")
	}

	// Get the current user if logged in.
	user := auth.GetUser(i)
	if user == nil {
		return "", nil
	}

	// Template render data.
	data := struct {
		User               *auth.User
		IsControlPanelPage bool
		ControlPanelUrl    string
		EnableEditmode     bool
		EditmodeActive     bool
		LeftMenu           Items
		RightMenu          Items
		EditmodeMenu       Items
		UserMenu           Items
	}{
		User:               user,
		IsControlPanelPage: controlpanel.IsCurrentPage(s),
		ControlPanelUrl:    controlpanel.PageUrl,
		EnableEditmode:     false,
		EditmodeActive:     false,
		LeftMenu:           leftMenuItems,
		RightMenu:          rightMenuItems,
		EditmodeMenu:       editmodeMenuItems,
		UserMenu:           userMenuItems,
	}

	// Only enable the edit mode for sysops and admins.
	if user.IsInGroup(auth.GroupSysOp, auth.GroupAdmin) {
		data.EnableEditmode = true
		data.EditmodeActive = editmode.IsActive(s)
	}

	// Execute the topbar template.
	body, _, err := topbarTemplate.ExecuteToString(s, template.ExecOpts{
		ID:   topbarTemplateID,
		Data: data,
	})
	if err != nil {
		return "", err
	}

	return body, nil
}

//###############//
//### Private ###//
//###############//

func onEndAuthenticatedSession(s *sessions.Session) {
	s.SendCommand("Bulldozer.topbar.space(false);")
}

//##############//
//### Events ###//
//##############//

type events struct{}

func (e *events) EventLogout(c *template.Context) {
	// Logout.
	// The loading indicator is hidden by the page reload.
	auth.Logout(c.Session())
}

func (e *events) EventStartEdit(c *template.Context) {
	// Start the edit mode.
	// The loading indicator is hidden by the page reload.
	editmode.Start(c.Session())
}

func (e *events) EventStopEdit(c *template.Context) {
	// Stop the edit mode.
	// The loading indicator is hidden by the page reload.
	editmode.Stop(c.Session())
}

func (e *events) EventSaveChanges(c *template.Context) {
	// Get the session pointer.
	s := c.Session()

	// Hide the loading indicator on return.
	defer s.HideLoadingIndicator()

	// Save all temporary changes.
	store.SaveTemporaryChanges(s)
}

func (e *events) EventDiscardChanges(c *template.Context) {
	// Get the session pointer.
	s := c.Session()

	// Hide the loading indicator on return.
	defer s.HideLoadingIndicator()

	// Remove all temporary changes.
	//store.DiscardTemporaryChanges(s)
}
