/*
 *  Bulldozer Framework
 *  Copyright (C) DesertBit
 */

package callback

import (
	"github.com/desertbit/bulldozer/log"
	"reflect"
	"sync"
)

var (
	funcs      map[string]reflect.Value = make(map[string]reflect.Value)
	funcsMutex sync.Mutex
)

//##############//
//### Public ###//
//##############//

// Register a new callback with the name as key.
// One optional booleam can be passed, to force a overwrite of
// a previous registered callback with the same name.
func Register(name string, f interface{}, vars ...bool) {
	overwrite := false
	if len(vars) > 0 {
		overwrite = vars[0]
	}

	// Get the reflect value of f.
	v := reflect.ValueOf(f)

	// Check if f is a function.
	if v.Kind() != reflect.Func {
		log.L.Error("ailed to register new callback with name '%s': passed interface is not a function!", name)
		return
	}

	// Lock the mutex.
	funcsMutex.Lock()
	defer funcsMutex.Unlock()

	if !overwrite {
		// Check if the callback is already registered.
		if _, ok := funcs[name]; ok {
			log.L.Error("failed to register new callback with name '%s': a callback already exists with the same name!", name)
			return
		}
	}

	// Add the callback to the map.
	funcs[name] = v
}

func Call(name string, args ...interface{}) {
	// Recover panics and log the error.
	defer func() {
		if e := recover(); e != nil {
			log.L.Error("callback '%s' call panic: %v", name, e)
		}
	}()

	in := make([]reflect.Value, len(args))
	for i, arg := range args {
		in[i] = reflect.ValueOf(arg)
	}

	callback, ok := func() (cb reflect.Value, ok bool) {
		// Lock the mutex.
		funcsMutex.Lock()
		defer funcsMutex.Unlock()

		// Get the callback by its name.
		cb, ok = funcs[name]
		return
	}()
	if !ok {
		log.L.Error("failed to call callback with name '%s': a callback with the name does not exists!", name)
		return
	}

	// Call the callback.
	callback.Call(in)
}
