/*
 *  Bulldozer Framework
 *  Copyright (C) DesertBit
 */

package bulldozer

import (
	"bytes"
	"code.desertbit.com/bulldozer/bulldozer/router"
	"code.desertbit.com/bulldozer/bulldozer/sessions"
	"code.desertbit.com/bulldozer/bulldozer/utils"
	"fmt"
)

var (
	mainRouter *router.Router = router.New()
)

//#############//
//### Types ###//
//#############//

type RouteFunc func(*sessions.Session, *router.Data) (string, error)

type pageRoute struct {
	UID          string
	TemplateName string
}

//##############//
//### Public ###//
//##############//

func RoutePage(path string, pageTemplate string, UID string) {
	// Create a new page route value.
	p := &pageRoute{
		UID:          UID,
		TemplateName: pageTemplate,
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

// execRoute executes the routes and returns the status code with the body string.
func execRoute(s *sessions.Session, path string) (int, string, error) {
	// Transform the path to a valid path.
	path = utils.ToPath(path)

	// Execute the route.
	data := mainRouter.Match(path)
	if data == nil {
		// Execute the not found page
		out, err := execNotFoundTemplate()
		if err != nil {
			return 500, "", err
		}

		return 404, out, nil
	}

	switch v := data.Value.(type) {
	case RouteFunc:
		o, err := v(s, data)
		if err != nil {
			// TODO!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!
			return 500, "Error executing route!", err
		}

		return 200, o, nil
	case *pageRoute:
		// Execute the template
		var b bytes.Buffer
		err := templates.ExecuteTemplate(s, &b, v.TemplateName, nil, v.UID)
		if err != nil {
			// TODO!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!
			return 500, "Error executing template!", err
		}

		return 200, b.String(), nil
	default:
		// TODO!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!
		return 500, "Error executing template!", fmt.Errorf("failed to execute route: '%s': unkown value type!", path)
	}
}
