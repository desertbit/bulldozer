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

//########################//
//### Plugin Interface ###//
//########################//

type Plugin interface {
	Type() string                                    // The plugin type.
	HasSection() bool                                // If the plugin requires a template section.
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
	t      *template.Template
}

func (tp *templatePlugin) Type() string {
	return tp.plugin.Type()
}

func (tp *templatePlugin) HasSection() bool {
	return tp.plugin.HasSection()
}

func (tp *templatePlugin) Prepare(s *template.PluginSettings) (err error) {
	// Recover panics and return the error
	defer func() {
		if e := recover(); e != nil {
			err = fmt.Errorf("plugin '%s': prepare panic: %v", tp.Type(), e)
		}
	}()

	// Create the plugin data value.
	data := &Data{
		Args:    Args(s.Args),
		Section: s.Section,
	}

	// Obtain the plugin ID from the arguments.
	id, ok := data.Args["id"]
	if !ok {
		return fmt.Errorf("an unique ID is required!")
	}
	if len(id) == 0 {
		return fmt.Errorf("the passed ID is emtpy!")
	}

	// Save the ID and remove it from the arguments map.
	data.id = id
	delete(data.Args, "id")

	// Parse for core plugin arguments.
	// Check if the class argument is set.
	if classes, ok := data.Args["class"]; ok {
		// Add the style classes to the data value.
		data.additionalStyleClasses = strings.Fields(classes)

		// Remove the class argument from the argument map
		delete(data.Args, "class")
	}

	// Save the plugin data.
	s.Value = data

	// Call the plugin interface prepare method.
	tp.plugin.Prepare(data)

	return nil
}

func (tp *templatePlugin) Render(c *template.Context, s *template.PluginSettings) (r interface{}, err error) {
	// Recover panics and return the error
	defer func() {
		if e := recover(); e != nil {
			err = fmt.Errorf("render panic: %v", e)
		}
	}()

	// Cast to plugin data pointer
	data := s.Value.(*Data)

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
func Register(p Plugin) (err error) {
	// Recover panics and return the error
	defer func() {
		if e := recover(); e != nil {
			err = fmt.Errorf("failed to register plugin '%s': %v", p.Type(), e)
		}
	}()

	// Call the interface initialize method to obtain the template.
	t := p.Initialize()
	if t == nil {
		return fmt.Errorf("failed to register plugin '%s': the plugin's Initialize method returned nil!", p.Type())
	}

	// Register the plugin events to the template.
	t.RegisterEvents(p)

	// Create a new template plugin value.
	tp := &templatePlugin{
		plugin: p,
		t:      t,
	}

	// Register the plugin to the template plugins.
	template.RegisterPlugin(tp)

	return nil
}
