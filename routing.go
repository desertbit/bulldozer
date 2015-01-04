/*
 *  Bulldozer Framework
 *  Copyright (C) DesertBit
 */

package bulldozer

import (
	"bytes"
	"code.desertbit.com/bulldozer/bulldozer/sessions"
	"code.desertbit.com/bulldozer/bulldozer/utils"
	"github.com/golang/glog"
	"sync"
)

// TODO: Implement a complex router and replace it with the current temporary fix

var (
	pageRoutes      map[string]*pageRoute = make(map[string]*pageRoute)
	pageRoutesMutex sync.Mutex
)

//#####################//
//### Private types ###//
//#####################//

type pageRoute struct {
	UID          string
	TemplateName string
}

//##############//
//### Public ###//
//##############//

func RoutePage(path string, pageTemplate string, UID string) {
	// Lock the mutex
	pageRoutesMutex.Lock()
	defer pageRoutesMutex.Unlock()

	// Create a new page route object
	p := &pageRoute{
		UID:          UID,
		TemplateName: pageTemplate,
	}

	// Print a warning if a previous route is set
	if _, ok := pageRoutes[path]; ok {
		glog.Warningf("overwriting previously set page route: '%s'", path)
	}

	// Set the new route
	pageRoutes[path] = p
}

/* TODO
type RouteFunc func(*Context)

func Route(path string, f RouteFunc) {
	mainRouter.Route(path, f)
}
*/

//###############//
//### Private ###//
//###############//

// execRoute executes the routes and returns the
// status code with the body string.
func execRoute(s *sessions.Session, path string) (int, string, error) {
	// This is a temporary fix
	path = utils.ToPath(path)

	// Lock the mutex
	pageRoutesMutex.Lock()
	defer pageRoutesMutex.Unlock()

	// Try to obtain the page route if present
	p, ok := pageRoutes[path]
	if !ok {
		// Execute the not found page
		out, err := execNotFoundTemplate()
		if err != nil {
			return 500, "", err
		}

		return 404, out, nil
	}

	// TODO!!!!!!!!!!!
	// Execute the template
	var b bytes.Buffer
	err := templates.ExecuteTemplate(s, &b, p.TemplateName, nil, p.UID)
	if err != nil {
		// TODO
		return 500, "Error executing template!", err
	}

	return 200, b.String(), nil
}
