/*
 *  Bulldozer Framework
 *  Copyright (C) DesertBit
 */

package bulldozer

import (
	"code.desertbit.com/bulldozer/bulldozer/global"
	"code.desertbit.com/bulldozer/bulldozer/router"
	"code.desertbit.com/bulldozer/bulldozer/sessions"
	"code.desertbit.com/bulldozer/bulldozer/template"
	"code.desertbit.com/bulldozer/bulldozer/utils"
	"fmt"
)

var (
	mainRouter *router.Router = router.New()
)

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

// TODO: route requests...

//###############//
//### Private ###//
//###############//

// execRoute executes the routes and returns the status code with the body string and title.
func execRoute(s *sessions.Session, path string) (int, string, string) {
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
		o, found, err := global.TemplatesStore.Templates.ExecuteTemplateToString(s, v.TemplateName, nil, v.UID)
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
