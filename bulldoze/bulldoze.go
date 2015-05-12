/*
 *  Bulldozer Framework
 *  Copyright (C) DesertBit
 */

package main

import (
	"code.desertbit.com/bulldozer/bulldozer/log"
	"code.desertbit.com/bulldozer/bulldozer/utils"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"runtime"
	"sync"
	"syscall"
)

const (
	bulldozerGoPath = "src/code.desertbit.com/bulldozer/bulldozer/"
)

var (
	GoPath                  string
	SrcPath                 string
	WatchPath               string
	BulldozerPrototypesPath string
	ProcessCmd              string
	ProcessArgs             []string

	quitChan       chan struct{} = make(chan struct{})
	quitChanClosed bool          = false
	quitMutex      sync.Mutex
)

func main() {
	// Set the maximum number of CPUs that can be executing simultaneously.
	runtime.GOMAXPROCS(runtime.NumCPU())

	defer release()

	// Catch interrupt signals
	go func() {
		// Wait for the signal
		sigchan := make(chan os.Signal, 10)
		signal.Notify(sigchan, os.Interrupt, os.Kill, syscall.SIGTERM, syscall.SIGKILL)
		<-sigchan

		// Exit the main loop, so all defers get called.
		quit()
	}()

	// Initialize the settings.
	initSettings()

	// Create missing project files.
	createMissingProjectFiles()

	// Build and run the bulldozer application.
	build()

	// Start watching for changes.
	watchFiles()

	// Just pause the main loop.
	<-quitChan
}

func release() {
	// Release the filewatcher.
	fileWatcher.Close()

	// Stop the process.
	stopProcess()
}

func quit() {
	// Lock the mutex.
	quitMutex.Lock()
	defer quitMutex.Unlock()

	// Check if already closed.
	if quitChanClosed {
		return
	}

	fmt.Println("Exiting...")

	// Set the flag and close the channel.
	quitChanClosed = true
	close(quitChan)
}

func initSettings() {
	// Set the GOPATH.
	GoPath = os.Getenv("GOPATH")
	if len(GoPath) == 0 {
		log.L.Fatal("GOPATH is not set!")
	}
	GoPath = utils.AddTrailingSlashToPath(GoPath)

	BulldozerPrototypesPath = GoPath + bulldozerGoPath + "data/prototypes"

	// Get the current working directory path.
	var err error
	SrcPath, err = os.Getwd()
	if err != nil {
		log.L.Fatal("failed to obtain current work directory path: %v", err)
	}

	// Flag if the complete GOPATH should be watched for changes.
	watchGoPath := false

	// Bind the variables to the flags.
	flag.StringVar(&SrcPath, "src", SrcPath, "set the project source path, instead of using the current directory.")
	flag.BoolVar(&watchGoPath, "watchall", watchGoPath, "whenever the complete GOPATH should be watched for changes.")

	// Parse the flags.
	flag.Parse()

	// Set the process arguments.
	// They are all arguments after --
	ProcessArgs = flag.Args()

	// Prepare the source paths.
	SrcPath = filepath.Clean(SrcPath)

	// Set the watch path
	if watchGoPath {
		WatchPath = GoPath + "src"
	} else {
		WatchPath = SrcPath
	}
	WatchPath = filepath.Clean(WatchPath)

	// Set the process command.
	ProcessCmd = GoPath + "bin" + "/" + filepath.Base(SrcPath)
}

func createMissingProjectFiles() {
	// Just be sure, a trailing slash is there.
	srcPath := utils.AddTrailingSlashToPath(SrcPath)
	prototypesPath := utils.AddTrailingSlashToPath(BulldozerPrototypesPath)

	// Restart the process if following files change.
	addPathToProcessRestartTrigger(srcPath + "settings.toml")

	// We use a map here to store the source and destination strings.
	filesToCopy := map[string]string{
		prototypesPath + "settings.toml": srcPath + "settings.toml",
		prototypesPath + "main.go":       srcPath + "main.go",
	}

	// Copy all missing files.
	for src, dest := range filesToCopy {
		exists, err := utils.Exists(dest)
		if err != nil {
			log.L.Fatal(err.Error())
		}
		if !exists {
			// Copy the file.
			if err = utils.CopyFileIfNotExists(src, dest); err != nil {
				log.L.Fatalf("failed to copy file '%s' to '%s': %v", src, dest, err)
			}
		}
	}
}

/*
func createMissingCoreTemplates() error {
	// Get all filenames of the bulldozer core templates
	coreFilenames, err := filepath.Glob(settings.Settings.BulldozerCoreTemplatesPath + "/*" + settings.TemplateSuffix)
	if err != nil {
		return err
	}
	if len(coreFilenames) == 0 {
		return nil
	}

	// Create missing template files
	for _, src := range coreFilenames {
		// Create the destination path
		dest := settings.Settings.CoreTemplatesPath + "/" + filepath.Base(src)

		// Copy the file if it doesn't exists
		if err = utils.CopyFileIfNotExists(src, dest); err != nil {
			return fmt.Errorf("failed to copy core template '%s' to '%s': %v", src, dest, err)
		}
	}

	return nil
}

*/
