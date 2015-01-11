/*
 *  Bulldozer Framework
 *  Copyright (C) DesertBit
 */

package dialog

import (
	htmlTemplate "html/template"

	"code.desertbit.com/bulldozer/bulldozer/sessions"
	"code.desertbit.com/bulldozer/bulldozer/template"
	"code.desertbit.com/bulldozer/bulldozer/utils"
	"strconv"
)

type Size string

const (
	SizeSmall  Size = "small"
	SizeMedium Size = "medium"
	SizeLarge  Size = "large"
)

//###########################//
//### Event Receiver type ###//
//###########################//

type receiver struct{}

func (r *receiver) EventClosed(c *template.Context) {
	// Release the context.
	c.Release()
}

//###################//
//### Dialog type ###//
//###################//

type Dialog struct {
	t        *template.Template
	size     Size
	closable bool
	receiver receiver
}

// New creates a new template and passes the UID to the template.
func New(uid string) *Dialog {
	// Create a new dialog.
	d := &Dialog{
		t:        template.New(uid, "dialog"),
		size:     SizeMedium,
		closable: true,
	}

	// Register the internal dialog events.
	d.t.RegisterEvents(&d.receiver, "dialog")

	// Add the custom dialog functions.
	d.t.Funcs(template.FuncMap{
		"closeDialog": closeModalTemplateFunc,
	})

	return d
}

// Size sets the dialog size specified by a dialog.Size value.
// The defaut size is SizeMedium.
func (d *Dialog) SetSize(size Size) {
	d.size = size
}

// Whenever the modal is closable with a backdrop click or x button
func (d *Dialog) SetClosable(closable bool) {
	d.closable = closable
}

// RegisterEvents is the same as template.RegisterEvents...
func (d *Dialog) RegisterEvents(i interface{}, vars ...string) {
	d.t.RegisterEvents(i, vars...)
}

// Parse a template text
func (d *Dialog) Parse(text string) (err error) {
	// Append the dialog javascript code
	text += `{{js load}}
	var e=$("#{{$.Context.DomID}}__d");
	Kepler.modal.closed(e,function(){
		if (e.data('serverClosedDialog')!==true) {{emit dialog.Closed()}}
	});
{{end js}}`

	// Parse the template text.
	_, err = d.t.Parse(text)
	if err != nil {
		return err
	}

	return nil
}

// Create and show a new Dialog. The data interface is passed to the template execution call.
func (d *Dialog) Show(s *sessions.Session, data interface{}) (*template.Context, error) {
	// Create the optional options for the template.
	opts := template.ExecOpts{
		ID:    s.NewUniqueId(),
		DomID: s.NewUniqueDomID(),
	}

	// Execute the template
	o, c, err := d.t.ExecuteToString(s, data, opts)
	if err != nil {
		return nil, err
	}

	// Create the dialog DOM ID.
	dialogDomID := opts.DomID + "__d"

	// Transform to string
	closableStr := strconv.FormatBool(d.closable)

	// Create the command
	cmd := `Bulldozer.utils.addAndShowTmpModal('` + utils.EscapeJS(o) + `',{
			domId:'` + dialogDomID + `',
			closable:` + closableStr + `,
			class:'radius shadow ` + string(d.size) + `'
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

	// Call the event manually.
	d.receiver.EventClosed(c)
}

//###############//
//### Private ###//
//###############//

func closeModalTemplateFunc(c *template.Context) htmlTemplate.JS {
	return htmlTemplate.JS(`Kepler.modal.close("#` + c.DomID() + `__d");`)
}
