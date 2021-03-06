/*
 *  Bulldozer Framework
 *  Copyright (C) DesertBit
 */

package bulldozer

import (
	tr "github.com/desertbit/bulldozer/translate"

	"fmt"
	"github.com/desertbit/bulldozer/firewall"
	"github.com/desertbit/bulldozer/log"
	"github.com/desertbit/bulldozer/mux"
	"github.com/desertbit/bulldozer/sessions"
	"github.com/desertbit/bulldozer/settings"
	"html/template"
	"net"
	"net/http"
)

const (
	escapedFragment        = "_escaped_fragment_"
	postKeyInstanceID      = "id"
	reconnectDataDelimiter = "&"
	responseRequestRefresh = "req_refresh"

	// Main template files:
	loadingIndicatorTemplate = "loadingindicator" + settings.TemplateExtension
	connectionLostTemplate   = "connectionlost" + settings.TemplateExtension
	noScriptTemplate         = "noscript" + settings.TemplateExtension
)

var (
	mainTemplates *template.Template

	// The required bulldozer stylesheets.
	bulldozerStyleSheets = []string{
		settings.UrlBulldozerResources + "css/bulldozer.css",
	}

	// The required bulldozer javascripts.
	bulldozerJavaScripts = []string{
		settings.UrlBulldozerResources + "js/jquery.min.js",
		settings.UrlBulldozerResources + "js/jquery.history.js",
		settings.UrlBulldozerResources + "libs/kepler/js/vendors/fastclick/fastclick.min.js",
		settings.UrlBulldozerResources + "libs/kepler/js/kepler.min.js",
		settings.UrlBulldozerResources + "js/sha256.js",
		settings.UrlBulldozerResources + "js/bulldozer.min.js",
	}
)

func init() {
	// Create the main template.
	mainTemplates = template.New("main")

	// Add some important core functions.
	mainTemplates.Funcs(template.FuncMap{
		"tr": tr.S,
	})

	// Parse the main template body.
	_, err := mainTemplates.Parse(htmlBody)
	if err != nil {
		log.L.Fatalf("main template body parsing error: %v", err)
	}

	// Parse the additional core template files.
	pattern := settings.Settings.BulldozerCoreTemplatesPath + "/*" + settings.TemplateExtension
	_, err = mainTemplates.ParseGlob(pattern)
	if err != nil {
		log.L.Fatalf("main templates parsing error: %v", err)
	}
}

//###############//
//### Private ###//
//###############//

func serve() error {
	// Create the html handlers.
	http.HandleFunc("/bulldozer/reconnect", reconnectSessionFunc)
	http.HandleFunc("/", handleHtmlFunc)

	// Serve the documents files in the document path if the settings value is set for it.
	// Another method of serving the files is to let nginx handle it.
	if settings.Settings.ServeFiles {
		http.Handle(settings.UrlBulldozerResources, http.StripPrefix(settings.UrlBulldozerResources, http.FileServer(http.Dir(settings.Settings.BulldozerResourcesPath))))
		http.Handle(settings.UrlPublic, http.StripPrefix(settings.UrlPublic, http.FileServer(http.Dir(settings.Settings.PublicPath))))
	}

	log.L.Info("Bulldozer server listening on '%s'", settings.Settings.ListenAddress)

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

func reconnectSessionFunc(rw http.ResponseWriter, req *http.Request) {
	// If the application is currently shutting down, then
	// don't process any new requests.
	if isShuttdingDown {
		http.Error(rw, "Service Unavailable", 503)
		return
	}

	// Only allow POST requests
	if req.Method != "POST" {
		http.Error(rw, "Bad Request", 400)
		return
	}

	// Block to many accesses from the same remote address
	if allow, remoteAddr := firewall.NewRequest(req); !allow {
		log.L.Info("blocked incomming request from remote address '%s': too many requests", remoteAddr)
		http.Error(rw, "Too Many Requests", 429)
		return
	}

	// Get the instance ID from the POST query.
	instanceID := req.PostFormValue(postKeyInstanceID)
	if len(instanceID) == 0 {
		// Something wrong...
		// Tell the client to perform a complete refresh, because previous event keys are invalid.
		rw.Write([]byte(responseRequestRefresh))
		return
	}

	// Create a new session object, pass the instance ID and
	// obtain the unique socket session token.
	session, accessToken, isNewSession, err := sessions.New(rw, req, false, instanceID)
	if err != nil {
		// Log the error
		log.L.Error("reconnect session error: %v", err)

		// Send an internal server error code.
		http.Error(rw, "Internal Server Error", 500)
		return
	}

	// Set the response data.
	var responseData string
	if isNewSession {
		// This is a new session. The previous session was invalid or expired.
		// Tell the client to perform a complete refresh, because previous event keys are invalid.
		responseData = responseRequestRefresh
	} else {
		// Send the new session ID and the socket access token to the client.
		responseData = session.SessionID() + reconnectDataDelimiter + accessToken
	}

	rw.Write([]byte(responseData))
}

func handleHtmlFunc(rw http.ResponseWriter, req *http.Request) {
	// Recover panics and log the error
	defer func() {
		if e := recover(); e != nil {
			log.L.Error("http handle panic: %v", e)
		}
	}()

	// If the application is currently shutting down, then
	// don't process any new requests.
	if isShuttdingDown {
		http.Error(rw, "Service Unavailable", 503)
		return
	}

	// Block to many accesses from the same remote address
	if allow, remoteAddr := firewall.NewRequest(req); !allow {
		log.L.Info("blocked incomming request from remote address '%s': too many requests", remoteAddr)
		http.Error(rw, "Too Many Requests", 429)
		return
	}

	// Check if this is a webcrawler request
	_, isWebCrawler := req.URL.Query()[escapedFragment]

	// Create a new session object and
	// obtain the unique socket session token.
	session, accessToken, _, err := sessions.New(rw, req, isWebCrawler)
	if err != nil {
		log.L.Error("new session error: %v", err)
		http.Error(rw, "Internal Server Error", 500)
		return
	}

	// Execute the route.
	statusCode, body, title, _ := mux.ExecRoute(session, req.URL.Path)

	// Create the template data struct
	data := struct {
		Session      *sessions.Session
		AccessToken  string
		Title        string
		Body         template.HTML
		JSLibs       []string
		Styles       []string
		StaticJSLibs []string
		StaticStyles []string
		IsWebCrawler bool
	}{
		session,
		accessToken,
		title,
		template.HTML(body),
		bulldozerJavaScripts,
		bulldozerStyleSheets,
		settings.Settings.StaticJavaScripts,
		settings.Settings.StaticStyleSheets,
		isWebCrawler,
	}

	// Set the http status code
	rw.WriteHeader(statusCode)

	// Execute the main body template
	err = mainTemplates.Execute(rw, data)
	if err != nil {
		log.L.Error("main template execution error: %v", err)
	}
}

// This is the static html template body loaded only on session initialization
const htmlBody = `
<!DOCTYPE html>
<html>
<head>
	<meta charset="utf-8">
	<meta name="fragment" content="!">
	<meta name="viewport" content="width=device-width, initial-scale=1.0">
	<title>{{.Title}}</title>
	{{range $style := .Styles}}
		<link rel="stylesheet" type="text/css" href="{{$style}}">
	{{end}}
	{{range $style := .StaticStyles}}
		<link rel="stylesheet" type="text/css" href="{{$style}}">
	{{end}}
	{{range $js := .JSLibs}}
		<script src="{{$js}}"></script>
	{{end}}
	{{range $js := .StaticJSLibs}}
		<script src="{{$js}}"></script>
	{{end}}
</head>
<body>
	{{if not .IsWebCrawler}}<noscript><div id="bud-noscript">{{template "` + noScriptTemplate + `"}}</div></noscript>
	<div id="bud-script"><script>
		$(document).ready(function() {
			Bulldozer.init("{{.Session.SessionID}}","{{.AccessToken}}");
			$("#bud-script").remove();
		});
	</script></div>
	<div id="bud-loading-indicator">{{template "` + loadingIndicatorTemplate + `"}}</div>
	<div id="bud-connection-lost">{{template "` + connectionLostTemplate + `"}}</div>{{end}}
	<div id="bud-body">{{.Body}}</div>
</body>
</html>`
