/*
 *  Bulldozer Framework
 *  Copyright (C) DesertBit
 */

package protect

// This plugin only shows the template section, if the session is a valid human session.

import (
	"code.desertbit.com/bulldozer/bulldozer/plugin"
	"code.desertbit.com/bulldozer/bulldozer/template"
	"fmt"
)

const (
	PluginType = "protect"

	templateUID = "budPluginProtect"
)

func init() {
	// Register the plugin.
	plugin.MustFatal(plugin.Register(new(Plugin), &plugin.Opts{
		Type:       PluginType,
		HasSection: true,
		RequireID:  false,
	}))
}

//##############//
//### Plugin ###//
//##############//

type Plugin struct {
}

func (p *Plugin) Initialize() *template.Template {
	// Parse the plugin template
	t, err := template.New(templateUID, PluginType).Parse(templateText)
	plugin.Must(err)

	return t
}

func (p *Plugin) Prepare(d *plugin.Data) interface{} {
	if len(d.Args()) > 0 {
		panic(fmt.Errorf("invalid protect plugin arguments: %v", d.Args()))
	}

	return nil
}

func (p *Plugin) Render(c *template.Context, d *plugin.Data) interface{} {
	return nil
}

//##############//
//### Events ###//
//##############//

func (p *Plugin) EventGetContent(c *template.Context) {
	// Get the plugin data.
	data, err := plugin.GetData(c)
	plugin.Must(err)

	// Send the template section to the client.
	c.TriggerEvent("setContent", data.Section())
}
