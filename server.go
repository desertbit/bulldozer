/*
 *  Bulldozer Framework
 *  Copyright (C) DesertBit
 */

package bulldozer

import (
	"code.desertbit.com/bulldozer/bulldozer/sessions"
	"code.desertbit.com/bulldozer/bulldozer/settings"
	"fmt"
	"github.com/golang/glog"
	"html/template"
	"net"
	"net/http"
)

var (
	bodyTemplate *template.Template
	javaScripts  []string
	styleSheets  []string
)

func init() {
	// Set the required stylesheets
	styleSheets = []string{
		settings.UrlBulldozerResources + "css/bulldozer.css",
	}

	// Set the required scripts
	javaScripts = []string{
		settings.UrlBulldozerResources + "js/jquery.min.js",
		settings.UrlBulldozerResources + "js/jquery.history.js",
		settings.UrlBulldozerResources + "libs/kepler/js/vendors/fastclick/fastclick.min.js",
		settings.UrlBulldozerResources + "libs/kepler/js/kepler.min.js",
		settings.UrlBulldozerResources + "js/sha256.js",
		settings.UrlBulldozerResources + "js/bulldozer.min.js",
	}

	// Create and parse the body template
	var err error
	bodyTemplate, err = template.New("body").Parse(htmlBody)
	if err != nil {
		glog.Fatalf("body template parsing error: %s", err.Error())
	}
}

//###############//
//### Private ###//
//###############//

func serve() error {
	// Create the default html handler
	http.HandleFunc("/", handleHtmlFunc)

	// Serve the documents files in the document path if the settings value is set for it.
	// Another method of serving the files is to let nginx handle it.
	if settings.Settings.ServeFiles {
		http.Handle(settings.UrlBulldozerResources, http.StripPrefix(settings.UrlBulldozerResources, http.FileServer(http.Dir(settings.Settings.BulldozerResourcesPath))))
		http.Handle(settings.UrlPublic, http.StripPrefix(settings.UrlPublic, http.FileServer(http.Dir(settings.Settings.PublicPath()))))
	}

	glog.Infof("Bulldozer server listening on '%s'", settings.Settings.ListenAddress)

	if settings.Settings.SocketType == settings.TypeUnixSocket {
		// Listen on the unix socket
		l, err := net.Listen("unix", settings.Settings.ListenAddress)
		if err != nil {
			return fmt.Errorf("Listen: %s", err.Error())
		}

		// Start the http server
		err = http.Serve(l, nil)
		if err != nil {
			return fmt.Errorf("Serve: %s", err.Error())
		}
	} else if settings.Settings.SocketType == settings.TypeTcpSocket {
		// Start the http server
		err := http.ListenAndServe(settings.Settings.ListenAddress, nil)
		if err != nil {
			return fmt.Errorf("ListenAndServe: %s", err.Error())
		}
	} else {
		return fmt.Errorf("invalid settings socket type: %s", settings.Settings.SocketType)
	}

	return nil
}

func handleHtmlFunc(rw http.ResponseWriter, req *http.Request) {
	// TODO: Block to many accesses on this function from the same IP
	// Also add this to websockets access handler

	// Create a new session object and
	// obtain the unique socket session token.
	session, accessToken, err := sessions.New(rw, req)
	if err != nil {
		glog.Errorf("new session error: %v", err)
		http.Error(rw, "Internal Server Error", 500)
		return
	}

	// Execute the route
	statusCode, body := execRoute(req.URL.Path)

	// Create the template data struct
	data := struct {
		SessionID        string
		AccessToken      string
		Body             string
		JSLibs           []string
		Styles           []string
		LoadingIndicator template.HTML
	}{
		session.SessionID(),
		accessToken,
		body,
		javaScripts,
		styleSheets,
		template.HTML(LoadingIndicator),
	}

	// Set the http status code
	rw.WriteHeader(statusCode)

	// Execute the body template
	bodyTemplate.Execute(rw, data)
}

// TODO: Remove the script tag on success

// This is the static html template body loaded only on session initialization
const htmlBody = `
<!DOCTYPE html>
<html>
<head>
	<meta charset="utf-8">
	<meta name="fragment" content="!">
	<meta name="viewport" content="width=device-width, initial-scale=1.0">
	{{range $style := .Styles}}
		<link rel="stylesheet" type="text/css" href="{{$style}}">
	{{end}}
	{{range $js := .JSLibs}}
		<script src="{{$js}}"></script>
	{{end}}
</head>
<body>
	{{.LoadingIndicator}}
	<div id="bulldozer-body">{{.Body}}</div>
	<div id="bulldozer-script"><script>
		$(document).ready(function() {
			Bulldozer.socket.init("{{.SessionID}}","{{.AccessToken}}");
			$("#bulldozer-script").remove();
		});
	</script></div>
	<noscript>
		<div class="">
			<h1>Oops!</h1>
			<h2>No Javascript enabled!</h2>
			<div class="error-details">
				Please enable JavaScript to load this page!
			</div>
		</div>
	</noscript>
</body>
</html>`
