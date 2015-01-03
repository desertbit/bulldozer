/*
 *  Bulldozer Framework
 *  Copyright (C) DesertBit
 */

package template

import (
	"github.com/golang/glog"
	"reflect"
	"strings"
)

const (
	MustMethodPrefix = "Must"
)

var (
	packages  map[string]interface{} = make(map[string]interface{})
	mustFuncs map[string]*mustFunc   = make(map[string]*mustFunc)
)

//#############//
//### Types ###//
//#############//

type mustFunc struct {
	receiver reflect.Value
	method   reflect.Value
}

type Action struct {
	stopped    bool
	redirectTo string
}

func newAction() *Action {
	return &Action{
		stopped: false,
	}
}

func (a *Action) Stop() {
	a.stopped = true
}

func (a *Action) Redirect(url string) {
	a.redirectTo = url
	a.stopped = true
}

//##############//
//### Public ###//
//##############//

// RegisterPackage registeres a new template package.
// This call is not thread-safe! Register packages during program initialization.
func RegisterPackage(name string, i interface{}) {
	// Log an error message if a package is overwritten,
	_, ok := packages[name]
	if ok {
		glog.Errorf("template: RegisterPackage: overwritting already present package: '%s'", name)
	}

	// Add the package to the packages map.
	packages[name] = i

	// Dummy values.
	dummyAction := new(Action)
	dummyContext := new(Context)

	// Find and register all must functions.
	iType := reflect.TypeOf(i)
	for x := 0; x < iType.NumMethod(); x++ {
		method := iType.Method(x)
		funcName := method.Name

		// Skip this method if it does not start with the method prefix.
		if !strings.HasPrefix(funcName, MustMethodPrefix) {
			continue
		}

		// Trim the prefix from the name.
		funcName = strings.TrimPrefix(funcName, MustMethodPrefix)

		// Create the function key.
		key := name + "." + funcName

		// Get the function and the type of the function.
		f := method.Func
		t := f.Type()

		// Check if the fixed parameters are defined.
		if t.NumIn() != 3 ||
			reflect.TypeOf(i) != t.In(0) ||
			reflect.TypeOf(dummyAction) != t.In(1) ||
			reflect.TypeOf(dummyContext) != t.In(2) {
			glog.Errorf("must function '%s': invalid function parameters! The first two parameters have to be an Action and Context pointer.", key)
			continue
		}

		// Create the mustFunc value.
		m := &mustFunc{
			receiver: reflect.ValueOf(i),
			method:   f,
		}

		// Add it to the map.
		mustFuncs[key] = m
	}
}

//###############//
//### Private ###//
//###############//

func (t *Template) callMustFuncs(c *Context) (action *Action) {
	if t.mustFuncs == nil || len(t.mustFuncs) == 0 {
		return nil
	}

	// Create a new action.
	action = newAction()

	// Create the parameters slice.
	in := make([]reflect.Value, 3)

	// Fill the parameters slice.
	in[1] = reflect.ValueOf(action)
	in[2] = reflect.ValueOf(c)

	// Iterate through all must functions.
	for _, f := range t.mustFuncs {
		// Set the receiver.
		in[0] = f.receiver

		// Call the method.
		f.method.Call(in)

		// Check if a stop is requested.
		if action.stopped {
			return
		}
	}

	return
}
