/*
 *  Bulldozer Framework
 *  Copyright (C) DesertBit
 */

package template

import (
	"code.desertbit.com/bulldozer/bulldozer/tr"
	"code.desertbit.com/bulldozer/bulldozer/utils"
	"fmt"
	"github.com/golang/glog"
	"strconv"
)

var (
	plugins map[string]Plugin = make(map[string]Plugin)
)

//####################//
//### Plugin types ###//
//####################//

type Plugin interface {
	Type() string                    // The plugin type.
	HasSection() bool                // If the plugin requires a template section.
	Prepare(s *PluginSettings) error // Prepare is called for each plugin context. Settings should be parsed...
	Render(c *Context, s *PluginSettings) (interface{}, error)
}

type PluginSettings struct {
	Value   interface{}
	Args    Args
	Section string
}

//##########################//
//### Plugin data struct ###//
//##########################//

type pluginDataMap map[int64]*pluginData

type pluginData struct {
	plugin   Plugin
	settings *PluginSettings
}

//##############//
//### Public ###//
//##############//

// RegisterPlugin adds and registers a template plugin with the given type.
// This method call is not thread-safe!
func RegisterPlugin(p Plugin) {
	// Get the type.
	typeStr := p.Type()

	// Print a warning if a previous plugin is overwritten.
	_, ok := plugins[typeStr]
	if ok {
		glog.Errorf("bulldozer template: plugin: overwritting already set plugin '%s'!", typeStr)
	}

	// Add the plugin interface to the map.
	plugins[typeStr] = p

	// Register the template parse function for the given plugin type.
	registerParseFunc(typeStr, parsePlugin)
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

	// Try to get the plugin interface.
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
		settings: &PluginSettings{
			Args: args,
		},
	}

	// Get the section if required.
	if plugin.HasSection() {
		data.settings.Section, err = getSection(typeStr, d.src, d)
		if err != nil {
			return fmt.Errorf("invalid plugin '%s' syntax! Missing end tag {{end %s}}", typeStr, typeStr)
		}
	}

	// Prepare the plugin settings
	err = plugin.Prepare(data.settings)
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
			glog.Errorf("render plugin panic: %v", e)
			r = utils.ErrorBox(tr.S("blz.template.plugin.error"), e)
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
			glog.Error(err)
			r = utils.ErrorBox(tr.S("blz.template.plugin.error"), err)
		}
		return
	}()
	if !ok {
		return
	}

	// Call the plugin render method.
	r, err = data.plugin.Render(c, data.settings)
	if err != nil {
		err = fmt.Errorf("plugin: failed to render plugin of type '%v': %v", data.plugin.Type(), err)
		glog.Error(err)
		return utils.ErrorBox(tr.S("blz.template.plugin.error"), err)
	}

	return
}
