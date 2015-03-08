/*
 *  Bulldozer Framework
 *  Copyright (C) DesertBit
 */

package dialog

import (
	"code.desertbit.com/bulldozer/bulldozer/sessions"
	"code.desertbit.com/bulldozer/bulldozer/template"
	"code.desertbit.com/bulldozer/bulldozer/utils"

	"fmt"
	"strconv"
)

type Size string

const (
	SizeSmall  Size = "small"
	SizeMedium Size = "medium"
	SizeLarge  Size = "large"
)

//###################//
//### Dialog type ###//
//###################//

type Dialog struct {
	t            *template.Template
	size         Size
	closable     bool
	StyleClasses []string
}

// New creates a new dialog.
func New() *Dialog {
	// Create a new dialog.
	d := &Dialog{
		size:     SizeMedium,
		closable: true,
	}

	return d
}

// AddStyleClasses adds the style classes to the dialog modal.
func (d *Dialog) AddStyleClasses(classes ...string) *Dialog {
	d.StyleClasses = append(d.StyleClasses, classes...)
	return d
}

// Size sets the dialog size specified by a dialog.Size value.
// The defaut size is SizeMedium.
func (d *Dialog) SetSize(size Size) *Dialog {
	d.size = size
	return d
}

// Whenever the modal is closable with a backdrop click or x button
func (d *Dialog) SetClosable(closable bool) *Dialog {
	d.closable = closable
	return d
}

// SetTemplate sets the dialog body template.
func (d *Dialog) SetTemplate(t *template.Template) *Dialog {
	d.t = t

	return d
}

// Create and show a new Dialog.
// The data interface is passed to the template execution call if passed.
func (d *Dialog) Show(s *sessions.Session, data ...interface{}) (*template.Context, error) {
	if d.t == nil {
		return nil, fmt.Errorf("failed to show dialog: template is nil!")
	}

	// Create the optional options for the template.
	opts := template.ExecOpts{
		ID:    s.NewUniqueId(),
		DomID: s.NewUniqueDomID(),
	}

	// Set the data if present.
	if len(data) > 0 {
		opts.Data = data[0]
	}

	// Execute the template
	o, c, err := d.t.ExecuteToString(s, opts)
	if err != nil {
		return nil, err
	}

	// Create the dialog DOM ID.
	dialogDomID := opts.DomID + "__d"

	// Transform to string
	closableStr := strconv.FormatBool(d.closable)

	// Transform the additional style classes to a string.
	var styles string
	for _, style := range d.StyleClasses {
		styles += " " + style
	}

	// Create the command
	cmd := `Bulldozer.utils.addAndShowTmpModal('` + utils.EscapeJS(o) + `',{
			domId:'` + dialogDomID + `',
			closable:` + closableStr + `,
			class:'radius shadow ` + string(d.size) + styles + `'
		});`

	// Execute the command on the client side.
	// The loading indicator is hidden automatically by the Bulldozer.core.execJsLoad() function.
	s.SendCommand(cmd)

	return c, nil
}

// Close the dialog.
func (d *Dialog) Close(c *template.Context) {
	// Close the dialog
	c.Session().SendCommand(`(function(){
		var e=$('#` + c.DomID() + `__d');
		e.data('serverClosedDialog', true);
		Kepler.modal.close(e);
	})();`)

	// Release the context.
	c.Release()
}
