/*
 *  Bulldozer Framework
 *  Copyright (C) DesertBit
 */

package main

import (
	"code.desertbit.com/bulldozer/bulldozer/filewatcher"
	"code.desertbit.com/bulldozer/bulldozer/utils"
	"log"
	"path/filepath"
	"strings"
	"time"
)

const (
	GoSuffix       = ".go"
	TemplateSuffix = ".bt"
)

const (
	actionNothing int = 1 << iota
	actionRestart int = 1 << iota
	actionRebuild int = 1 << iota
)

var (
	restartProcessOnChangedFiles []string
	fileWatcher                  *filewatcher.FileWatcher

	action         int = actionNothing
	debounceAction func()
)

//###############//
//### Private ###//
//###############//

func init() {
	debounceAction = utils.Debounce(500*time.Millisecond, performAction)
}

func performAction() {
	if action == actionRestart {
		restartProcess()
	} else if action == actionRebuild {
		build()
	}

	action = actionNothing
}

//###################//
//### FileWatcher ###//
//###################//

func addPathToProcessRestartTrigger(path string) {
	// First clean the path.
	path = filepath.Clean(path)

	// Add it to the slice
	restartProcessOnChangedFiles = append(restartProcessOnChangedFiles, path)
}

func watchFiles() {
	// Create the filewatcher.
	var err error
	fileWatcher, err = filewatcher.New()
	if err != nil {
		log.Fatalf("failed to create filewatcher: %v", err)
	}

	// Set the event function.
	fileWatcher.OnEvent(onFileChange)

	// Add the path which should be watched.
	fileWatcher.Add(WatchPath)
}

func onFileChange(event *filewatcher.Event) {
	// Get the path.
	path := filepath.Clean(event.Path)

	// Restart the application if any file changed which
	// matches the defined slice.
	for _, p := range restartProcessOnChangedFiles {
		if p == path {
			// Restart the process.
			if action != actionRebuild {
				action = actionRestart
			}
			debounceAction()
			return
		}
	}

	// Restart the application if any template file changed.
	if strings.HasSuffix(path, TemplateSuffix) {
		// Restart the process.
		if action != actionRebuild {
			action = actionRestart
		}
		debounceAction()
		return
	}

	// Rebuild the application if any go file changed.
	if strings.HasSuffix(path, GoSuffix) {
		// Rebuild the bulldozer application.
		action = actionRebuild
		debounceAction()
		return
	}
}
