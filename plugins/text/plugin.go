/*
 *  Bulldozer Framework
 *  Copyright (C) DesertBit
 */

package text

import (
	tr "code.desertbit.com/bulldozer/bulldozer/translate"
	htmlT "html/template"

	"code.desertbit.com/bulldozer/bulldozer/editmode"
	"code.desertbit.com/bulldozer/bulldozer/log"
	"code.desertbit.com/bulldozer/bulldozer/plugin"
	"code.desertbit.com/bulldozer/bulldozer/settings"
	"code.desertbit.com/bulldozer/bulldozer/store"
	"code.desertbit.com/bulldozer/bulldozer/template"
	"code.desertbit.com/bulldozer/bulldozer/ui/messagebox"
)

const (
	PluginType  = "text"
	TemplateUID = "blzPluginText"

	ckEditorBaseUrl   = settings.UrlBulldozerResources + "libs/ckeditor/"
	ckEditorScriptUrl = ckEditorBaseUrl + "ckeditor.js"
)

func init() {
	// Register the plugin.
	plugin.MustFatal(plugin.Register(new(Plugin), &plugin.Opts{
		Type:       PluginType,
		HasSection: false,
		RequireID:  true,
	}))
}

// ############## //
// ### Plugin ### //
// ############## //

type Plugin struct {
}

func (p *Plugin) Initialize() *template.Template {
	// Parse the plugin template
	t, err := template.New(TemplateUID, PluginType).Parse(templateText)
	plugin.Must(err)

	return t
}

func (p *Plugin) Prepare(d *plugin.Data) {
	// TODO: Parse settings.
}

func (p *Plugin) Render(c *template.Context, d *plugin.Data) interface{} {
	// Get the session pointer.
	s := c.Session()

	// Check if in edit mode.
	editModeActive := editmode.IsActive(s)

	if editModeActive && !s.IsJavaScriptLoaded(ckEditorScriptUrl) {
		// When loaded with Bulldozer.loadScript (Ajax call), CKEDITOR.basePath won't be set correctly.
		// Here's the fix:
		s.SendCommand("window.CKEDITOR_BASEPATH = '" + ckEditorBaseUrl + "';")

		// Load the CKEditor javascript library if not already loaded.
		// Also disable CKEditor's auto inline mode.
		s.LoadJavaScript(ckEditorScriptUrl, "CKEDITOR.disableAutoInline = true;")
	}

	// This will panic on error.
	var text string
	i, ok, err := store.Get(c)
	if err != nil {
		log.L.Error("plugin text: failed to get data from database: %v", err)
		text = tr.S("blz.plugin.text.error.getDataFromDatabase")
	}

	if !ok {
		text = tr.S("blz.plugin.text.placeholder")
	} else {
		text, ok = i.(string)
		if !ok {
			log.L.Error("plugin text: failed to cast database data to string!")
			text = tr.S("blz.plugin.text.placeholder")
		}
	}

	return struct {
		Text           htmlT.HTML
		EditModeActive bool // We won't use the buildin %.editmode.IsActive function, because we already obtained the state here. This is one method call less...
	}{
		Text:           htmlT.HTML(text),
		EditModeActive: editModeActive,
	}
}

// ############## //
// ### Events ### //
// ############## //

func (p *Plugin) EventLock(c *template.Context) {
	// Lock the context.
	if !store.Lock(c) {
		// Show a messagebox.
		messagebox.New().
			SetTitle(tr.S("blz.plugin.text.error.alreadyLockedTitle")).
			SetText(tr.S("blz.plugin.text.error.alreadyLockedText")).
			SetType(messagebox.TypeWarning).
			Show(c.Session())
		return
	}

	// Start editing.
	c.TriggerEvent("edit")
}

func (p *Plugin) EventUnlock(c *template.Context) {
	// Unlock the context again.
	store.Unlock(c)
}

func (p *Plugin) EventSetText(c *template.Context, text string) {
	err := store.Set(c, text)
	if err != nil {
		// Log the error.
		log.L.Error("plugin text: failed to set text to store: %v", err)

		// Show a messagebox.
		messagebox.New().
			SetTitle(tr.S("blz.plugin.text.error.saveChangesTitle")).
			SetText(tr.S("blz.plugin.text.error.saveChangesText")).
			SetType(messagebox.TypeWarning).
			Show(c.Session())
		return
	}
}
