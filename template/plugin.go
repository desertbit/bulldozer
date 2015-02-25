/*
 *  Bulldozer Framework
 *  Copyright (C) DesertBit
 */

package template

import (
	tr "code.desertbit.com/bulldozer/bulldozer/translate"

	"code.desertbit.com/bulldozer/bulldozer/log"
	"code.desertbit.com/bulldozer/bulldozer/utils"
	"fmt"
	"strconv"
)

var (
	plugins map[string]*pluginWrapper = make(map[string]*pluginWrapper)
)

//####################//
//### Plugin types ###//
//####################//

type pluginWrapper struct {
	i    Plugin
	opts *PluginOpts
}

type Plugin interface {
	Prepare(d *PluginData) error // Prepare is called for each plugin context. Settings should be parsed...
	Render(c *Context, d *PluginData) (interface{}, error)
}

type PluginData struct {
	Value   interface{}
	Args    Args
	Section string
}

type PluginOpts struct {
	Type       string // The plugin type.
	HasSection bool   // If the plugin requires a template section.
	RequireID  bool   // If the plugin requires a unique ID as argument.
}

//##########################//
//### Plugin data struct ###//
//##########################//

type pluginDataMap map[int64]*pluginData

type pluginData struct {
	plugin *pluginWrapper
	data   *PluginData
}

//##############//
//### Public ###//
//##############//

// RegisterPlugin adds and registers a template plugin with the given type.
// This method call is not thread-safe!
func RegisterPlugin(p Plugin, opts *PluginOpts) error {
	// Just be sure.
	if opts == nil {
		return fmt.Errorf("bulldozer template: plugin: failed to register new plugin: options value is nil!")
	} else if len(opts.Type) == 0 {
		return fmt.Errorf("bulldozer template: plugin: failed to register new plugin: type string is empty!")
	}

	// Print a warning if a previous plugin is overwritten.
	_, ok := plugins[opts.Type]
	if ok {
		log.L.Error("bulldozer template: plugin: overwritting already set plugin '%s'!", opts.Type)
	}
	// Add the plugin wrapper to the map.
	plugins[opts.Type] = &pluginWrapper{
		i:    p,
		opts: opts,
	}

	// Register the template parse function for the given plugin type.
	registerParseFunc(opts.Type, parsePlugin)

	return nil
}

// GetPlugin returns the plugin defined by the type.
// if not found, nil is returned.
func GetPlugin(typeStr string) Plugin {
	p, ok := plugins[typeStr]
	if !ok {
		return nil
	}

	return p.i
}

//###############//
//### Private ###//
//###############//

func parsePlugin(typeStr string, token string, d *parseData) (err error) {
	// Recover panics and log the error
	defer func() {
		if e := recover(); e != nil {
			err = fmt.Errorf("parse plugin panic: %v", e)
		}
	}()

	// Try to get the plugin wrapper.
	plugin, ok := plugins[typeStr]
	if !ok {
		return fmt.Errorf("plugin: no plugin exists with name '%s'", typeStr)
	}

	// Get the plugin arguments.
	args, err := getArgs(token)
	if err != nil {
		return fmt.Errorf("plugin: failed to parse plugin arguments '%s': %v", token, err)
	}

	// Create a new plugin data value.
	data := &pluginData{
		plugin: plugin,
		data: &PluginData{
			Args: args,
		},
	}

	// Get the section if required.
	if plugin.opts.HasSection {
		data.data.Section, err = getSection(typeStr, d.src, d)
		if err != nil {
			return fmt.Errorf("invalid plugin '%s' syntax! Missing end tag {{end %s}}", typeStr, typeStr)
		}
	}

	// Prepare the plugin data.
	err = plugin.i.Prepare(data.data)
	if err != nil {
		return fmt.Errorf("failed to prepare plugin '%s': %v", typeStr, err)
	}

	// Get the template pointer.
	t := d.t

	// Lock the mutex.
	t.pluginDataMapMutex.Lock()
	defer t.pluginDataMapMutex.Unlock()

	// Create a new unique plugin key.
	t.pluginDataMapUID++
	uid := t.pluginDataMapUID

	// Add the new plugin data to the map.
	t.pluginDataMap[uid] = data

	*d.final += `{{plugin $.Context ` + strconv.FormatInt(uid, 10) + `}}`

	return nil
}

func renderPlugin(c *Context, uid int64) (r interface{}) {
	// Recover panics and log the error
	defer func() {
		if e := recover(); e != nil {
			log.L.Error("render plugin panic: %v", e)
			r = utils.ErrorBox(tr.S("bud.template.plugin.error"), e)
		}
	}()

	var err error

	// Get the template pointer.
	t := c.t

	// Get the plugin data value.
	data, ok := func() (data *pluginData, ok bool) {
		// Lock the mutex.
		t.pluginDataMapMutex.Lock()
		defer t.pluginDataMapMutex.Unlock()

		data, ok = t.pluginDataMap[uid]
		if !ok {
			err = fmt.Errorf("plugin: no plugin data exists with uid '%v'", uid)
			log.L.Error(err.Error())
			r = utils.ErrorBox(tr.S("bud.template.plugin.error"), err)
		}
		return
	}()
	if !ok {
		return
	}

	// Call the plugin render method.
	r, err = data.plugin.i.Render(c, data.data)
	if err != nil {
		err = fmt.Errorf("plugin: failed to render plugin of type '%v': %v", data.plugin.opts.Type, err)
		log.L.Error(err.Error())
		return utils.ErrorBox(tr.S("bud.template.plugin.error"), err)
	}

	return
}
