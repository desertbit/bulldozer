/*
 *  Bulldozer Framework
 *  Copyright (C) DesertBit
 */

package bulldozer

import (
	_ "code.desertbit.com/bulldozer/bulldozer/plugins"
	templateStore "code.desertbit.com/bulldozer/bulldozer/template/store"
	tr "code.desertbit.com/bulldozer/bulldozer/translate"

	"code.desertbit.com/bulldozer/bulldozer/auth"
	"code.desertbit.com/bulldozer/bulldozer/backend/topbar"
	"code.desertbit.com/bulldozer/bulldozer/database"
	"code.desertbit.com/bulldozer/bulldozer/editmode"
	"code.desertbit.com/bulldozer/bulldozer/log"
	"code.desertbit.com/bulldozer/bulldozer/sessions"
	"code.desertbit.com/bulldozer/bulldozer/settings"
	"code.desertbit.com/bulldozer/bulldozer/store"
	"code.desertbit.com/bulldozer/bulldozer/template"
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
	// Parameters.
	var paramSettingsPath string

	// Set the isInitialized flag to true
	isInitialized = true

	// Bind the variables to the flags.
	flag.StringVar(&paramSettingsPath, "settings", paramSettingsPath, "set the path to an additional settings file.")

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

	// Load the settings file if it exists in the project folder.
	settingsPath := settings.Settings.WorkingPath + settings.DefaultSettingsFileName
	exists, err := utils.Exists(settingsPath)
	if err != nil {
		log.L.Fatal(err)
	}
	if exists {
		if err = settings.Load(settingsPath); err != nil {
			log.L.Fatal(err)
		}
	}

	// Load the addional settings file.
	if len(paramSettingsPath) > 0 {
		if err = settings.Load(paramSettingsPath); err != nil {
			log.L.Fatal(err)
		}
	}

	// Prepare the settings.
	err = settings.Prepare()
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

	// Initialize the bulldozer sub packages.
	sessions.Init()
	template.Init(backend)

	// Connect to the database server.
	if err = database.Connect(); err != nil {
		log.L.Fatal(err)
	}

	// Load the templates.
	if err = loadTemplates(); err != nil {
		log.L.Fatal(err)
	}

	// Initialize the store package.
	if err = store.Init(backend); err != nil {
		log.L.Fatal(err)
	}

	// Initialize the authentication package.
	if err = auth.Init(backend); err != nil {
		log.L.Fatal(err)
	}

	// Initialize the editmode package.
	if err = editmode.Init(backend); err != nil {
		log.L.Fatal(err)
	}

	// Initialize the topbar package.
	if err = topbar.Init(); err != nil {
		log.L.Fatal(err)
	}

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

	// Stop the filewatchers
	templateStore.Release()
	scssFileWatcher.Close()

	// Release the bulldozer sub packages
	sessions.Release()
	tr.Release()
	auth.Release()
	store.Release()

	// Close the database
	database.Close()

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
