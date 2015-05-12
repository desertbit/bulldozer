/*
 *  Bulldozer Framework
 *  Copyright (C) DesertBit
 */

package mux

import (
	"errors"
	"fmt"

	"github.com/desertbit/bulldozer/log"
	"github.com/desertbit/bulldozer/router"
	"github.com/desertbit/bulldozer/sessions"
	"github.com/desertbit/bulldozer/template"
	"github.com/desertbit/bulldozer/templates"
	"github.com/desertbit/bulldozer/utils"
	"github.com/desertbit/bulldozer/webcrawler"
)

var (
	backendI      Interface
	mainRouter    *router.Router = router.New()
	notFoundError                = errors.New("Not Found")
)

const (
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

//#################//
//### Interface ###//
//#################//

type Interface interface {
	ExecTopBar(i interface{}) (string, error)
}

//#############//
//### Types ###//
//#############//

type RouteFunc func(*sessions.Session, *Request)

type pageRoute struct {
	ID           string
	TemplateName string
	Title        string
}

//####################//
//### Request Type ###//
//####################//

type Request struct {
	RouteData *router.Data
	Title     string
	Body      string

	err error
}

// NotFoundError will show a not found error page.
func (r *Request) NotFoundError() {
	r.err = notFoundError
}

// Error will show an error page with the error message.
func (r *Request) Error(err error) {
	r.err = err
}

//##############//
//### Public ###//
//##############//

// Init initializes this packages.
// This is handled by the bulldozer package.
func Init(i Interface) {
	backendI = i
}

// Route the given path.
func Route(path string, f RouteFunc) {
	// Add the callback to the router.
	mainRouter.Route(path, f)
}

// RoutePage creates a new page route and executes the given template by its name.
// The templates of the bulldozer templates package are used.
// One optional variadic argument can be passed, which defines the page template ID.
// This ID is passed to the template.ExecOpts.
// The path is automatically added to the webcrawler sitemap paths.
// If you don't want to have this added, then remove the path from the webcrawler's sitemap again.
func RoutePage(path string, title string, templateName string, vars ...string) {
	// Create a new page route value.
	p := &pageRoute{
		TemplateName: templateName,
		Title:        title,
	}

	// Set the ID if present.
	if len(vars) > 0 {
		p.ID = vars[0]
	}

	// Add the value to the router.
	mainRouter.Route(path, p)

	// Add the path to the webcrawler's sitemap.
	webcrawler.AddSitemapPath(path)
}

// ExecRoute executes the routes and returns the status code
// with the body string, the title and the current path. The path might have changed, because it is normalized.
func ExecRoute(s *sessions.Session, requestedPath string) (statusCode int, body string, title string, path string) {
	// Recover panics and log the error message.
	defer func() {
		if e := recover(); e != nil {
			log.L.Error("mux: execute route panic: %v", e)
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
		statusCode, body, title = templates.ExecNotFound(s)
		return
	}

	// Show the error page if a parse template error occurred.
	if templates.ParseError != nil {
		// Execute the error template.
		statusCode, body, title = templates.ExecError(s, templates.ParseError.Error(), false)
		return
	}

	switch v := data.Value.(type) {
	case RouteFunc:
		// Create a new request value.
		r := &Request{
			RouteData: data,
		}

		// Call the function
		v(s, r)

		// Check the request.
		if r.err == notFoundError {
			// Execute the not found template.
			statusCode, body, title = templates.ExecNotFound(s)
			return
		} else if r.err != nil {
			// Execute the error template.
			statusCode, body, title = templates.ExecError(s, fmt.Sprintf("failed to execute route: '%s': %v", path, r.err))
			return
		}

		// Execute the topbar.
		topBarO, err := backendI.ExecTopBar(s)
		if err != nil {
			// Execute the error template.
			statusCode, body, title = templates.ExecError(s, fmt.Sprintf("failed to execute the topbar template: %v", err))
			return
		}

		// Set the title and the body from the request value.
		title = r.Title

		// The body is composed of the topbar and the page body.
		body = topBarO + r.Body

		return
	case *pageRoute:
		// Create the optional options for the template.
		opts := template.ExecOpts{
			ID:           v.ID,
			StyleClasses: []string{"bud-page"},
		}

		// Execute the template
		o, c, found, err := templates.Templates.ExecuteTemplateToString(s, v.TemplateName, opts)

		if err != nil {
			if found {
				// Execute the error template.
				statusCode, body, title = templates.ExecError(s, fmt.Sprintf("page '%s': '%s': %v", v.TemplateName, path, err))
				return
			} else {
				// Execute the not found template.
				statusCode, body, title = templates.ExecNotFound(s)
				return
			}
		}

		// Execute the topbar.
		topBarO, err := backendI.ExecTopBar(c)
		if err != nil {
			// Execute the error template.
			statusCode, body, title = templates.ExecError(s, fmt.Sprintf("failed to execute the topbar template: %v", err))
			return
		}

		// Set the title.
		title = v.Title

		// The body is composed of the topbar and the page body.
		body = topBarO + o

		return
	default:
		// Execute the error template.
		statusCode, body, title = templates.ExecError(s, fmt.Sprintf("failed to execute route: '%s': unkown value type!", path))
		return
	}
}

//###############//
//### Private ###//
//###############//

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
