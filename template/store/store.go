/*
 *  Bulldozer Framework
 *  Copyright (C) DesertBit
 */

package store

import (
	"code.desertbit.com/bulldozer/bulldozer/filewatcher"
	"code.desertbit.com/bulldozer/bulldozer/log"
	"code.desertbit.com/bulldozer/bulldozer/settings"
	"code.desertbit.com/bulldozer/bulldozer/template"
	"code.desertbit.com/bulldozer/bulldozer/utils"
	"fmt"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"
)

const (
	templatesUID = "store_"
)

var (
	changedPaths        []string
	changedPathsMutex   sync.Mutex
	reparseTemplates    func()
	templateFileWatcher *filewatcher.FileWatcher

	stores        map[int64]*Store = make(map[int64]*Store)
	storesCounter int64
	storesMutex   sync.Mutex
)

func init() {
	reparseTemplates = utils.Debounce(300*time.Millisecond, parseTemplates)
}

//##################//
//### Store type ###//
//##################//

type Store struct {
	Templates *template.Template

	key   int64
	paths []string
}

// Release releases the store from being watched.
func (s *Store) Release() {
	// Lock the mutex.
	storesMutex.Lock()
	defer storesMutex.Unlock()

	// Remove the current store from the map.
	delete(stores, s.key)
}

func (s *Store) Paths() []string {
	return s.paths
}

func (s *Store) Parse() {
	// Parse all the templates in the paths.
	var tmpls *template.Template
	var err error

	for _, path := range s.paths {
		// Create the pattern.
		pattern := path + "/*" + settings.TemplateSuffix

		// Parse the template files in in the path.
		if tmpls != nil {
			_, err = tmpls.ParseGlob(pattern)
		} else {
			tmpls, err = template.ParseGlob(templatesUID+strconv.FormatInt(s.key, 10), pattern)
		}
		if err != nil {
			// Just log this error.
			log.L.Error("failed to parse templates in path '%s': %v", path, err)
		}

	}

	// Just return if tmpls is nil.
	if tmpls == nil {
		return
	}

	// Finally set the templates pointer.
	s.Templates = tmpls
}

//##############//
//### Public ###//
//##############//

// New creates a new template store and loads all templates in the passed paths slice.
func New(paths []string) (*Store, error) {
	if paths == nil {
		return nil, fmt.Errorf("templates store: nil passed as path slice!")
	}

	for i, path := range paths {
		// Clean the path.
		paths[i] = filepath.Clean(path)

		// Check if the directory is valid.
		ok, err := utils.IsDir(path)
		if err != nil {
			return nil, fmt.Errorf("templates store: %v", err)
		}
		if !ok {
			return nil, fmt.Errorf("templates store: '%s' is not a valid directory!", path)
		}
	}

	// Create a new store.
	store := &Store{
		paths: paths,
	}

	// Lock the mutex.
	storesMutex.Lock()
	defer storesMutex.Unlock()

	// Increment the counter and set it to the store.
	storesCounter++
	store.key = storesCounter

	// Add the store to the map.
	stores[storesCounter] = store

	return store, nil
}

// Release closes the filewatcher.
func Release() {
	// Stop the filewatcher.
	if templateFileWatcher != nil {
		templateFileWatcher.Close()
	}
}

// Watch starts a filewatcher and reloads all the store templates on any change.
func Watch() {
	// Stop any previous filewatcher.
	if templateFileWatcher != nil {
		templateFileWatcher.Close()
	}

	// Create the filewatcher.
	var err error
	templateFileWatcher, err = filewatcher.New()
	if err != nil {
		log.L.Fatalf("failed to create templates filewatcher: %v", err)
	}

	// Set the event function.
	templateFileWatcher.OnEvent(onTemplatesFileChange)

	// Add the paths which should be watched.
	templateFileWatcher.Add(settings.Settings.PagesPath)
	templateFileWatcher.Add(settings.Settings.TemplatesPath)
}

//###############//
//### Private ###//
//###############//

// parseTemplates parses the templates in all store paths.
func parseTemplates() {
	log.L.Info("Loading templates...")

	// Lock the mutex.
	storesMutex.Lock()
	defer storesMutex.Unlock()

	// Get the changed paths.
	changedPathsMutex.Lock()
	paths := changedPaths
	changedPaths = nil
	changedPathsMutex.Unlock()

	if paths == nil {
		return
	}

	// Parse all the store templates.
	var found bool
	for _, store := range stores {
		// Only reload store templates, if any changed happened in the
		// specific store paths.
		found = false
		for _, storePath := range store.paths {
			for _, path := range paths {
				if strings.HasPrefix(path, storePath) {
					found = true
					break
				}
			}

			if found {
				break
			}
		}

		if found {
			store.Parse()
		}
	}
}

func onTemplatesFileChange(event *filewatcher.Event) {
	// Skip if the path is not a template file.
	if !strings.HasSuffix(event.Path, settings.TemplateSuffix) {
		return
	}

	// Add the changed path to the slice.
	changedPathsMutex.Lock()
	changedPaths = append(changedPaths, event.Path)
	changedPathsMutex.Unlock()

	// Reparse all the template files.
	reparseTemplates()
}
