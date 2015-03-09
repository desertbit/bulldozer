/*
 *  Bulldozer Framework
 *  Copyright (C) DesertBit
 */

package text

import (
	tr "code.desertbit.com/bulldozer/bulldozer/translate"
	htmlT "html/template"

	"code.desertbit.com/bulldozer/bulldozer/editmode"
	"code.desertbit.com/bulldozer/bulldozer/libs/ckeditor"
	"code.desertbit.com/bulldozer/bulldozer/log"
	"code.desertbit.com/bulldozer/bulldozer/plugin"
	"code.desertbit.com/bulldozer/bulldozer/store"
	"code.desertbit.com/bulldozer/bulldozer/template"
	"code.desertbit.com/bulldozer/bulldozer/ui/messagebox"

	"fmt"
	"strings"
)

const (
	PluginType = "text"

	templateUID          = "budPluginText"
	emptyTextPlaceholder = "<p><br></p>"
)

const (
	ModeFull    = "full"
	ModeDefault = "default"
	ModeMinimal = "minimal"
	ModePlain   = "plain"
)

func init() {
	// Register the plugin.
	plugin.MustFatal(plugin.Register(new(Plugin), &plugin.Opts{
		Type:       PluginType,
		HasSection: false,
		RequireID:  true,
	}))
}

//#######################//
//### Plugin Settings ###//
//#######################//

type Settings struct {
	Mode    string
	Protect bool
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
	// Default settings
	s := &Settings{
		Mode: ModeDefault,
	}

	// Parse the arguments
	for k, v := range d.Args() {
		switch k {
		case "mode":
			// Set the editor mode
			switch v {
			case ModeFull:
				s.Mode = ModeFull
			case ModeMinimal:
				s.Mode = ModeMinimal
			case ModePlain:
				s.Mode = ModePlain
			case ModeDefault:
				s.Mode = ModeDefault
			default:
				panic(fmt.Errorf("invalid text plugin mode argument: %s", v))
			}
		case "protect":
			switch v {
			case "true":
				s.Protect = true
			case "false":
				s.Protect = false
			default:
				panic(fmt.Errorf("invalid text plugin protect argument: %s", v))
			}
		default:
			panic(fmt.Errorf(`invalid text plugin argument: %s="%s"`, k, v))
		}
	}

	return s
}

func (p *Plugin) Render(c *template.Context, d *plugin.Data) interface{} {
	// Get the settings.
	settings := d.Value().(*Settings)

	// Get the session pointer.
	s := c.Session()

	// Check if in edit mode.
	editModeActive := editmode.IsActive(s)

	if editModeActive {
		// Load the CKEditor javascript library if not already loaded.
		ckeditor.Load(s)
	}

	// Get the text.
	text := getText(c)

	return struct {
		Text           htmlT.HTML
		EditModeActive bool // We won't use the buildin %.editmode.IsActive function, because we already obtained the state here. This is one method call less...
		Mode           string
		Protect        bool
	}{
		Text:           htmlT.HTML(text),
		EditModeActive: editModeActive,
		Mode:           settings.Mode,
		Protect:        settings.Protect,
	}
}

//##############//
//### Events ###//
//##############//

func (p *Plugin) EventLock(c *template.Context) {
	// Lock the context.
	if !store.Lock(c) {
		// Notify the client.
		c.TriggerEvent("lockFailed")

		// Show a messagebox.
		messagebox.New().
			SetTitle(tr.S("bud.plugin.text.error.alreadyLockedTitle")).
			SetText(tr.S("bud.plugin.text.error.alreadyLockedText")).
			SetType(messagebox.TypeWarning).
			Show(c.Session())
		return
	}
}

func (p *Plugin) EventUnlock(c *template.Context) {
	// Unlock the context again.
	store.Unlock(c)
}

func (p *Plugin) EventSetText(c *template.Context, text string) {
	// Set the empty text placeholder if the text is empty.
	if len(strings.TrimSpace(text)) == 0 {
		text = emptyTextPlaceholder
	}

	err := store.Set(c, text)
	if err != nil {
		// Log the error.
		log.L.Error("plugin text: failed to set text to store: %v", err)

		// Show a messagebox.
		messagebox.New().
			SetTitle(tr.S("bud.plugin.text.error.saveChangesTitle")).
			SetText(tr.S("bud.plugin.text.error.saveChangesText")).
			SetType(messagebox.TypeWarning).
			Show(c.Session())
		return
	}
}

func (p *Plugin) EventGetProtectedData(c *template.Context) {
	// Get the text.
	text := getText(c)

	// Send the text to the client.
	c.TriggerEvent("setProtectedData", text)
}

//###############//
//### Private ###//
//###############//

func getText(c *template.Context) (text string) {
	i, ok, err := store.Get(c)
	if err != nil {
		log.L.Error("plugin text: failed to get data from database: %v", err)
		text = tr.S("bud.plugin.text.error.getDataFromDatabase")
	}

	if !ok {
		text = tr.S("bud.plugin.text.placeholder")
	} else {
		text, ok = i.(string)
		if !ok {
			log.L.Error("plugin text: failed to cast database data to string!")
			text = tr.S("bud.plugin.text.placeholder")
		}
	}

	return
}
