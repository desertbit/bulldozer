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

//###############//
//### Private ###//
//###############//

// execRoute executes the routes and returns the status code with the body string and title.
func execRoute(s *sessions.Session, path string) (int, string, string) {
	// Recover panics and log the error message.
	defer func() {
		if e := recover(); e != nil {
			log.L.Error("bulldozer execute route panic: %v", e)
		}
	}()

	// Release the previous temmplate session events.
	template.ReleaseSessionEvents(s)

	// Transform the path to a valid path.
	path = utils.ToPath(path)

	// Execute the route.
	data := mainRouter.Match(path)
	if data == nil {
		// Execute the not found template.
		return global.ExecNotFoundTemplate(s)
	}

	switch v := data.Value.(type) {
	case RouteFunc:
		o, title, err := v(s, data)
		if err != nil {
			// Execute the error template.
			return global.ExecErrorTemplate(s, fmt.Sprintf("failed to execute route: '%s': %v", path, err))
		}

		return 200, o, title
	case *pageRoute:
		// Execute the template
		o, found, err := global.TemplatesStore.Templates.ExecuteTemplateToString(s, v.TemplateName, nil, v.UID, "bulldozer-page")
		if err != nil {
			if found {
				// Execute the error template.
				return global.ExecErrorTemplate(s, fmt.Sprintf("page '%s': '%s': %v", v.TemplateName, path, err))
			} else {
				// Execute the not found template.
				return global.ExecNotFoundTemplate(s)
			}
		}

		return 200, o, v.Title
	default:
		// Execute the error template.
		return global.ExecErrorTemplate(s, fmt.Sprintf("failed to execute route: '%s': unkown value type!", path))
	}
}

// sessionRequestRoute is triggered from the client side.
func sessionRequestRoute(s *sessions.Session, data map[string]string) error {
	// Try to obtain the route path.
	path, ok := data[keyRoutePath]
	if !ok {
		return fmt.Errorf("failed to execute route: missing route path!")
	}

	// Execute the route.
	_, body, title := execRoute(s, path)

	// Create the client command.
	cmd := `Bulldozer.render.page('` +
		utils.EscapeJS(body) + `','` +
		utils.EscapeJS(title) + `','` +
		utils.EscapeJS(path) + `');`

	// Send the new render request to the client.
	s.SendCommand(cmd)

	return nil
}
