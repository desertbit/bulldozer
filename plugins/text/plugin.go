/*
 *  Bulldozer Framework
 *  Copyright (C) DesertBit
 */

package text

// TODO: Add database stuff and the CKEDITOR.

import (
	"code.desertbit.com/bulldozer/bulldozer/plugin"
	"code.desertbit.com/bulldozer/bulldozer/template"
	"strings"
)

const (
	TemplateUID = "plugin-text"
)

func init() {
	// Register the plugin.
	plugin.MustFatal(plugin.Register(new(Plugin), &plugin.Opts{
		Type:       "text",
		HasSection: true,
		RequireID:  false,
	}))
}

// ############## //
// ### Plugin ### //
// ############## //

type Plugin struct {
}

func (p *Plugin) Initialize() *template.Template {
	// Parse the plugin template
	t, err := template.New(TemplateUID, "text").Parse(templateText)
	plugin.Must(err)

	return t
}

func (p *Plugin) Prepare(d *plugin.Data) {
	d.Section = strings.TrimSpace(d.Section)
}

func (p *Plugin) Render(c *template.Context, d *plugin.Data) interface{} {
	return d.Section
}
