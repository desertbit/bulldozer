/*
 *  Bulldozer Framework
 *  Copyright (C) DesertBit
 */

package controlcenter

/*
import (
	"code.desertbit.com/bulldozer/bulldozer/plugin"
	"code.desertbit.com/bulldozer/bulldozer/settings"
	"code.desertbit.com/bulldozer/bulldozer/template"
)

const (
	PluginType = "controlcenter"

	templateUID  = "budControlC"
	templatesDir = "controlcenter/"

	// Template names:
	controlcenterTemplate = "controlcenter" + settings.TemplateExtension
)


func init() {
	// Register the plugin.
	plugin.MustFatal(plugin.Register(new(Plugin), &plugin.Opts{
		Type: PluginType,
	}))
}

//##############//
//### Plugin ###//
//##############//

type Plugin struct {
}

func (p *Plugin) Initialize() *template.Template {
	// Create the file path.
	file := settings.GetCoreTemplatePath(templatesDir + controlcenterTemplate)

	// Parse the plugin template
	t, err := template.ParseFiles(templateUID, file)
	plugin.Must(err)

	return t
}

func (p *Plugin) Prepare(d *plugin.Data) interface{} {
	return nil
}

func (p *Plugin) Render(c *template.Context, d *plugin.Data) interface{} {
	return nil
}
*/
