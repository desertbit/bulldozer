/*
 *  Bulldozer Framework
 *  Copyright (C) DesertBit
 */

package log

import (
	"github.com/op/go-logging"
	"os"
)

var (
	L = logging.MustGetLogger("example")

	// Custom format string. Everything except the message has a custom color
	// which is dependent on the log level. Many fields have a custom output
	// formatting too, eg. the time returns the hour down to the milli second.
	stderrFormat = logging.MustStringFormatter(
		"%{color}%{time:15:04:05} ▶ %{level:.4s} %{id:03x}%{color:reset} %{message}",
	)

	/* Hint: syslog logging backend is currently disabled.
	   There is no syslog backend available in a docker instance.
	syslogFormat = logging.MustStringFormatter(
		"%{color}%{time:15:04:05.000000} %{shortfunc} ▶ %{level:.4s} %{id:03x}%{color:reset} %{message}",
	)
	*/
)

func init() {
	// Create the log backends.
	stderrBackend := logging.NewLogBackend(os.Stderr, "", 0)

	/*
		syslogBackend, err := logging.NewSyslogBackend("")
		if err != nil {
			L.Fatalf("failed to initialize syslog logging backend!")
		}
	*/

	// Set the custom formats.
	stderrBackendFormatter := logging.NewBackendFormatter(stderrBackend, stderrFormat)
	//syslogBackendFormatter := logging.NewBackendFormatter(syslogBackend, syslogFormat)

	// Set the backends to be used.
	logging.SetBackend(stderrBackendFormatter) //, syslogBackendFormatter)
}
