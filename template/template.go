/*
 *  Bulldozer Framework
 *  Copyright (C) DesertBit
 */

// This template extends the golang template with a custom parse method.
// Some methods are copied from the original golang template to ensure compatibility.

package template

import (
	htmlTemplate "html/template"

	"fmt"
	"io/ioutil"
	"path/filepath"
	"sync"
)

//################################//
//### Private namespace struct ###//
//################################//

// nameSpace is the data structure shared by all templates in an association.
type nameSpace struct {
	mutex sync.Mutex
	set   map[string]*Template
}

func newNameSpace() *nameSpace {
	return &nameSpace{set: make(map[string]*Template)}
}

func (ns *nameSpace) Set(t *Template) {
	// Lock the mutex
	ns.mutex.Lock()
	defer ns.mutex.Unlock()

	// Add the template to the map with the template name as key.
	ns.set[t.Name()] = t
}

func (ns *nameSpace) Get(name string) *Template {
	// Lock the mutex
	ns.mutex.Lock()
	defer ns.mutex.Unlock()

	// Try to find the template
	return ns.set[name]
}

//#######################//
//### Template struct ###//
//#######################//

type Template struct {
	// We could embed the text/template field, but it's safer not to because
	// we need to keep our version of the name space and the underlying
	// template's in sync.
	template *htmlTemplate.Template

	// The namespace
	ns *nameSpace

	// Delimiters
	leftDelim, rightDelim string

	// Style classes
	styleClasses []string
}

// New allocates a new bulldozer template associated with the given one
// and with the same delimiters. The association, which is transitive,
// allows one template to invoke another with a {{template}} action.
func (t *Template) New(name string) *Template {
	tt := &Template{
		template:   t.template.New(name),
		leftDelim:  t.leftDelim,
		rightDelim: t.rightDelim,
		ns:         t.ns,
	}

	// Set the functions
	tt.Funcs(bulldozerFuncMap)

	// Add the new template to the namespace
	tt.ns.Set(tt)

	return tt
}

// Lookup returns the template with the given name that is associated with t,
// or nil if there is no such template.
func (t *Template) Lookup(name string) *Template {
	// Lookup the template in the namespace
	return t.ns.Get(name)
}

// Name returns the name of the template.
func (t *Template) Name() string {
	return t.template.Name()
}

// AddStyleClass adds a style class to the template div.
// The return value is the template, so calls can be chained.
func (t *Template) AddStyleClass(styleClass string) *Template {
	t.styleClasses = append(t.styleClasses, styleClass)
	return t
}

// Delims sets the action delimiters to the specified strings, to be used in
// subsequent calls to Parse, ParseFiles, or ParseGlob. Nested template
// definitions will inherit the settings. An empty delimiter stands for the
// corresponding default: {{ or }}.
// The return value is the template, so calls can be chained.
func (t *Template) Delims(left, right string) *Template {
	if left == "" {
		left = "{{"
	}
	if right == "" {
		right = "}}"
	}

	t.leftDelim = left
	t.rightDelim = right
	t.template.Delims(left, right)
	return t
}

// FuncMap is the type of the map defining the mapping from names to
// functions. Each function must have either a single return value, or two
// return values of which the second has type error. In that case, if the
// second (error) argument evaluates to non-nil during execution, execution
// terminates and Execute returns that error. FuncMap has the same base type
// as FuncMap in "text/template", copied here so clients need not import
// "text/template".
type FuncMap map[string]interface{}

// Funcs adds the elements of the argument map to the template's function map.
// It panics if a value in the map is not a function with appropriate return
// type. However, it is legal to overwrite elements of the map. The return
// value is the template, so calls can be chained.
func (t *Template) Funcs(funcMap FuncMap) *Template {
	t.template.Funcs(htmlTemplate.FuncMap(funcMap))
	return t
}

// Templates returns a slice of the templates associated with t, including t
// itself.
func (t *Template) Templates() []*Template {
	t.ns.mutex.Lock()
	defer t.ns.mutex.Unlock()

	// Return a slice so we don't expose the map.
	m := make([]*Template, 0, len(t.ns.set))
	for _, v := range t.ns.set {
		m = append(m, v)
	}
	return m
}

// Parse parses a string into a template. Nested template definitions
// will be associated with the top-level template t. Parse may be
// called multiple times to parse definitions of templates to associate
// with t. It is an error if a resulting template is non-empty (contains
// content other than template definitions) and would replace a
// non-empty template with the same name.  (In multiple calls to Parse
// with the same receiver template, only one call can contain text
// other than space, comments, and template definitions.)
func (t *Template) Parse(src string) (*Template, error) {
	// Call the custom bulldozer parse method
	src, err := parse(t, src, 0)
	if err != nil {
		return nil, err
	}

	// Call the html template parse method
	ret, err := t.template.Parse(src)
	if err != nil {
		return nil, err
	}

	// In general, all the named templates might have changed underfoot.
	// Regardless, some new ones may have been defined.
	// The template.Template set has been updated; update ours.
	t.ns.mutex.Lock()
	defer t.ns.mutex.Unlock()

	for _, v := range ret.Templates() {
		name := v.Name()
		tmpl := t.ns.set[name]

		if tmpl != nil {
			continue
		}

		tmpl = &Template{
			template:   v,
			leftDelim:  t.leftDelim,
			rightDelim: t.rightDelim,
			ns:         t.ns,
		}

		// Set the functions
		tmpl.Funcs(bulldozerFuncMap)

		// Add the template to the namespace
		t.ns.set[name] = tmpl
	}

	return t, nil
}

// ParseFiles parses the named files and associates the resulting templates with
// t. If an error occurs, parsing stops and the returned template is nil;
// otherwise it is t. There must be at least one file.
func (t *Template) ParseFiles(filenames ...string) (*Template, error) {
	return parseFiles(t, filenames...)
}

// ParseGlob parses the template definitions in the files identified by the
// pattern and associates the resulting templates with t. The pattern is
// processed by filepath.Glob and must match at least one file. ParseGlob is
// equivalent to calling t.ParseFiles with the list of files matched by the
// pattern.
func (t *Template) ParseGlob(pattern string) (*Template, error) {
	return parseGlob(t, pattern)
}

//##############//
//### Public ###//
//##############//

// New allocates a new bulldozer template with the given name.
func New(name string) *Template {
	t := &Template{
		template:   htmlTemplate.New(name),
		leftDelim:  "{{",
		rightDelim: "}}",
		ns:         newNameSpace(),
	}

	// Set the functions
	t.Funcs(bulldozerFuncMap)

	// Add the new template to the namespace
	t.ns.Set(t)

	return t
}

// Must is a helper that wraps a call to a function returning (*Template, error)
// and panics if the error is non-nil. It is intended for use in variable initializations
// such as
//	var t = template.Must(template.New("name").Parse("html"))
func Must(t *Template, err error) *Template {
	if err != nil {
		panic(err)
	}
	return t
}

// ParseFiles creates a new Template and parses the template definitions from
// the named files. The returned template's name will have the (base) name and
// (parsed) contents of the first file. There must be at least one file.
// If an error occurs, parsing stops and the returned *Template is nil.
func ParseFiles(filenames ...string) (*Template, error) {
	return parseFiles(nil, filenames...)
}

// ParseGlob creates a new Template and parses the template definitions from the
// files identified by the pattern, which must match at least one file. The
// returned template will have the (base) name and (parsed) contents of the
// first file matched by the pattern. ParseGlob is equivalent to calling
// ParseFiles with the list of files matched by the pattern.
func ParseGlob(pattern string) (*Template, error) {
	return parseGlob(nil, pattern)
}

//###############//
//### Private ###//
//###############//

// parseFiles is the helper for the method and function. If the argument
// template is nil, it is created from the first file.
func parseFiles(t *Template, filenames ...string) (*Template, error) {
	if len(filenames) == 0 {
		// Not really a problem, but be consistent.
		return nil, fmt.Errorf("bulldozer/template: no files named in call to ParseFiles")
	}
	for _, filename := range filenames {
		b, err := ioutil.ReadFile(filename)
		if err != nil {
			return nil, err
		}
		s := string(b)
		name := filepath.Base(filename)
		// First template becomes return value if not already defined,
		// and we use that one for subsequent New calls to associate
		// all the templates together. Also, if this file has the same name
		// as t, this file becomes the contents of t, so
		//  t, err := New(name).Funcs(xxx).ParseFiles(name)
		// works. Otherwise we create a new template associated with t.
		var tmpl *Template
		if t == nil {
			t = New(name)
		}
		if name == t.Name() {
			tmpl = t
		} else {
			tmpl = t.New(name)
		}
		_, err = tmpl.Parse(s)
		if err != nil {
			return nil, err
		}
	}
	return t, nil
}

// parseGlob is the implementation of the function and method ParseGlob.
func parseGlob(t *Template, pattern string) (*Template, error) {
	filenames, err := filepath.Glob(pattern)
	if err != nil {
		return nil, err
	}
	if len(filenames) == 0 {
		return nil, fmt.Errorf("bulldozer/template: pattern matches no files: %#q", pattern)
	}
	return parseFiles(t, filenames...)
}
