/*
 *  Bulldozer Framework
 *  Copyright (C) DesertBit
 */

package main

import (
	"fmt"
	"os/exec"
	"strings"
	"sync"
)

var (
	buildMutex sync.Mutex
)

//###############//
//### Private ###//
//###############//

func build() {
	// Lock the mutex
	buildMutex.Lock()
	defer buildMutex.Unlock()

	fmt.Printf(">  building...\n")

	// Build.
	cmd := exec.Command("go", "install")
	cmd.Dir = SrcPath
	output, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Printf(">  build error: %v\n\n=========================\n===    BUILD OUPUT    ===\n=========================\n\n%s\n\n=========================\n===  ╭∩╮（︶︿︶）╭∩╮ ===\n=========================\n\n",
			err, strings.Trim(strings.TrimSpace(string(output)), "\n"))
	} else {
		fmt.Printf(">  build success 【ツ】\n")

		// Restart the process.
		restartProcess()
	}
}
