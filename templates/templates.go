/*
 *  Bulldozer Framework
 *  Copyright (C) DesertBit
 */

package templates

import (
	tr "github.com/desertbit/bulldozer/translate"

	"fmt"
	"github.com/desertbit/bulldozer/log"
	"github.com/desertbit/bulldozer/sessions"
	"github.com/desertbit/bulldozer/template"
)

const (
	templatesUID = "budTmpls"

	// Template names:
	templateError    = "bud/error"
	templateNotFound = "bud/notfound"
)

var (
	Templates *template.Template = template.New(templatesUID, "")

	// ParseError is set if any template parsing error occurred.
	ParseError error
)

//##############//
//### Public ###//
//##############//

// Load all templates recursivly in the directory.
// The namespace is prepended to the templates names delimited by a slash.
func Load(namespace string, dir string, excludeDirs ...string) error {
	// Load all the template files recursivly.
	_, err := Templates.ParseRecToNamespace(namespace, dir, excludeDirs...)
	if err != nil &&
		err != template.ErrNoFilesFound &&
		err != template.ErrPatternMatchesNoFiles {
		// Just store the templates parse error.
		// The application startup should not be interrupted...
		ParseError = fmt.Errorf("failed to load templates: %v", err)
		return ParseError
	}

	return nil
}

// ExecNotFound executes the not found template.
// @return:
//  1: http status code
//  2: body
//  3: title
func ExecNotFound(s *sessions.Session) (int, string, string) {
	// Custom template options.
	opts := template.ExecOpts{
		StyleClasses: []string{
			"bud-sys-page",
			"bud-not-found-page",
		},
	}

	// Execute the not found page
	out, _, _, err := Templates.ExecuteTemplateToString(s, templateNotFound, opts)
	if err != nil {
		return ExecError(s, err.Error())
	}

	return 404, out, tr.S("bud.page.notFound.pageTitle")
}

// ExecError executes the error template and shows the error message if the
// user is authenticated. The error message will be also logged.
// One optional boolean can be set. If set to false, the error message won't be logged.
// @return:
//  1: http status code
//  2: body
//  3: title
func ExecError(s *sessions.Session, errorMessage string, vars ...bool) (int, string, string) {
	// Create the template data struct.
	data := struct {
		ErrorMessage string
	}{
		errorMessage,
	}

	if len(vars) <= 0 || vars[0] != false {
		// Log the error.
		log.L.Error(errorMessage)
	}

	// Custom template options.
	opts := template.ExecOpts{
		Data: data,
		StyleClasses: []string{
			"bud-sys-page",
			"bud-error-page",
		},
	}

	// Execute the template.
	out, _, _, err := Templates.ExecuteTemplateToString(s, templateError, opts)
	if err != nil {
		log.L.Error("failed to execute error core template: %v", err)
		return 500, "Internal Server Error", tr.S("bud.page.error.pageTitle")
	}

	return 500, out, tr.S("bud.page.error.pageTitle")
}
