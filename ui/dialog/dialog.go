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
	"io/ioutil"
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

// AddStyleClass adds one style classes.
func (d *Dialog) AddStyleClass(class string) *Dialog {
	d.t.AddStyleClass(class)
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

// RegisterEvents is the same as template.RegisterEvents...
func (d *Dialog) RegisterEvents(i interface{}, vars ...string) *Dialog {
	d.t.RegisterEvents(i, vars...)
	return d
}

// OnGetData is the same as template.OnGetData...
func (d *Dialog) OnGetData(f template.GetDataFunc) *Dialog {
	d.t.OnGetData(f)
	return d
}

// ParseFile parses a template file.
func (d *Dialog) ParseFile(filename string) (err error) {
	b, err := ioutil.ReadFile(filename)
	if err != nil {
		return err
	}

	return d.Parse(string(b))
}

// Parse a template text.
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

// Create and show a new Dialog.
// The data interface is passed to the template execution call if passed.
func (d *Dialog) Show(s *sessions.Session, data ...interface{}) (*template.Context, error) {
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
