/*
 *  Bulldozer Framework
 *  Copyright (C) DesertBit
 */

package plugin

import (
	"code.desertbit.com/bulldozer/bulldozer/template"
	"encoding/gob"
	"fmt"
)

const (
	contextStoreKeyData = "blzPluginData"
)

func init() {
	// Register the custom type.
	gob.Register(&storeData{})
}

//##########################//
//### Plugin data struct ###//
//##########################//

type storeData struct {
	ID                     string
	PluginType             string
	Section                string
	Args                   Args
	AdditionalStyleClasses []string
}

type Data struct {
	data  *storeData
	value interface{}
}

func newData(id string, pluginType string) *Data {
	return &Data{
		data: &storeData{
			ID:         id,
			PluginType: pluginType,
		},
	}
}

// Section returns the plugin template section.
func (d *Data) Section() interface{} {
	return d.data.Section
}

// Args returns the plugin template arguments.
func (d *Data) Args() Args {
	return d.data.Args
}

// Value returns the interface value which was returned by the Prepare method.
func (d *Data) Value() interface{} {
	return d.value
}

func (d *Data) prepareValue(vars ...Plugin) error {
	// Obtain the plugin interface.
	var plugin Plugin
	if len(vars) > 0 {
		plugin = vars[0]
	} else {
		// No plugin interface passed as optional argument.
		// Obtain it with the saved plugin type string.
		p := template.GetPlugin(d.data.PluginType)
		if p == nil {
			return fmt.Errorf("plugin: failed to prepare data value: plugin with type '%s' not found!", d.data.PluginType)
		}

		// Assert to plugin template type.
		tp, ok := p.(*templatePlugin)
		if !ok {
			return fmt.Errorf("plugin: failed to prepare data value: failed to assert plugin with type '%s'!", d.data.PluginType)
		}

		plugin = tp.plugin
	}

	// Call the plugin interface prepare method to
	// obtain the value.
	d.value = plugin.Prepare(d)

	return nil
}

//##############//
//### Public ###//
//##############//

// GetData obtains the plugin data from the context.
// The context has to be a plugin context.
func GetData(c *template.Context) (*Data, error) {
	// Get the plugin store data from the context store.
	i, ok := c.StoreGet(contextStoreKeyData)
	if !ok {
		return nil, fmt.Errorf("plugin: failed to obtain data value from context store: does not exists!")
	}

	// Assertion.
	d, ok := i.(*storeData)
	if !ok {
		return nil, fmt.Errorf("plugin: failed to assert data interface to store data type!")
	}

	// Create a new data value.
	data := &Data{
		data: d,
	}

	// Obtain the value by calling the prepare method.
	err := data.prepareValue()
	if err != nil {
		return nil, err
	}

	return data, nil
}

//###############//
//### Private ###//
//###############//

func setDataToContext(c *template.Context, data *Data) {
	// Save the plugin store data to the context store.
	c.StoreSet(contextStoreKeyData, data.data)
}
