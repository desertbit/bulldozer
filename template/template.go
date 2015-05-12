/*
 *  Bulldozer Framework
 *  Copyright (C) DesertBit
 */

// This template extends the golang template with a custom parse method.
// Some methods are copied from the original golang template to ensure compatibility.

package template

import (
	htmlTemplate "html/template"

	"github.com/desertbit/bulldozer/log"
	"github.com/desertbit/bulldozer/sessions"
	"github.com/desertbit/bulldozer/settings"

	"errors"
	"fmt"
	"github.com/chuckpreslar/emission"
	"io/ioutil"
	"os"
	"path/filepath"
	"sync"
)

var (
	// The global namespace.
	nameSpaces      map[string]*nameSpace = make(map[string]*nameSpace)
	nameSpacesMutex sync.Mutex

	// Custom error types.
	ErrNoFilesFound          = errors.New("bulldozer/template: no files named in call to ParseFiles")
	ErrPatternMatchesNoFiles = errors.New("bulldozer/template: pattern matches no files!")
)

//################################//
//### Private namespace struct ###//
//################################//

// nameSpace is the data structure shared by all templates in an association.
type nameSpace struct {
	// The UID is the unique ID of this template namespace.
	uid string

	// Templates map
	set   map[string]*Template
	mutex sync.Mutex
}

func newNameSpace(uid string) *nameSpace {
	// Create a new namespace
	ns := &nameSpace{
		uid: uid,
		set: make(map[string]*Template),
	}

	// Lock the mutex
	nameSpacesMutex.Lock()
	defer nameSpacesMutex.Unlock()

	// Print a error message if the UID is not unique!
	/* TODO: Sometimes a template overwrite is required and desired. Don't always show this message...
	if _, ok := nameSpaces[uid]; ok {
		log.L.Warning("template: the template UID '%s' is not unique! Overwriting already present namespace! The previous namespace is not more accessible through events,...", uid)
	}*/

	// Add the new namespace to the map
	nameSpaces[uid] = ns

	return ns
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

// GetDataFunc is called during template execution
// if no data is specified with the optional execute option.
// Return the template render data.
type GetDataFunc func(c *Context) interface{}

type Template struct {
	// We could embed the text/template field, but it's safer not to because
	// we need to keep our version of the name space and the underlying
	// template's in sync.
	template *htmlTemplate.Template

	// The namespace
	ns *nameSpace

	// Delimiters
	leftDelim, rightDelim string

	// On render function pointer
	getDataFunc GetDataFunc

	// DOM specific stuff
	staticDomID  string
	styleClasses []string

	// Events emitter
	emitter *emission.Emitter

	// Functions
	funcsMap tFuncs

	// Events
	eventsMap      map[string]*events
	eventsMapMutex sync.Mutex

	// Plugins
	pluginDataMap      pluginDataMap
	pluginDataMapUID   int64
	pluginDataMapMutex sync.Mutex

	// Must functions
	mustFuncs []*mustFunc

	globalContextID string
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
		eventsMap:  make(map[string]*events),
	}

	// Initialize the template default values.
	tt.initDefaults()

	// Add the new template to the namespace
	tt.ns.Set(tt)

	return tt
}

// OnGetData set the function which is called during template execution
// if no data is specified with the optional execute option.
// Return the template render data.
func (t *Template) OnGetData(f GetDataFunc) *Template {
	t.getDataFunc = f
	return t
}

// Lookup returns the template with the given name that is associated with t,
// or nil if there is no such template.
func (t *Template) Lookup(name string) *Template {
	// Lookup the template in the namespace
	return t.ns.Get(name)
}

// LookupFatal fatals if the template does not exists with the given name.
func (t *Template) LookupFatal(name string) (tt *Template) {
	tt = t.Lookup(name)
	if tt == nil {
		log.L.Fatalf("failed to lookup template '%s': template does not exists!", name)
	}

	return
}

// LookupMust panics if the template does not exists with the given name.
func (t *Template) LookupMust(name string) (tt *Template) {
	tt = t.Lookup(name)
	if tt == nil {
		panic(fmt.Errorf("failed to lookup template '%s': template does not exists!", name))
	}

	return
}

// Name returns the name of the template.
func (t *Template) Name() string {
	return t.template.Name()
}

// UID returns the unique ID of the template.
func (t *Template) UID() string {
	return t.ns.uid
}

// SetStaticDomID sets a static DOM ID instead of using an automatic generated one.
// The return value is the template, so calls can be chained.
func (t *Template) SetStaticDomID(id string) *Template {
	t.staticDomID = id
	return t
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
// Functions have to be registered before any template which use these functions are parsed.
// This method is equivalent of calling the html/template Funcs method.
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
func (t *Template) Parse(src string) (tt *Template, err error) {
	// Reset the plugin data.
	func() {
		t.pluginDataMapMutex.Lock()
		defer t.pluginDataMapMutex.Unlock()

		t.pluginDataMapUID = 0
		t.pluginDataMap = make(pluginDataMap)
	}()

	// Reset the must functions slice.
	t.mustFuncs = nil

	// Call the custom bulldozer parse method.
	src, err = parse(t, src, 0)
	if err != nil {
		return nil, err
	}

	// Call the html template parse method.
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
			eventsMap:  make(map[string]*events),
		}

		// Initialize the template default values.
		tmpl.initDefaults()

		// Add the template to the namespace
		t.ns.set[name] = tmpl
	}

	return t, nil
}

// ParseFiles parses the named files and associates the resulting templates with
// t. If an error occurs, parsing stops and the returned template is nil;
// otherwise it is t. There must be at least one file.
func (t *Template) ParseFiles(filenames ...string) (*Template, error) {
	return parseFiles(t.ns.uid, "", t, filenames, false)
}

// ParseGlob parses the template definitions in the files identified by the
// pattern and associates the resulting templates with t. The pattern is
// processed by filepath.Glob and must match at least one file. ParseGlob is
// equivalent to calling t.ParseFiles with the list of files matched by the
// pattern.
func (t *Template) ParseGlob(pattern string) (*Template, error) {
	return parseGlob(t.ns.uid, t, pattern)
}

// ParseRec parses all template files recursivly in the given directory path.
func (t *Template) ParseRec(dir string, excludeDirs ...string) (*Template, error) {
	return parseRec(t.ns.uid, "", t, dir, excludeDirs...)
}

// ParseRec parses all template files recursivly in the given directory path.
// The namespace is prepended to the template name.
// namespace: bulldozer
// template name: bulldozer/templatename
func (t *Template) ParseRecToNamespace(namespace string, dir string, excludeDirs ...string) (*Template, error) {
	return parseRec(t.ns.uid, namespace, t, dir, excludeDirs...)
}

// New allocates a new bulldozer template with the given name.
// The uid has to be a unique ID for this new template set.
// This uid is used to access templates over events even during
// a bulldozer application restart...
func New(uid string, name string) *Template {
	t := &Template{
		template:   htmlTemplate.New(name),
		leftDelim:  "{{",
		rightDelim: "}}",
		ns:         newNameSpace(uid),
		eventsMap:  make(map[string]*events),
	}

	// Initialize the template default values.
	t.initDefaults()

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
func ParseFiles(uid string, filenames ...string) (*Template, error) {
	return parseFiles(uid, "", nil, filenames, false)
}

// ParseGlob creates a new Template and parses the template definitions from the
// files identified by the pattern, which must match at least one file. The
// returned template will have the (base) name and (parsed) contents of the
// first file matched by the pattern. ParseGlob is equivalent to calling
// ParseFiles with the list of files matched by the pattern.
func ParseGlob(uid string, pattern string) (*Template, error) {
	return parseGlob(uid, nil, pattern)
}

// ParseRec parses all template files recursivly in the given directory path.
func ParseRec(uid string, dir string, excludeDirs ...string) (*Template, error) {
	return parseRec(uid, "", nil, dir, excludeDirs...)
}

// ParseRec parses all template files recursivly in the given directory path.
// The namespace is prepended to the template name.
// namespace: bulldozer
// template name: bulldozer/templatename
func ParseRecToNamespace(uid string, namespace string, dir string, excludeDirs ...string) (*Template, error) {
	return parseRec(uid, namespace, nil, dir, excludeDirs...)
}

// ResetEnvironment releases all registered session events.
// All previously registered client event calls will be invalid.
// If a context store is present, then this is also resetted.
func ResetEnvironment(s *sessions.Session) {
	releaseAllSessionEvents(s)
	resetContextStore(s)
}

//###############//
//### Private ###//
//###############//

func (t *Template) initDefaults() {
	// Create the map.
	t.pluginDataMap = make(pluginDataMap)

	// Set the bulldozer template functions.
	t.Funcs(bulldozerFuncMap)

	// Create a new emitter and set the recover function
	t.emitter = emission.NewEmitter().
		RecoverWith(recoverEmitter)
}

// parseFiles is the helper for the method and function. If the argument
// template is nil, it is created from the first file.
func parseFiles(uid string, namespace string, t *Template, filenames []string, recursiveName bool, pathPrefix ...string) (*Template, error) {
	if len(filenames) == 0 {
		// Not really a problem, but be consistent.
		return nil, ErrNoFilesFound
	}

	var prefix string
	if len(pathPrefix) > 0 {
		prefix = pathPrefix[0]
	}

	var tmpl *Template
	var name, errorMessage string

	for _, filename := range filenames {
		// Clean the path.
		filename = filepath.Clean(filename)

		// Get the template name.
		ext := filepath.Ext(filename)
		name = filename[0 : len(filename)-len(ext)]
		if recursiveName {
			name = filepath.ToSlash(name)
		} else {
			name = filepath.Base(name)
		}

		// Add the namespace if present.
		if len(namespace) > 0 {
			name = namespace + "/" + name
		}

		// Add the path prefix if present.
		if len(prefix) > 0 {
			filename = filepath.Clean(prefix + "/" + filename)
		}

		// Read the file.
		b, err := ioutil.ReadFile(filename)
		if err != nil {
			// Don't exit on error, because the other templates should be loaded anyway.
			// Just add the error and return it at the end.
			errorMessage += fmt.Sprint(err) + "\n"
			continue
		}

		// First template becomes return value if not already defined,
		// and we use that one for subsequent New calls to associate
		// all the templates together. Also, if this file has the same name
		// as t, this file becomes the contents of t, so
		//  t, err := New(name).Funcs(xxx).ParseFiles(name)
		// works. Otherwise we create a new template associated with t.
		if t == nil {
			t = New(uid, name)
		}
		if name == t.Name() {
			tmpl = t
		} else {
			tmpl = t.New(name)
		}

		// Parse the template.
		_, err = tmpl.Parse(string(b))
		if err != nil {
			// Don't exit on error, because the other templates should be loaded anyway.
			// Just add the error and return it at the end.
			errorMessage += fmt.Sprintf("'%s': %v", filename, err) + "\n"
			continue
		}
	}

	// Return the error if present.
	if len(errorMessage) > 0 {
		return nil, fmt.Errorf(errorMessage)
	}

	return t, nil
}

// parseGlob is the implementation of the function and method ParseGlob.
func parseGlob(uid string, t *Template, pattern string) (*Template, error) {
	filenames, err := filepath.Glob(pattern)
	if err != nil {
		return nil, err
	}
	if len(filenames) == 0 {
		return nil, ErrPatternMatchesNoFiles
	}
	return parseFiles(uid, "", t, filenames, false)
}

// parseRec is the implementation of the function and method ParseRec.
func parseRec(uid string, namespace string, t *Template, dir string, excludeDirs ...string) (*Template, error) {
	var filenames []string

	// Prepare the paths.
	dir = filepath.Clean(dir)
	for i := 0; i < len(excludeDirs); i++ {
		r, err := filepath.Rel(dir, excludeDirs[i])
		if err != nil {
			return nil, fmt.Errorf("exclude path error: %v", err)
		}

		excludeDirs[i] = r
	}

	// Get all template files recursive.
	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		r, err := filepath.Rel(dir, path)
		if err != nil {
			return err
		}

		for _, ep := range excludeDirs {
			if r == ep {
				return filepath.SkipDir
			}
		}

		// Add the file to the slice if this is a bulldozer template file.
		if filepath.Ext(r) == settings.TemplateExtension {
			filenames = append(filenames, r)
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return parseFiles(uid, namespace, t, filenames, true, dir)
}

func getNameSpace(uid string) (ns *nameSpace, ok bool) {
	// Lock the mutex
	nameSpacesMutex.Lock()
	defer nameSpacesMutex.Unlock()

	// Try to obtain the namespace
	ns, ok = nameSpaces[uid]

	return
}
