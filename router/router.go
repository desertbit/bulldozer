/*
 *  Bulldozer Framework
 *  Copyright (C) DesertBit
 */

package router

import (
	"github.com/desertbit/bulldozer/log"
	"github.com/desertbit/bulldozer/utils"
	"strings"
	"sync"
)

//#############//
//### Types ###//
//#############//

type Params map[string]string

type Data struct {
	Value    interface{}
	Path     string
	RestPath string
	Params   Params
}

//###################//
//### Router type ###//
//###################//

type Router struct {
	route *routePart

	paths      []string
	pathsMutex sync.Mutex
}

func New() *Router {
	return &Router{
		route: newRoutePart(),
	}
}

func (r *Router) Route(path string, value interface{}) {
	// Transform the path to a valid path string.
	path = utils.ToPath(path)

	// Split the path into parts.
	parts := strings.Split(path, "/")

	// Remove empty parts.
	var tmp []string
	for _, p := range parts {
		if len(p) != 0 {
			tmp = append(tmp, p)
		}
	}
	parts = tmp

	// Check if the parts slice is empty.
	if parts == nil || len(parts) == 0 {
		if r.route.value != nil {
			log.L.Warning("router: overwriting already set route for the root path!")
		}

		// Set the value to the root part.
		r.route.value = value

		// Add the path to the paths slice.
		r.addPath(path)

		return
	}

	// Add the route.
	if overwritten := r.route.Set(parts, value); overwritten {
		log.L.Warning("router: overwriting already set route path: '%s'", path)
	}

	// Add the path to the paths slice.
	r.addPath(path)
}

func (r *Router) Match(path string) *Data {
	// Transform the path to a valid path string.
	path = utils.ToPath(path)

	// Split the path into parts.
	parts := strings.Split(path, "/")

	// Create the route data value.
	rData := &Data{
		Path:   path,
		Params: make(Params),
	}

	// Get the part value for the path.
	d := r.route.Get(parts, rData)

	// Call the route callback.
	if d != nil && d.value != nil {
		rData.Value = d.value
		return rData
	}

	return nil
}

// Paths returns all the current set route paths.
func (r *Router) Paths() []string {
	// Lock the mutex.
	r.pathsMutex.Lock()
	defer r.pathsMutex.Unlock()

	return r.paths
}

func (r *Router) addPath(path string) {
	// Lock the mutex.
	r.pathsMutex.Lock()
	defer r.pathsMutex.Unlock()

	// Check if already present in the slice.
	for _, p := range r.paths {
		if p == path {
			return
		}
	}

	// Add the path to the slice.
	r.paths = append(r.paths, path)
}

//###############//
//### Private ###//
//###############//

type routePart struct {
	parts   map[string]*routePart
	mutex   sync.Mutex
	varName string
	value   interface{}
}

func newRoutePart() *routePart {
	return &routePart{}
}

func (r *routePart) GetCreatePart(key string) *routePart {
	// Lock the mutex.
	r.mutex.Lock()
	defer r.mutex.Unlock()

	// Create the map of nil.
	if r.parts == nil {
		r.parts = make(map[string]*routePart)
	}

	// Set the data to the route part.
	rP, ok := r.parts[key]
	if !ok {
		rP = newRoutePart()
		r.parts[key] = rP
	}

	return rP
}

func (r *routePart) GetPart(key string) *routePart {
	// Lock the mutex.
	r.mutex.Lock()
	defer r.mutex.Unlock()

	// Check if the map is nil.
	if r.parts == nil {
		return nil
	}

	// Get the value.
	return r.parts[key]
}

func (r *routePart) Set(parts []string, value interface{}) (overwritten bool) {
	// Get the path part and remove it from the slice.
	part := parts[0]
	parts = parts[1:]

	// Check if this path part is a path variable.
	var varName string
	if strings.HasPrefix(part, ":") {
		// Save the variable name.
		varName = part[1:]

		// Set the placeholder.
		part = "%"
	}

	// Get or create the route part.
	rP := r.GetCreatePart(part)

	// Set the variable name to the value if not emtpy.
	if len(varName) != 0 {
		rP.varName = varName
	}

	if len(parts) == 0 {
		if rP.value != nil {
			overwritten = true
		}

		// Set the value.
		rP.value = value
	} else {
		// Pass the data to the underlying parts...
		overwritten = rP.Set(parts, value)
	}

	return
}

func (r *routePart) Get(parts []string, data *Data) *routePart {
	// Get the path part and remove it from the slice.
	// Skip empty parts.
	var part string
	for {
		if len(parts) == 0 {
			return r
		}

		part = parts[0]
		parts = parts[1:]

		if len(part) == 0 {
			continue
		}

		break
	}

	// Check if a route exists.
	rP := r.GetPart(part)
	if rP != nil {
		// If there is no more path part, then the route matches rP.
		if len(parts) == 0 {
			return rP
		} else {
			// Check for more sub route matches.
			rP = rP.Get(parts, data)
			if rP != nil {
				return rP
			}
		}
	}

	// Check if a variable is defined.
	rP = r.GetPart("%")
	if rP != nil {
		// If there is no more path part, then the route matches rP.
		if len(parts) == 0 {
			// Add the variable to the parameters map.
			data.Params[rP.varName] = part

			return rP
		} else {
			// Check for more sub route matches.
			d := rP.Get(parts, data)
			if d != nil {
				// Add the variable to the parameters map.
				data.Params[rP.varName] = part

				return d
			}
		}
	}

	// Check if a wildcard is defined.
	rP = r.GetPart("*")
	if rP != nil {
		// Save the rest of the path.
		data.RestPath = strings.TrimSuffix(part+"/"+strings.Join(parts, "/"), "/")

		return rP
	}

	return nil
}
