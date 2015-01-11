/*
 *  Bulldozer Framework
 *  Copyright (C) DesertBit
 */

package messagebox

import (
	"code.desertbit.com/bulldozer/bulldozer/log"
	"code.desertbit.com/bulldozer/bulldozer/sessions"
	"code.desertbit.com/bulldozer/bulldozer/template"
	"code.desertbit.com/bulldozer/bulldozer/tr"
	"code.desertbit.com/bulldozer/bulldozer/ui/dialog"
	"fmt"
	"strconv"
	"strings"
	"sync"
)

const (
	messageBoxTemplateUID = "blz_msgbox"
	cacheValueKey         = "bzrMsgBox"
)

const (
	ButtonOk     Button = 1 << iota
	ButtonYes    Button = 1 << iota
	ButtonNo     Button = 1 << iota
	ButtonCancel Button = 1 << iota
)

const (
	TypeDefault  MessageBoxType = 1 << iota
	TypeSuccess  MessageBoxType = 1 << iota
	TypeWarning  MessageBoxType = 1 << iota
	TypeAlert    MessageBoxType = 1 << iota
	TypeInfo     MessageBoxType = 1 << iota
	TypeQuestion MessageBoxType = 1 << iota
)

type Button int
type MessageBoxType int
type Callback func(button Button)

type MessageBox struct {
	title          string
	text           string
	buttons        Button
	messageBoxType MessageBoxType
	icon           string // CSS icon class. Predefined Kepler classes or font awesome classes...
	callback       Callback
}

type templButton struct {
	Id   string
	Text string
	Type string
}

var (
	d *dialog.Dialog
)

func init() {
	// Create the dialog
	d = dialog.New(messageBoxTemplateUID)

	// The messagebox should not be closable
	d.SetClosable(false)

	// Set the dialog template text and parse it
	err := d.Parse(messageBoxText)
	if err != nil {
		log.L.Fatalf("failed to parse message box dialog template: %v", err)
	}

	// Set the message box size
	d.SetSize(dialog.SizeSmall)

	// Register the internal dialog events.
	d.RegisterEvents(&receiver{})
}

//################################//
//### Private cache value type ###//
//################################//

type callbacks struct {
	mutex sync.Mutex
	set   map[string]Callback
}

func newCallbacks() *callbacks {
	return &callbacks{
		set: make(map[string]Callback),
	}
}

//###########################//
//### Event Receiver type ###//
//###########################//

type receiver struct{}

func (r *receiver) EventButtonClicked(c *template.Context, b int) {
	// Get the callbacks from the session cache.
	i, _ := c.Session().CacheGet(cacheValueKey, func() interface{} {
		return newCallbacks()
	})

	// Assertion.
	callbacks := i.(*callbacks)

	// The access key is the context ID.
	id := c.ID()

	// Lock the mutex.
	callbacks.mutex.Lock()
	defer callbacks.mutex.Unlock()

	// Try to find the callback for the current context ID.
	cb, ok := callbacks.set[id]
	if !ok {
		log.L.Error("messagebox: failed to get messagebox callback for id: '%s'", id)
		return
	}

	// Remove the callback value from the cache.
	delete(callbacks.set, id)

	if cb == nil {
		log.L.Error("messagebox: failed to get messagebox callback for id: '%s': callback is nil!", id)
		return
	}

	// Call the callback after the mutex got unlocked.
	defer cb(Button(b))
}

//#######################//
//### MessageBox type ###//
//#######################//

// New creates a new MessageBox
func New() *MessageBox {
	return &MessageBox{
		buttons:        ButtonOk,
		messageBoxType: TypeDefault,
	}
}

// SetTitle sets the messagebox title
func (m *MessageBox) SetTitle(title string) *MessageBox {
	m.title = title
	return m
}

// SetText sets the messagebix text
func (m *MessageBox) SetText(text string) *MessageBox {
	m.text = text
	return m
}

// SetType sets the messagebox type
func (m *MessageBox) SetType(t MessageBoxType) *MessageBox {
	m.messageBoxType = t
	return m
}

// SetButtons sets the messagebox buttons
func (m *MessageBox) SetButtons(buttons Button) *MessageBox {
	m.buttons = buttons
	return m
}

// SetIcon sets the CSS icon class. Predefined Kepler classes or font awesome classes...
func (m *MessageBox) SetIcon(iconClass string) *MessageBox {
	m.icon = " " + strings.TrimSpace(iconClass)
	return m
}

// SetCallback sets the callback which is called as soon as any messagebox button is clicked.
// Note: This callback is saved in the session cache and it won't survive application restarts!
func (m *MessageBox) SetCallback(c Callback) *MessageBox {
	m.callback = c
	return m
}

// Show shows the messagebox. Errors are always logged.
func (m *MessageBox) Show(s *sessions.Session) (err error) {
	// Prepare the buttons
	var templButtons []templButton

	buttonCount := 0
	if m.buttons&ButtonOk == ButtonOk {
		templButtons = append(templButtons, templButton{
			"ok",
			tr.S("blz.messagebox.buttonOk"),
			strconv.Itoa(int(ButtonOk)),
		})
		buttonCount++
	}
	if m.buttons&ButtonYes == ButtonYes {
		templButtons = append(templButtons, templButton{
			"yes",
			tr.S("blz.messagebox.buttonYes"),
			strconv.Itoa(int(ButtonYes)),
		})
		buttonCount++
	}
	if m.buttons&ButtonNo == ButtonNo {
		templButtons = append(templButtons, templButton{
			"no",
			tr.S("blz.messagebox.buttonNo"),
			strconv.Itoa(int(ButtonNo)),
		})
		buttonCount++
	}
	if m.buttons&ButtonCancel == ButtonCancel {
		templButtons = append(templButtons, templButton{
			"cancel",
			tr.S("blz.messagebox.buttonCancel"),
			strconv.Itoa(int(ButtonCancel)),
		})
		buttonCount++
	}

	// Check if no buttons are passed
	if buttonCount == 0 {
		err = fmt.Errorf("failed to show message box: no buttons!")
		log.L.Error(err.Error())
		return
	}

	// Get the type class
	typeClass := ""
	if m.messageBoxType == TypeInfo {
		typeClass = " info"
	} else if m.messageBoxType == TypeSuccess {
		typeClass = " success"
	} else if m.messageBoxType == TypeWarning {
		typeClass = " warning"
	} else if m.messageBoxType == TypeAlert {
		typeClass = " alert"
	} else if m.messageBoxType == TypeQuestion {
		typeClass = " question"
	}

	// Create the template data
	data := struct {
		Title        string
		Text         string
		Buttons      []templButton
		ButtonColumn int
		IconClass    string
		TypeClass    string
	}{
		Title:        m.title,
		Text:         m.text,
		Buttons:      templButtons,
		ButtonColumn: 12 / buttonCount,
		IconClass:    m.icon,
		TypeClass:    typeClass,
	}

	// Show the message box
	c, err := d.Show(s, data)
	if err != nil {
		err = fmt.Errorf("failed to show message box: %v", err)
		log.L.Error(err.Error())
		return
	}

	// Add the callback if present.
	if m.callback != nil {
		// Get the callbacks from the session cache.
		i, _ := s.CacheGet(cacheValueKey, func() interface{} {
			return newCallbacks()
		})

		// Assertion.
		callbacks := i.(*callbacks)

		// Lock the mutex.
		callbacks.mutex.Lock()
		defer callbacks.mutex.Unlock()

		// Add the callback to the map.
		callbacks.set[c.ID()] = m.callback
	}

	return nil
}

const messageBoxText = `<div class="topbar{{#.TypeClass}}">
    <div class="icon{{#.IconClass}}"></div>
    <div class="title">
        <h3>{{#.Title}}</h3>
    </div>
</div>
<div class="kepler grid">
	<div class="large-12 column"><p>{{#.Text}}</p></div>
	<div class="large-12 column"><hr></hr></div>
	{{range $b := #.Buttons}}
		<div class="medium-{{#.ButtonColumn}} column">
			<a id="{{id $b.Id}}" class="kepler button expand">{{$b.Text}}</a>
		</div>
		{{js load}}
			$("#{{id $b.Id}}").click(function() {
				var t = "{{$b.Type}}";
				{{emit ButtonClicked(t)}}
				{{closeDialog $.Context}}
			});
		{{end js}}
	{{end}}
</div>`
