/*
 *  Bulldozer Framework
 *  Copyright (C) DesertBit
 */

package bulldozer

import (
	"bytes"
	"code.desertbit.com/bulldozer/bulldozer/firewall"
	"code.desertbit.com/bulldozer/bulldozer/sessions"
	"code.desertbit.com/bulldozer/bulldozer/settings"
	"code.desertbit.com/bulldozer/bulldozer/utils"
	"fmt"
	"code.desertbit.com/bulldozer/bulldozer/log"
	"html/template"
	"net"
	"net/http"
	"path/filepath"
)

const (
	escapedFragment        = "_escaped_fragment_"
	postKeyInstanceID      = "id"
	reconnectDataDelimiter = "&"

	errorTemplateFilename            = "error" + settings.TemplateSuffix
	notFoundTemplateFilename         = "notfound" + settings.TemplateSuffix
	loadingIndicatorTemplateFilename = "loadingindicator" + settings.TemplateSuffix
	connectionLostTemplateFilename   = "connectionlost" + settings.TemplateSuffix
	noScriptTemplateFilename         = "noscript" + settings.TemplateSuffix
)

var (
	coreTemplate *template.Template
)

//###############//
//### Private ###//
//###############//

func loadCoreTemplates() (err error) {
	// Create the pattern string
	pattern := "*" + settings.TemplateSuffix

	// Create missing core templates in the working path
	if err = createMissingCoreTemplates(pattern); err != nil {
		return err
	}

	pattern = settings.Settings.CoreTemplatesPath + "/" + pattern

	// Create and parse the core template
	coreTemplate, err = template.New("core").Parse(htmlBody)
	if err != nil {
		return fmt.Errorf("core template parsing error: %v", err)
	}

	// Parse the templates files in the core templates directory
	coreTemplate, err = coreTemplate.ParseGlob(pattern)
	if err != nil {
		return fmt.Errorf("core templates parsing error: %v", err)
	}

	return nil
}

func createMissingCoreTemplates(pattern string) error {
	// Get all filenames of the bulldozer core templates
	coreFilenames, err := filepath.Glob(settings.Settings.BulldozerCoreTemplatesPath + "/" + pattern)
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

func serve() error {
	// Create the default html handler
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

func execNotFoundTemplate() (string, error) {
	// Execute the template
	var b bytes.Buffer
	err := coreTemplate.ExecuteTemplate(&b, notFoundTemplateFilename, nil)
	if err != nil {
		return "", err
	}

	return b.String(), nil
}

func execErrorTemplate(errorMessage string) (string, error) {
	// Create the template data struct
	data := struct {
		ErrorMessage string
	}{
		errorMessage,
	}

	// Execute the template
	var b bytes.Buffer
	err := coreTemplate.ExecuteTemplate(&b, errorTemplateFilename, data)
	if err != nil {
		return "", err
	}

	return b.String(), nil
}

func reconnectSessionFunc(rw http.ResponseWriter, req *http.Request) {
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

	// Get the instance ID from the POST query
	instanceID := req.PostFormValue(postKeyInstanceID)
	if len(instanceID) == 0 {
		http.Error(rw, "Bad Request", 400)
		return
	}

	// Create a new session object, pass the instance ID and
	// obtain the unique socket session token.
	session, accessToken, err := sessions.New(rw, req, instanceID)
	if err != nil {
		// Log the error
		log.L.Error("reconnect session error: %v", err)

		// Send an internal server error code.
		http.Error(rw, "Internal Server Error", 500)
		return
	}

	// Send the new session ID and the socket access token to the client.
	responseData := session.SessionID() + reconnectDataDelimiter + accessToken
	rw.Write([]byte(responseData))
}

func handleHtmlFunc(rw http.ResponseWriter, req *http.Request) {
	// Recover panics and log the error
	defer func() {
		if e := recover(); e != nil {
			log.L.Error("http handle panic: %v", e)
		}
	}()

	// Block to many accesses from the same remote address
	if allow, remoteAddr := firewall.NewRequest(req); !allow {
		log.L.Info("blocked incomming request from remote address '%s': too many requests", remoteAddr)
		http.Error(rw, "Too Many Requests", 429)
		return
	}

	// Check if this is a webcrawler request
	_, isWebCrawler := req.URL.Query()[escapedFragment]

	var statusCode int
	var body string

	// Create a new session object and
	// obtain the unique socket session token.
	session, accessToken, err := sessions.New(rw, req)
	if err != nil {
		// Log the error
		log.L.Error("new session error: %v", err)

		// Set the error status code and the error body
		statusCode = 500

		// Execute the error template
		body, err = execErrorTemplate("Internal Server Error")
		if err != nil {
			log.L.Error("failed to execute error core template: %v", err)
			http.Error(rw, "Internal Server Error", 500)
			return
		}
	} else {
		// Execute the route
		statusCode, body, err = execRoute(session, req.URL.Path)
		if err != nil {
			// Log the error
			log.L.Error("failed to execute route: %v", err)

			// Set the error status code and the error body
			statusCode = 500

			// Execute the error template
			body, err = execErrorTemplate("Internal Server Error")
			if err != nil {
				log.L.Error("failed to execute error core template: %v", err)
				http.Error(rw, "Internal Server Error", 500)
				return
			}
		}
	}

	// TODO: Don't load session scripts and javascripts twice if already added to the HTML head!

	// Create the template data struct
	data := struct {
		SessionID     string
		AccessToken   string
		Body          template.HTML
		JSLibs        []string
		Styles        []string
		SessionJSLibs []string
		SessionStyles []string
		IsWebCrawler  bool
	}{
		session.SessionID(),
		accessToken,
		template.HTML(body),
		settings.Settings.StaticJavaScripts,
		settings.Settings.StaticStyleSheets,
		session.JavaScripts(),
		session.StyleSheets(),
		isWebCrawler,
	}

	// Set the http status code
	rw.WriteHeader(statusCode)

	// Execute the body template
	err = coreTemplate.Execute(rw, data)
	if err != nil {
		log.L.Error("core template execution error: %v", err)
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
	{{range $style := .Styles}}
		<link rel="stylesheet" type="text/css" href="{{$style}}">
	{{end}}
	{{range $style := .SessionStyles}}
		<link rel="stylesheet" type="text/css" href="{{$style}}">
	{{end}}
	{{range $js := .JSLibs}}
		<script src="{{$js}}"></script>
	{{end}}
	{{range $js := .SessionJSLibs}}
		<script src="{{$js}}"></script>
	{{end}}
</head>
<body>
	{{if not .IsWebCrawler}}<noscript>{{template "` + noScriptTemplateFilename + `"}}</noscript>{{end}}
	<div id="bulldozer-script"><script>
		$(document).ready(function() {
			Bulldozer.socket.init("{{.SessionID}}","{{.AccessToken}}");
			$("#bulldozer-script").remove();
		});
	</script></div>
	<div id="bulldozer-loading-indicator">{{template "` + loadingIndicatorTemplateFilename + `"}}</div>
	<div id="bulldozer-connection-lost">{{template "` + connectionLostTemplateFilename + `"}}</div>
	<div id="bulldozer-body">{{.Body}}</div>
</body>
</html>`
