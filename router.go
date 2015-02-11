/*
 *  Bulldozer Framework
 *  Copyright (C) DesertBit
 */

package bulldozer

import (
	"code.desertbit.com/bulldozer/bulldozer/backend/topbar"
	"code.desertbit.com/bulldozer/bulldozer/log"
	"code.desertbit.com/bulldozer/bulldozer/router"
	"code.desertbit.com/bulldozer/bulldozer/sessions"
	"code.desertbit.com/bulldozer/bulldozer/template"
	"code.desertbit.com/bulldozer/bulldozer/utils"
	"fmt"
)

const (
	requestTypeRoute = "route"
	keyRoutePath     = "path"
)

var (
	mainRouter *router.Router = router.New()
)

func init() {
	// Set the session navigate function hook.
	sessions.SetNavigateFunc(sessionNavigateRequest)

	// Register the route server request.
	err := sessions.Request(requestTypeRoute, sessionRequestRoute)
	if err != nil {
		log.L.Fatalf("failed to register session route request: %v", err)
	}
}

//#############//
//### Types ###//
//#############//

// RouteFunc returns the body output, a title and an error if present.
type RouteFunc func(*sessions.Session, *router.Data) (string, string, error)

type pageRoute struct {
	UID          string
	TemplateName string
	Title        string
}

//##############//
//### Public ###//
//##############//

// RoutePage creates a new page route.
func RoutePage(path string, pageTitle string, pageTemplate string, UID string) {
	// Create a new page route value.
	p := &pageRoute{
		UID:          UID,
		TemplateName: pageTemplate,
		Title:        pageTitle,
	}

	// Add the value to the router.
	mainRouter.Route(path, p)
}

// Route the given path.
func Route(path string, f RouteFunc) {
	// Add the callback to the router.
	mainRouter.Route(path, f)
}

// RoutePaths returns all current route paths.
func RoutePaths() []string {
	return mainRouter.Paths()
}

//###############//
//### Private ###//
//###############//

// sessionNavigateRequest navigates the session to the given route path.
func sessionNavigateRequest(s *sessions.Session, path string) {
	// Execute the route.
	_, body, title, path := execRoute(s, path)

	// Create the client command.
	cmd := `Bulldozer.render.page('` +
		utils.EscapeJS(body) + `','` +
		utils.EscapeJS(title) + `','` +
		utils.EscapeJS(path) + `');`

	// Send the new render request to the client.
	s.SendCommand(cmd)
}

// execRoute executes the routes and returns the status code
// with the body string, the title and the current path. The path might have changed...
func execRoute(s *sessions.Session, requestedPath string) (statusCode int, body string, title string, path string) {
	// Recover panics and log the error message.
	defer func() {
		if e := recover(); e != nil {
			log.L.Error("bulldozer execute route panic: %v", e)
		}
	}()

	// Set the default status code value.
	statusCode = 200

	// Reset the template environment.
	template.ResetEnvironment(s)

	// Transform the path to a valid path.
	path = utils.ToPath(requestedPath)

	// Set the current session path.
	s.SetCurrentPath(path)

	// Execute the route.
	data := mainRouter.Match(path)
	if data == nil {
		// Execute the not found template.
		statusCode, body, title = execNotFoundTemplate(s)
		return
	}

	var err error

	switch v := data.Value.(type) {
	case RouteFunc:
		body, title, err = v(s, data)
		if err != nil {
			// Execute the error template.
			statusCode, body, title = execErrorTemplate(s, fmt.Sprintf("failed to execute route: '%s': %v", path, err))
			return
		}

		return
	case *pageRoute:
		// Create the optional options for the template.
		opts := template.ExecOpts{
			ID:           v.UID,
			StyleClasses: []string{"bulldozer-page"},
		}

		// Execute the template
		o, c, found, err := TemplatesStore.Templates.ExecuteTemplateToString(s, v.TemplateName, opts)

		if err != nil {
			if found {
				// Execute the error template.
				statusCode, body, title = execErrorTemplate(s, fmt.Sprintf("page '%s': '%s': %v", v.TemplateName, path, err))
				return
			} else {
				// Execute the not found template.
				statusCode, body, title = execNotFoundTemplate(s)
				return
			}
		}

		// Execute the topbar.
		topBarO, err := topbar.ExecTopBar(c)
		if err != nil {
			// Execute the error template.
			statusCode, body, title = execErrorTemplate(s, fmt.Sprintf("failed to execute the topbar template: %v", err))
			return
		}

		// The body is composed of the topbar and the page body.
		body = topBarO + o

		return 200, body, v.Title, path
	default:
		// Execute the error template.
		statusCode, body, title = execErrorTemplate(s, fmt.Sprintf("failed to execute route: '%s': unkown value type!", path))
		return
	}
}

// sessionRequestRoute is triggered from the client side.
func sessionRequestRoute(s *sessions.Session, data map[string]string) error {
	// Try to obtain the route path.
	path, ok := data[keyRoutePath]
	if !ok {
		return fmt.Errorf("failed to execute route: missing route path!")
	}

	// Navigate to the path.
	s.Navigate(path)

	return nil
}
