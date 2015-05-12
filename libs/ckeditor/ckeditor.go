/*
 *  Bulldozer Framework
 *  Copyright (C) DesertBit
 */

package ckeditor

import (
	"github.com/desertbit/bulldozer/sessions"
	"github.com/desertbit/bulldozer/settings"
)

const (
	ckEditorBaseUrl   = settings.UrlBulldozerResources + "libs/ckeditor/"
	ckEditorScriptUrl = ckEditorBaseUrl + "ckeditor.js"
)

// Load loads the required files for the CKEditor plugin.
// CKEditor's auto inline mode is disabled by default.
// This is required for the text plugin.
func Load(s *sessions.Session) {
	if s.IsJavaScriptLoaded(ckEditorScriptUrl) {
		return
	}

	// When loaded with Bulldozer.loadScript (Ajax call), CKEDITOR.basePath won't be set correctly.
	// Here's the fix:
	s.SendCommand("window.CKEDITOR_BASEPATH='" + ckEditorBaseUrl + "';")

	// Load the CKEditor javascript library if not already loaded.
	// Also disable CKEditor's auto inline mode.
	s.LoadJavaScript(ckEditorScriptUrl, "CKEDITOR.disableAutoInline = true;")
}
