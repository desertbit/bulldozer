/*
 *  Bulldozer Framework
 *  Copyright (C) DesertBit
 */

package tr

import (
	"bufio"
	"encoding/json"
	"fmt"
	"github.com/desertbit/bulldozer/log"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

const (
	Suffix = ".tr"

	defaultLocale      = "en"
	reloadTimerTimeout = 1 * time.Second
)

var (
	currentLocale string            = defaultLocale
	messages      map[string]string = make(map[string]string)
	directories   []string
	mutex         sync.Mutex

	reloadTimer     *time.Timer
	reloadTimerStop chan struct{} = make(chan struct{})
)

func init() {
	reloadTimer = time.NewTimer(reloadTimerTimeout)

	startReloadLoop()
}

//##############//
//### Loops ####//
//##############//

func startReloadLoop() {
	go func() {
		defer func() {
			// Stop the timer
			reloadTimer.Stop()
		}()

		for {
			select {
			case <-reloadTimer.C:
				// Reload the translations
				_reload()
			case <-reloadTimerStop:
				// Just exit the loop
				return
			}
		}
	}()
}

//##############//
//### Public ###//
//##############//

// Load all translation files.
// If already loaded, they will be reloaded.
func Load() {
	// Stop the timer
	reloadTimer.Stop()

	// Reload the translations
	_reload()
}

// Release releases all goroutines and performs a cleanup
func Release() {
	// Stop the filewatcher
	fileWatcher.Close()

	// Stop the reload goroutine
	close(reloadTimerStop)
}

// SetLocale sets the current locale.
// The translations files are all reloaded.
func SetLocale(locale string) {
	// Lock the mutex
	mutex.Lock()
	defer mutex.Unlock()

	// Set the new locale
	currentLocale = locale

	// Reload the translation messages
	reload()
}

// S obtains the translated string for the given ID.
func S(id string, args ...interface{}) string {
	// Try to get the translated string for the ID
	s, ok := getMessage(id)

	if !ok {
		log.L.Warning("no translated string found for ID '%s'", id)
		return "???" + id + "???"
	}

	return fmt.Sprintf(s, args...)
}

// Add adds a translation directory path
func Add(dirPath string) {
	// Lock the mutex
	mutex.Lock()
	defer mutex.Unlock()

	// Clean the path
	dirPath = filepath.Clean(dirPath)

	// Check if the path is already in the slice
	for _, p := range directories {
		if p == dirPath {
			// The path is already present.
			// Just return...
			return
		}
	}

	// Add the new path to the slice
	directories = append(directories, dirPath)

	// Add the new path to the filewatcher
	err := fileWatcher.Add(dirPath)
	if err != nil {
		log.L.Error("translation: failed to add path '%s' to file watcher: %v", dirPath, err)
	}

	// Reload the translation messages
	reload()
}

// Remove removes the translation directory path
func Remove(dirPath string) {
	// Lock the mutex
	mutex.Lock()
	defer mutex.Unlock()

	// Clean the path
	dirPath = filepath.Clean(dirPath)

	// Find the index position of the path
	index := -1
	for i, p := range directories {
		if p == dirPath {
			index = i
			break
		}
	}

	// Return if the path was not found
	if index < 0 || index >= len(directories) {
		return
	}

	// Remove the entry
	directories[index], directories = directories[len(directories)-1], directories[:len(directories)-1]

	// Remove the path again from the filewatcher
	fileWatcher.Remove(dirPath)

	// Reload the translation messages
	reload()
}

//###############//
//### Private ###//
//###############//

func getMessage(id string) (s string, ok bool) {
	// Lock the mutex
	mutex.Lock()
	defer mutex.Unlock()

	s, ok = messages[id]
	return
}

func reload() {
	// Reset the timer.
	reloadTimer.Reset(reloadTimerTimeout)
}

func _reload() {
	// Lock the mutex
	mutex.Lock()
	defer mutex.Unlock()

	log.L.Info("translation: reloading translation files")

	// Empty the current messages map
	messages = make(map[string]string)

	// Go through all directories
	for _, d := range directories {
		// Skip if the base translation directory contains no directories.
		entries, err := ioutil.ReadDir(d)
		if err != nil {
			log.L.Error("failed to obtain entry list of directory '%s': %v", d, err)
			continue
		}
		empty := true
		for _, e := range entries {
			if e.IsDir() {
				empty = false
				break
			}
		}
		if empty {
			continue
		}

		dirPath := d + "/" + currentLocale

		// Check if the current locale translation folder exists
		ok, err := dirExists(dirPath)
		if err != nil {
			log.L.Error("translate reload error: %v", err)
			continue
		}

		// If not try to load the default locale
		if !ok {
			log.L.Warning("translation: missing translation files '%s' for current locale '%s'", d, currentLocale)

			dirPath = d + "/" + defaultLocale

			// Check if the default locale translation folder exists
			ok, err := dirExists(dirPath)
			if err != nil {
				log.L.Error("translate reload error: %v", err)
				continue
			}
			if !ok {
				log.L.Error("translation: missing translation files '%s' for default locale '%s'", d, defaultLocale)
				continue
			}
		}

		err = filepath.Walk(dirPath, loadFile)
		if err != nil {
			log.L.Error("translation: filepath walk error: %v", err)
		}
	}
}

func dirExists(path string) (bool, error) {
	// Get the state
	stat, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}

		return false, err
	}

	// Check if it is a directory
	return stat.IsDir(), nil
}

func loadFile(path string, f os.FileInfo, err error) error {
	if f == nil {
		log.L.Error("filepath walk: file info object is nil!")
		return nil
	}

	// Skip if this is a directory or the path is missing the translation suffix
	if f.IsDir() || !strings.HasSuffix(path, Suffix) {
		return nil
	}

	type Message struct {
		ID, Text string
	}

	// Open the file
	file, err := os.Open(path)
	if err != nil {
		log.L.Error("failed to open translation file '%s': %v", path, err)
		return nil
	}
	defer file.Close()

	// Parse the JSON translation file
	var ok bool
	dec := json.NewDecoder(bufio.NewReader(file))
	for {
		var m Message
		if err = dec.Decode(&m); err == io.EOF {
			break
		} else if err != nil {
			log.L.Error("failed to parse translation file '%s': %v", path, err)
			return nil
		}

		// Check if overwriting the value
		_, ok = messages[m.ID]
		if ok {
			log.L.Warning("%s: overwriting duplicate translation message with ID '%s'!", path, m.ID)
		}

		// Add the text to the message map
		messages[m.ID] = m.Text
	}

	return nil
}
