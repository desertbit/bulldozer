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
	"github.com/golang/glog"
	"html/template"
	"net"
	"net/http"
	"path/filepath"
)

const (
	errorTemplateFilename            = "error" + settings.TemplateSuffix
	notFoundTemplateFilename         = "notfound" + settings.TemplateSuffix
	loadingIndicatorTemplateFilename = "loadingindicator" + settings.TemplateSuffix
	connectionLostTemplateFilename   = "connectionlost" + settings.TemplateSuffix
	noScriptTemplateFilename         = "noscript" + settings.TemplateSuffix
)

var (
	coreTemplate *template.Template
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
}

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
		return fmt.Errorf("core template parsing error: %s", err.Error())
	}

	// Parse the templates files in the core templates directory
	coreTemplate, err = coreTemplate.ParseGlob(pattern)
	if err != nil {
		return fmt.Errorf("core templates parsing error: %s", err.Error())
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
	http.HandleFunc("/", handleHtmlFunc)

	// Serve the documents files in the document path if the settings value is set for it.
	// Another method of serving the files is to let nginx handle it.
	if settings.Settings.ServeFiles {
		http.Handle(settings.UrlBulldozerResources, http.StripPrefix(settings.UrlBulldozerResources, http.FileServer(http.Dir(settings.Settings.BulldozerResourcesPath))))
		http.Handle(settings.UrlPublic, http.StripPrefix(settings.UrlPublic, http.FileServer(http.Dir(settings.Settings.PublicPath))))
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

func handleHtmlFunc(rw http.ResponseWriter, req *http.Request) {
	// Recover panics and log the error
	defer func() {
		if e := recover(); e != nil {
			glog.Errorf("http handle panic: %v", e)
		}
	}()

	// Block to many accesses from the same remote address
	if allow, remoteAddr := firewall.NewRequest(req); !allow {
		glog.Infof("blocked incomming request from remote address '%s': too many requests", remoteAddr)
		http.Error(rw, "Too Many Requests", 429)
		return
	}

	var statusCode int
	var body string

	// Create a new session object and
	// obtain the unique socket session token.
	session, accessToken, err := sessions.New(rw, req)
	if err != nil {
		// Log the error
		glog.Errorf("new session error: %v", err)

		// Set the error status code and the error body
		statusCode = 500

		// Execute the error template
		body, err = execErrorTemplate("Internal Server Error")
		if err != nil {
			glog.Errorf("failed to execute error core template: %v", err)
			http.Error(rw, "Internal Server Error", 500)
			return
		}
	} else {
		// Execute the route
		statusCode, body, err = execRoute(req.URL.Path)
		if err != nil {
			// Log the error
			glog.Errorf("failed to execute route: %v", err)

			// Set the error status code and the error body
			statusCode = 500

			// Execute the error template
			body, err = execErrorTemplate("Internal Server Error")
			if err != nil {
				glog.Errorf("failed to execute error core template: %v", err)
				http.Error(rw, "Internal Server Error", 500)
				return
			}
		}
	}

	// Create the template data struct
	data := struct {
		SessionID   string
		AccessToken string
		Body        template.HTML
		JSLibs      []string
		Styles      []string
	}{
		session.SessionID(),
		accessToken,
		template.HTML(body),
		javaScripts,
		styleSheets,
	}

	// Set the http status code
	rw.WriteHeader(statusCode)

	// Execute the body template
	err = coreTemplate.Execute(rw, data)
	if err != nil {
		glog.Errorf("core template execution error: %v", err)
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
	{{range $js := .JSLibs}}
		<script src="{{$js}}"></script>
	{{end}}
</head>
<body>
	<noscript>{{template "` + noScriptTemplateFilename + `"}}</noscript>
	<div id="bulldozer-loading-indicator">{{template "` + loadingIndicatorTemplateFilename + `"}}</div>
	<div id="bulldozer-connection-lost">{{template "` + connectionLostTemplateFilename + `"}}</div>
	<div id="bulldozer-body">{{.Body}}</div>
	<div id="bulldozer-script"><script>
		$(document).ready(function() {
			Bulldozer.socket.init("{{.SessionID}}","{{.AccessToken}}");
			$("#bulldozer-script").remove();
		});
	</script></div>
</body>
</html>`
