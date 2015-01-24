/*
 *  Goji Framework
 *  Copyright (C) Roland Singer
 */

package plugin

import (
	htemplate "html/template"

	"bytes"
	"code.desertbit.com/bulldozer/bulldozer/log"
	"code.desertbit.com/bulldozer/bulldozer/template"
	"fmt"
	"strings"
)

type Args map[string]string

//######################//
//### Plugin options ###//
//######################//

type Opts struct {
	Type       string // The plugin type.
	HasSection bool   // If the plugin requires a template section.
	RequireID  bool   // If the plugin requires a unique ID as argument.
}

//########################//
//### Plugin Interface ###//
//########################//

type Plugin interface {
	Initialize() *template.Template                  // Called only once during plugin initialization.
	Prepare(d *Data)                                 // Prepare is called for each plugin context. Settings should be parsed...
	Render(c *template.Context, d *Data) interface{} // Render is called during each plugin template rendering request. Template render data can be returned.
}

//##########################//
//### Plugin data struct ###//
//##########################//

type Data struct {
	// Private
	id                     string
	additionalStyleClasses []string

	// Public
	Value   interface{}
	Section string
	Args    Args
}

//##############################//
//### Template plugin struct ###//
//##############################//

type templatePlugin struct {
	plugin Plugin
	opts   *Opts
	t      *template.Template
}

func (tp *templatePlugin) Prepare(d *template.PluginData) (err error) {
	// Recover panics and return the error
	defer func() {
		if e := recover(); e != nil {
			err = fmt.Errorf("plugin '%s': prepare panic: %v", tp.opts.Type, e)
		}
	}()

	// Create the plugin data value.
	data := &Data{
		Args:    Args(d.Args),
		Section: d.Section,
	}

	// Obtain the plugin ID from the arguments.
	id, ok := data.Args["id"]
	if ok {
		if len(id) == 0 {
			return fmt.Errorf("the passed ID is emtpy!")
		}

		// Remove it from the arguments map.
		delete(data.Args, "id")
	} else {
		// Return an error if the ID is required
		if tp.opts.RequireID {
			return fmt.Errorf("an unique ID is required!")
		}

		// Fallback to the plugin type string as ID.
		id = tp.opts.Type
	}

	// Set the ID.
	data.id = id

	// Parse for core plugin arguments.
	// Check if the class argument is set.
	if classes, ok := data.Args["class"]; ok {
		// Add the style classes to the data value.
		data.additionalStyleClasses = strings.Fields(classes)

		// Remove the class argument from the argument map
		delete(data.Args, "class")
	}

	// Save the plugin data.
	d.Value = data

	// Call the plugin interface prepare method.
	tp.plugin.Prepare(data)

	return nil
}

func (tp *templatePlugin) Render(c *template.Context, d *template.PluginData) (r interface{}, err error) {
	// Recover panics and return the error
	defer func() {
		if e := recover(); e != nil {
			err = fmt.Errorf("render panic: %v", e)
		}
	}()

	// Cast to plugin data pointer
	data := d.Value.(*Data)

	// Create the unique plugin template ID.
	id := c.ID() + "~" + data.id

	// Create the plugin template context.
	pContext := template.NewContext(c.Session(), tp.t, id, c.ParentID(), data.additionalStyleClasses)

	// Get the render data for the template.
	renderData := tp.plugin.Render(pContext, data)

	// Execute the plugin template.
	var b bytes.Buffer
	err = template.ExecuteContext(pContext, &b, renderData)
	if err != nil {
		return "", fmt.Errorf("failed to execute template: %v", err)
	}

	// Return the final html text.
	// By casting to a html template, we effectively unescape this portion.
	return htemplate.HTML(b.String()), nil
}

//##############//
//### Public ###//
//##############//

// Must will on error panic.
func Must(err error) {
	if err != nil {
		panic(err)
	}
}

// Must will log the error and exit the application.
func MustFatal(err error) {
	if err != nil {
		log.L.Fatal(err.Error())
	}
}

// Register registers a new plugin.
func Register(p Plugin, o *Opts) (err error) {
	if len(o.Type) == 0 {
		return fmt.Errorf("bulldozer plugin: failed to register new plugin: type string is empty!")
	}

	// Recover panics and return the error.
	defer func() {
		if e := recover(); e != nil {
			err = fmt.Errorf("failed to register plugin '%s': %v", o.Type, e)
		}
	}()

	// Call the interface initialize method to obtain the template.
	t := p.Initialize()
	if t == nil {
		return fmt.Errorf("failed to register plugin '%s': the plugin's Initialize method returned nil!", o.Type)
	}

	// Register the plugin events to the template.
	t.RegisterEvents(p)

	// Create a new template plugin value.
	tp := &templatePlugin{
		plugin: p,
		opts:   o,
		t:      t,
	}

	// Conversion.
	templateOpts := template.PluginOpts(*o)

	// Register the plugin to the template plugins.
	return template.RegisterPlugin(tp, &templateOpts)
}
