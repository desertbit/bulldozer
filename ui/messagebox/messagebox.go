/*
 *  Bulldozer Framework
 *  Copyright (C) DesertBit
 */

package messagebox

import (
	tr "code.desertbit.com/bulldozer/bulldozer/translate"

	"code.desertbit.com/bulldozer/bulldozer/callback"
	"code.desertbit.com/bulldozer/bulldozer/log"
	"code.desertbit.com/bulldozer/bulldozer/sessions"
	"code.desertbit.com/bulldozer/bulldozer/template"
	"code.desertbit.com/bulldozer/bulldozer/ui/dialog"
	"fmt"
	"strconv"
	"strings"
)

const (
	messageBoxTemplateUID = "budMsgbox"
	sessionValueKeyPrefix = "budMsgBox_"
	callbackPrefixName    = "budMsgBox_"
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
type Callback func(s *sessions.Session, button Button)

type MessageBox struct {
	title          string
	text           string
	buttons        Button
	messageBoxType MessageBoxType
	icon           string // CSS icon class. Predefined Kepler classes or font awesome classes...
	callbackFunc   Callback
	callbackName   string
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
	// Create the dialog and set the default values.
	d = dialog.New().
		SetSize(dialog.SizeSmall).
		SetClosable(false)

	// Parse the messagebox template.
	t, err := template.New(messageBoxTemplateUID, "msgbox").Parse(messageBoxText)
	if err != nil {
		log.L.Fatalf("failed to parse message box dialog template: %v", err)
	}

	// Register the events.
	t.RegisterEvents(new(receiver))

	// Set the template.
	d.SetTemplate(t)
}

//###########################//
//### Event Receiver type ###//
//###########################//

type receiver struct{}

func (r *receiver) EventButtonClicked(c *template.Context, b int) {
	// Save the session pointer.
	s := c.Session()

	// Close the messagebox and hide the loading indicator.
	s.HideLoadingIndicator()
	d.Close(c)

	// Create the session value access key.
	key := sessionValueKeyPrefix + c.ID()

	// Get the callbacks from the session cache.
	i, ok := s.InstancePull(key)
	if !ok {
		i, ok = s.CachePull(key)
		if !ok {
			log.L.Warning("messagebox: failed to get messagebox callback for id: '%s': this is caused, because messagebox callbacks set with messagebox.CallbackFunc are stored in the session cache and don't survive application restarts! Use messagebox.SetCallback instead...", c.ID())
			return
		}
	}

	// Assertion
	switch i.(type) {
	case string:
		// Assert and call the callback.
		name := i.(string)
		callback.Call(name, s, Button(b))
	case Callback:
		// Assert and call the callback.
		cb := i.(Callback)
		if cb != nil {
			cb(s, Button(b))
		}
	default:
		log.L.Error("messagebox: failed to get messagebox callback for id: '%s': unkown callback type!", c.ID())
		return
	}
}

//##############//
//### Public ###//
//##############//

// New creates a new MessageBox
func New() *MessageBox {
	return &MessageBox{
		buttons:        ButtonOk,
		messageBoxType: TypeDefault,
	}
}

// RegisterCallback registers a callback. This is necessary, because
// otherwise callbacks could not be called after application restarts.
// They have to be registered globally...
// One optional boolean can be passed, to force a overwrite of
// a previous registered callback with the same name.
func RegisterCallback(name string, cb Callback, vars ...bool) {
	// Register the callback.
	callback.Register(callbackPrefixName+name, cb, vars...)
}

//#######################//
//### MessageBox type ###//
//#######################//

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
// Use RegisterCallback to register a callback with a name.
func (m *MessageBox) SetCallback(callbackName string) *MessageBox {
	m.callbackName = callbackPrefixName + callbackName
	return m
}

// SetCallbackFunc sets the callback which is called as soon as any messagebox button is clicked.
// Note: This callback is saved in the session cache and it won't survive application restarts!
// Use SetCallback instead!
func (m *MessageBox) SetCallbackFunc(c Callback) *MessageBox {
	m.callbackFunc = c
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
			tr.S("bud.messagebox.buttonOk"),
			strconv.Itoa(int(ButtonOk)),
		})
		buttonCount++
	}
	if m.buttons&ButtonYes == ButtonYes {
		templButtons = append(templButtons, templButton{
			"yes",
			tr.S("bud.messagebox.buttonYes"),
			strconv.Itoa(int(ButtonYes)),
		})
		buttonCount++
	}
	if m.buttons&ButtonNo == ButtonNo {
		templButtons = append(templButtons, templButton{
			"no",
			tr.S("bud.messagebox.buttonNo"),
			strconv.Itoa(int(ButtonNo)),
		})
		buttonCount++
	}
	if m.buttons&ButtonCancel == ButtonCancel {
		templButtons = append(templButtons, templButton{
			"cancel",
			tr.S("bud.messagebox.buttonCancel"),
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

	// Save the callback to the session...
	key := sessionValueKeyPrefix + c.ID()
	if len(m.callbackName) > 0 {
		// Set the callback name to the session instance values.
		s.InstanceSet(key, m.callbackName)
	} else {
		// Hint: This won't survive application restarts.
		// Set the callbacks to the session cache.
		s.CacheSet(key, m.callbackFunc)
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
				Bulldozer.loadingIndicator.show();
				{{emit ButtonClicked(t)}}
			});
		{{end js}}
	{{end}}
</div>`
