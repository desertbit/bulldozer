/*
 *  Bulldozer Framework
 *  Copyright (C) DesertBit
 */

package template

import (
	"fmt"
	"reflect"
	"strings"

	"code.desertbit.com/bulldozer/bulldozer/log"
)

const (
	FuncMethodPrefix = "Func"

	globalFuncsNameSpace = "global"
)

func init() {
	// Register the template parse function.
	registerParseFunc("call", parsecall)
}

//#############//
//### Types ###//
//#############//

type tFuncs map[string]*tFunc

type tFunc struct {
	receiver interface{}
	method   reflect.Value
}

//###############################//
//### Public Template methods ###//
//###############################//

// RegisterFuncs registeres the template function methods of the interface.
// Template function method names have to start with the 'Func' prefix.
// They are called from the template without this prefix.
// One optional parameter can be set, to define the functions namespace.
// If no namespace is defined, then the functions are registered in the global namespace.
// Only call this method before any template execution. This is not thread-safe.
//
// The first parameter of a template function has to be a template Context pointer.
// Valid return values are following:
//   - no return values
//   - error
//   - (interface{}, error)
func (t *Template) RegisterFuncs(i interface{}, vars ...string) *Template {
	// Get the namespace
	var namespace string
	if len(vars) > 0 {
		namespace = vars[0]
	}
	if len(namespace) == 0 {
		namespace = globalFuncsNameSpace
	}

	// Register all methods of the interface which start
	// with the function prefix.
	noMethods := true
	iType := reflect.TypeOf(i)
	for x := 0; x < iType.NumMethod(); x++ {
		method := iType.Method(x)
		name := method.Name

		// Skip this method if it does not start with the method prefix.
		if !strings.HasPrefix(name, FuncMethodPrefix) {
			continue
		}

		// Trim the prefix from the name
		name = strings.TrimPrefix(name, FuncMethodPrefix)

		// Create a new function map item value.
		f := &tFunc{
			receiver: i,
			method:   method.Func,
		}

		// Create the access key.
		key := namespace + "." + name

		// Create the template functions map if nil.
		createFuncsMapIfNil(t)

		// Log a warning if a previously registered method is overwritten.
		if _, overwritten := t.funcsMap[key]; overwritten {
			// Print a warning message.
			log.L.Error("template '%s': RegisterFuncs: overwriting template function: '%s'!", t.Name(), key)
		}

		// Add the method to the map.
		t.funcsMap[key] = f

		// Set the flag
		noMethods = false
	}

	if noMethods {
		log.L.Warning("template '%s': RegisterFuncs: registered interface template methods, but no template function methods where found!", t.Name())
	}

	return t
}

//###############//
//### Private ###//
//###############//

func createFuncsMapIfNil(t *Template) {
	if t.funcsMap == nil {
		t.funcsMap = make(tFuncs)
	}
}

//  Bulldozer function. This is called during the template execution.
// {{call FuncName arg1 arg2 arg3}}
func parsecall(typeStr string, token string, d *parseData) error {
	var funcName string

	// Try to find the first empty space.
	pos := strings.Index(token, " ")

	// If not found, then there are no arguments.
	if pos == -1 {
		funcName = token
		token = ""
	} else {
		// Extract the function name and remove it from the original token.
		funcName = strings.TrimSpace(token[0:pos])
		token = strings.TrimSpace(token[pos+1:])
	}

	// Set the namespace to the default global one.
	namespace := globalFuncsNameSpace

	// Find the function namespace if present.
	pos = strings.Index(funcName, ".")
	if pos != -1 {
		// Extract the function namespace and remove it from the function name.
		namespace = funcName[:pos]
		funcName = funcName[pos+1:]
	}

	// Check if the function name and namespace are valid.
	if funcName == "" {
		return fmt.Errorf("empty function name!")
	} else if namespace == "" {
		return fmt.Errorf("empty function namespace!")
	}

	// Create the function access key.
	funcKey := namespace + "." + funcName

	// Create the final template function call.
	*d.final += `{{callFunc $.Context "` + funcKey + `" ` + token + `}}`

	return nil
}

func callTemplateFunc(c *Context, funcKey string, params ...interface{}) (r interface{}, err error) {
	// Recover panics and return the error.
	defer func() {
		if e := recover(); e != nil {
			err = fmt.Errorf("failed to call template function '%s': %v", funcKey, e)
		}
	}()

	// Check if the map is not created.
	if c.t.funcsMap == nil {
		return "", fmt.Errorf("failed to call template function '%s': the function does not exists!", funcKey)
	}

	// Get the function.
	f, ok := c.t.funcsMap[funcKey]
	if !ok {
		return "", fmt.Errorf("failed to call template function '%s': the function does not exists!", funcKey)
	}

	// Create the parameters slice.
	in := make([]reflect.Value, len(params)+2)

	// Add the receiver and the context as first function parameters.
	in[0] = reflect.ValueOf(f.receiver)
	in[1] = reflect.ValueOf(c)

	// Add all the passed parameters.
	for i, p := range params {
		in[i+2] = reflect.ValueOf(p)
	}

	// Call the method
	values := f.method.Call(in)

	// Helper function to obtain the error value from the return values.
	getErrorValue := func(v reflect.Value) (err error) {
		// Get the interface.
		i := v.Interface()
		if i == nil {
			return nil
		}

		// Get the error
		err, ok = i.(error)
		if !ok {
			err = fmt.Errorf("the function return value is not of type error!")
		}

		// Add the prefix description.
		if err != nil {
			err = fmt.Errorf("failed to call template function '%s': %v", funcKey, err)
		}

		return
	}

	l := len(values)
	if l == 0 {
		// No return value specified.
		return "", nil
	} else if l == 1 {
		// one return value. Must be nil or error.
		return "", getErrorValue(values[0])
	} else if l == 2 {
		// Two return values specified.
		return values[0].Interface(), getErrorValue(values[1])
	}

	return "", fmt.Errorf("failed to call template function '%s': invalid function return values!", funcKey)
}
