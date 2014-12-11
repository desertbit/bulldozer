/*
 *  Bulldozer Framework
 *  Copyright (C) DesertBit
 */

package bulldozer

import (
	"github.com/golang/glog"
	"strings"
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
	UId          string
	TemplateName string
}

//##############//
//### Public ###//
//##############//

func RoutePage(path string, pageTemplate string, uId string) {
	// Lock the mutex
	pageRoutesMutex.Lock()
	defer pageRoutesMutex.Unlock()

	// Create a new page route object
	p := &pageRoute{
		UId:          uId,
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
func execRoute(path string) (int, string) {
	// This is a temporary fix
	path = toPath(path)

	// Lock the mutex
	pageRoutesMutex.Lock()
	defer pageRoutesMutex.Unlock()

	// Try to obtain the page route if present
	_, ok := pageRoutes[path]
	if !ok {
		// TODO: Handle 404 not found
		return 404, "404 Not found"
	}

	return 200, "Hello World"
}

// TODO: Remove this again. This is a temporary fix
// toPath returns a valid path
func toPath(path string) string {
	// Trim, to lower and replace all empty spaces
	path = strings.Replace(strings.ToLower(strings.TrimSpace(path)), " ", "-", -1)

	// Remove the following / if necessary
	if strings.HasSuffix(path, "/") {
		path = path[:len(path)-1]
	}

	// Append a leading / if necessary
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}

	return path
}
