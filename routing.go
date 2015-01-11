/*
 *  Bulldozer Framework
 *  Copyright (C) DesertBit
 */

package bulldozer

import (
	"code.desertbit.com/bulldozer/bulldozer/global"
	"code.desertbit.com/bulldozer/bulldozer/log"
	"code.desertbit.com/bulldozer/bulldozer/router"
	"code.desertbit.com/bulldozer/bulldozer/sessions"
	"code.desertbit.com/bulldozer/bulldozer/template"
	"code.desertbit.com/bulldozer/bulldozer/utils"
	"fmt"
)

var (
	mainRouter *router.Router = router.New()

	valueKeyCurrentPath = "bzrCurrentPath"

	requestTypeRoute = "route"
	keyRoutePath     = "path"
)

func init() {
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

func Route(path string, f RouteFunc) {
	// Add the callback to the router.
	mainRouter.Route(path, f)
}

// GetCurrentPath returns the current session route path.
func GetCurrentPath(s *sessions.Session) string {
	// Get the session current path. Create and add it, if not present.
	i, _ := s.Get(valueKeyCurrentPath, func() interface{} {
		return "/"
	})

	// Assertion
	path, ok := i.(string)
	if !ok {
		// Log the error
		log.L.Error("get currrent session path: failed to assert session value to string!")

		// Just set it to the session.
		path = "/"
		s.Set(valueKeyCurrentPath, path)
	}

	return path
}

// ReloadPage reloads the current session page.
func ReloadPage(s *sessions.Session) {
	Navigate(s, GetCurrentPath(s))
}

// Navigate navigates the session to the given route path.
func Navigate(s *sessions.Session, path string) {
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

//###############//
//### Private ###//
//###############//

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

	// Release the previous temmplate session events.
	template.ReleaseSessionEvents(s)

	// Transform the path to a valid path.
	path = utils.ToPath(requestedPath)

	// Set the path to the current session path.
	s.Set(valueKeyCurrentPath, path)

	// Execute the route.
	data := mainRouter.Match(path)
	if data == nil {
		// Execute the not found template.
		statusCode, body, title = global.ExecNotFoundTemplate(s)
		return
	}

	var err error

	switch v := data.Value.(type) {
	case RouteFunc:
		body, title, err = v(s, data)
		if err != nil {
			// Execute the error template.
			statusCode, body, title = global.ExecErrorTemplate(s, fmt.Sprintf("failed to execute route: '%s': %v", path, err))
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
		o, _, found, err := global.TemplatesStore.Templates.ExecuteTemplateToString(s, v.TemplateName, nil, opts)

		if err != nil {
			if found {
				// Execute the error template.
				statusCode, body, title = global.ExecErrorTemplate(s, fmt.Sprintf("page '%s': '%s': %v", v.TemplateName, path, err))
				return
			} else {
				// Execute the not found template.
				statusCode, body, title = global.ExecNotFoundTemplate(s)
				return
			}
		}

		return 200, o, v.Title, path
	default:
		// Execute the error template.
		statusCode, body, title = global.ExecErrorTemplate(s, fmt.Sprintf("failed to execute route: '%s': unkown value type!", path))
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
	Navigate(s, path)

	return nil
}
