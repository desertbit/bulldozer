/*
 *  Bulldozer Framework
 *  Copyright (C) DesertBit
 */

package bulldozer

import (
	"code.desertbit.com/bulldozer/bulldozer/log"
	"code.desertbit.com/bulldozer/bulldozer/sessions"
	"code.desertbit.com/bulldozer/bulldozer/settings"
	"code.desertbit.com/bulldozer/bulldozer/tr"
	"code.desertbit.com/bulldozer/bulldozer/utils"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"runtime"
	"sync"
	"time"
)

const (
	InterruptExitCode = 5
)

var (
	isInitialized bool = false
	isReleased    bool = false

	releaseMutex sync.Mutex
)

//##############//
//### Public ###//
//##############//

// Init initializes the bulldozer workspace and settings.
// If you want to set any bulldozer settings, then do this before calling this method!
func Init() {
	// Set the isInitialized flag to true
	isInitialized = true

	// Set the maximum number of CPUs that can be executing simultaneously.
	if settings.Settings.AutoSetGOMAXPROCS {
		runtime.GOMAXPROCS(runtime.NumCPU())
	}

	// Parse the flags
	if settings.Settings.AutoParseFlags {
		flag.Parse()
	}

	// Catch interrupt signals
	if settings.Settings.AutoCatchInterrupts {
		go func() {
			// Wait for the signal
			sigchan := make(chan os.Signal, 10)
			signal.Notify(sigchan, os.Interrupt)
			<-sigchan

			fmt.Println("Exiting...")

			// First cleanup
			release()

			// Exit the application
			os.Exit(InterruptExitCode)
		}()
	}

	// Check if the settings are valid
	err := settings.Check()
	if err != nil {
		log.L.Fatal(err)
	}

	// Create the important directories if they don't exist
	if err = createDirectories(); err != nil {
		log.L.Fatal(err)
	}

	// Add the translation paths and load all translation files.
	tr.Add(settings.Settings.BulldozerTranslationPath)
	tr.Add(settings.Settings.TranslationPath)
	tr.Load()

	// Load the core templates
	if err = loadCoreTemplates(); err != nil {
		log.L.Fatal(err)
	}

	// Load the templates
	parseTemplates()

	// Watch the template files and reload them on changes,
	watchTemplates()

	// Build the scss files.
	buildScss()

	// Watch the scss files and rebuild them on changes.
	watchScss()
}

// Bulldoze starts the Bulldozer server
func Bulldoze() {
	// Cleanup on exit
	defer release()

	// Check if the Init() method was called. This is a must!
	if !isInitialized {
		log.L.Fatalf("failed to bulldoze: bulldozer is not initialized! It is required to call the bulldoze.Init method before bulldozing!")
	}

	// Initialize the bulldozer sub packages
	sessions.Init()

	// Start the server
	err := serve()
	if err != nil {
		log.L.Fatal(err)
	}
}

//###############//
//### Private ###//
//###############//

func release() {
	// Lock the mutex
	releaseMutex.Lock()
	defer releaseMutex.Unlock()

	// Check if already released
	if isReleased {
		return
	}

	// Set the flag
	isReleased = true

	// Stop the filewatcher
	templateFileWatcher.Close()
	scssFileWatcher.Close()

	// Release the bulldozer sub packages
	sessions.Release()
	tr.Release()

	// Just wait for a moment before exiting to be
	// sure, all defers get called and the program
	// makes a clean exit.
	time.Sleep(150 * time.Millisecond)
}

// Create important folders if missing
func createDirectories() (err error) {
	// Create the slice of folder paths
	dirs := [...]string{
		settings.Settings.TmpPath,
		settings.Settings.PublicPath,
		settings.Settings.PagesPath,
		settings.Settings.TemplatesPath,
		settings.Settings.CoreTemplatesPath,
		settings.Settings.TranslationPath,
		settings.Settings.DataPath,
		settings.Settings.ScssPath,
	}

	// Create the directories
	for _, dir := range dirs {
		err = utils.MkDirIfNotExists(dir)
		if err != nil {
			return fmt.Errorf("failed to create directory: '%s': %v", dir, err)
		}
	}

	return nil
}
