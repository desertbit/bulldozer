/*
 *  Bulldozer Framework
 *  Copyright (C) DesertBit
 */

package tr

import (
	"code.desertbit.com/bulldozer/bulldozer/filewatcher"
	"github.com/golang/glog"
	"strings"
)

var (
	fileWatcher *filewatcher.FileWatcher
)

func init() {
	// Create the filewatcher
	var err error
	fileWatcher, err = filewatcher.New()
	if err != nil {
		glog.Fatalf("failed to create translate filewatcher: %v", err)
	}

	// Set the event function
	fileWatcher.OnEvent(onFileChange)
}

func onFileChange(event *filewatcher.Event) {
	ok, err := dirExists(event.Path)
	if err != nil {
		glog.Errorln(err)
		return
	}

	// Skip if the path is not a directory and is not a translation file
	if !ok && !strings.HasSuffix(event.Path, Suffix) {
		return
	}

	// Reload the translation messages
	reload()
}
