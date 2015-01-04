/*
 *  Bulldozer Framework
 *  Copyright (C) DesertBit
 */

package bulldozer

import (
	"code.desertbit.com/bulldozer/bulldozer/filewatcher"
	"code.desertbit.com/bulldozer/bulldozer/settings"
	"code.desertbit.com/bulldozer/bulldozer/template"
	"code.desertbit.com/bulldozer/bulldozer/utils"
	"github.com/golang/glog"
	"strings"
	"time"
)

const (
	templatesUID = "tmpls"
)

var (
	templates           *template.Template
	reparseTemplates    func()
	templateFileWatcher *filewatcher.FileWatcher
)

func init() {
	reparseTemplates = utils.Debounce(300*time.Millisecond, parseTemplates)
}

//###############//
//### Private ###//
//###############//

func parseTemplates() {
	glog.Infof("Loading templates...")

	// Create the pattern strings.
	patternPages := settings.Settings.PagesPath + "/" + "*" + settings.TemplateSuffix
	patternTemplates := settings.Settings.TemplatesPath + "/" + "*" + settings.TemplateSuffix

	// Parse the template files in the pages directory.
	tmpls, err := template.ParseGlob(templatesUID, patternPages)
	if err != nil {
		// Just log this error.
		glog.Errorf("failed to parse page templates: %v", err)
	}

	// Parse the template files in the templates directory.
	if tmpls != nil {
		_, err = tmpls.ParseGlob(patternTemplates)
	} else {
		tmpls, err = template.ParseGlob(templatesUID, patternTemplates)
	}
	if err != nil {
		// Just log this error.
		glog.Errorf("failed to parse templates: %v", err)
	}

	// Finally set the templates pointer.
	templates = tmpls
}

//###################//
//### FileWatcher ###//
//###################//

func watchTemplates() {
	// Create the filewatcher.
	var err error
	templateFileWatcher, err = filewatcher.New()
	if err != nil {
		glog.Fatalf("failed to create templates filewatcher: %v", err)
	}

	// Set the event function.
	templateFileWatcher.OnEvent(onTemplatesFileChange)

	// Add the paths which should be watched.
	templateFileWatcher.Add(settings.Settings.PagesPath)
	templateFileWatcher.Add(settings.Settings.TemplatesPath)
}

func onTemplatesFileChange(event *filewatcher.Event) {
	// Skip if the path is not a template file.
	if !strings.HasSuffix(event.Path, settings.TemplateSuffix) {
		return
	}

	// Reparse all the template files.
	reparseTemplates()
}
