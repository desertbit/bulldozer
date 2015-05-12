/*
 *  Bulldozer Framework
 *  Copyright (C) DesertBit
 */

package bulldozer

import (
	"github.com/desertbit/bulldozer/filewatcher"
	"github.com/desertbit/bulldozer/log"
	"github.com/desertbit/bulldozer/settings"
	"github.com/desertbit/bulldozer/utils"
	"os/exec"
	"strings"
	"time"
)

var (
	rebuildScss     func()
	scssFileWatcher *filewatcher.FileWatcher
)

func init() {
	rebuildScss = utils.Debounce(300*time.Millisecond, buildScss)
}

//###############//
//### Private ###//
//###############//

// TODO: Log scss build errors to a gui page....
func buildScss() {
	log.L.Info("Building SCSS files...")

	// Create the CSS directory if not present.
	utils.MkDirIfNotExists(settings.Settings.CssPath)

	// Run the Sass build command.
	cmd := exec.Command(settings.Settings.ScssCmd, settings.Settings.ScssArgs...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		log.L.Error("scss build error: %v\n==================\n=== SCSS OUPUT ===\n==================\n%s\n==================\n==================\n",
			err, strings.Trim(strings.TrimSpace(string(output)), "\n"))
	}
}

//###################//
//### FileWatcher ###//
//###################//

func watchScss() {
	// Create the filewatcher.
	var err error
	scssFileWatcher, err = filewatcher.New()
	if err != nil {
		log.L.Fatalf("failed to create scss filewatcher: %v", err)
	}

	// Set the event function.
	scssFileWatcher.OnEvent(onScssFileChange)

	// Add the paths which should be watched.
	scssFileWatcher.Add(settings.Settings.ScssPath)
}

func onScssFileChange(event *filewatcher.Event) {
	// Skip if the path is not a scss file.
	if !strings.HasSuffix(event.Path, settings.ScssSuffix) {
		return
	}

	// Rebuild the scss files.
	rebuildScss()
}
