package gouidom

import (
	"fmt"
	"hash/fnv"
	"strings"
	"syscall/js"
)

// Element
type Element struct {
	// ID is the id of the HTML element. In this implementation it is required that every
	// element have an ID and they IDs are unique within the page.
	ID string

	// typ is usedful to set at create type of the elements
	Typ string

	// Text is the initial inner text of the element. This property is only used to setAttribute
	// the initial value
	Text string

	//jsValue is the underlying dom item
	jsValue *js.Value

	// Parent is the parent path intentded in the vdom. It will them be appended to the actual DOM
	// as a child of the parent element.
	Parent string

	// Fulfillment is a function that fulFills the event.
	Fulfillment jsFulFill
}
type jsFulFill func(js.Value, []js.Value) interface{}

func (ele *Element) AddClass(name string) {
	ele.jsValue.Get("classList").Call(JSMethod.AddClass, name)
}

func (ele *Element) nodeCount() int {
	return ele.jsValue.Get("childElementCount").Int()
}

func (ele *Element) getVDOMPath() string {
	var sa []string
	val := *ele.jsValue

	for val.Type() != js.TypeUndefined {
		if val.Type() == js.TypeNull {
			break
		}
		if val.Get("id").Type() == js.TypeUndefined {
			break
		}
		sa = append(sa, val.Get("id").String())
		val = val.Get("parentNode")

	}
	// Reverse the slice
	for i := len(sa)/2 - 1; i >= 0; i-- {
		opp := len(sa) - 1 - i
		sa[i], sa[opp] = sa[opp], sa[i]
	}
	return strings.Join(sa, "/")

}

func (ele *Element) SetInnerHTML(c string) {
	ele.jsValue.Set("innerHTML", c)
}

// NewElement an element that can be append to the vdom. Using "" as the ID will auto generate a unique string
// for the id value.
func NewElement(id string, parent string, typ string, initialText string, class ...string) (*Element, error) {
	// Create the js.Value to be published to the DOM
	jv, err := jsMethodCall(nil, JSMethod.CreateElement, typ)
	if err != nil {
		return nil, err
	}
	if initialText != "" {
		jv.Set("innerHTML", initialText)
	}

	ele := &Element{
		ID:      id,
		Typ:     typ,
		Text:    initialText,
		Parent:  parent,
		jsValue: jv,
	}
	for _, c := range class {
		ele.AddClass(c)
	}
	return ele, nil

}

// AppendChild appends the element to the dom.
func (ele *Element) AppendChild(child *Element) error {

	if _, err := jsMethodCall(ele.jsValue, JSMethod.AppendChild, child.jsValue); err != nil {
		return err
	}

	if err := child.SetAttribute("id", child.ID); err != nil {
		return err
	}

	if ele.Text != "" {
		if err := child.SetAttribute("text", child.Text); err != nil {
			return err
		}
	}

	return nil
}

func (ele *Element) ScrollIntoView(alignTo bool) error {
	if _, err := jsMethodCall(ele.jsValue, JSMethod.ScrollIntoView, alignTo); err != nil {
		return err
	}
	return nil
}

// SetAttribute set a attribute in the actual dom. There is no change to the vdom
func (ele *Element) SetAttribute(name string, value string) error {
	if ele.jsValue == nil {
		return fmt.Errorf("noExistingJSValue: existing js value requried to call this method")
	}
	_, err := jsMethodCall(ele.jsValue, JSMethod.SetAttribute, name, value)
	return err
}

// AddEventListener adds an event listener to a dom element
func (ele *Element) AddEventListener(eventName string, cbs ...js.Func) error {
	for _, cb := range cbs {
		_, err := jsMethodCall(ele.jsValue, JSMethod.AddEventListener, eventName, cb)
		if err != nil {
			return err
		}
	}
	return nil
}

// GetElementByID get and element from the existing DOM and return a helper element
func GetElementByID(id string) (*Element, error) {
	val, err := jsMethodCall(nil, JSMethod.GetElementByID, id)
	if err != nil {
		return nil, err
	}
	return &Element{
		ID:      id,
		Typ:     val.Get("type").String(),
		jsValue: val,
	}, nil
}

func jsMethodCall(jsVal *js.Value, method string, params ...interface{}) (*js.Value, error) {

	if jsVal == nil {
		g := js.Global().Get("document")
		jsVal = &g
	}
	val := jsVal.Call(method, params...)
	if val.Type() == js.TypeNull {
		return nil, fmt.Errorf("unexpectedNull: a null js value is not expected")
	}
	return &val, nil
}

// newElementIdFromDOMPath dsa
func newElementIDFromDOMPath(s string) string {
	hash := fnv.New32a()
	hash.Write([]byte(s))
	return fmt.Sprint(hash.Sum32())
}

// HTMLTagNames is create to contol which HTML tag can be used in the solution. The tags below are tested.
type HTMLTagNames struct {
	Body    string
	Div     string
	Script  string
	Header  string
	Footer  string
	Input   string
	Img     string
	Head    string
	Span    string
	Style   string
	Button  string
	Section string
	Article string
	Pre     string
	Code    string
}

// JSMethodNames contrains the correct string property name for tested javascript methods
type JSMethodNames struct {
	CreateElement    string
	AppendChild      string
	GetElementByID   string
	SetAttribute     string
	AddEventListener string
	Toggle           string
	Contains         string
	AddClass         string
	ScrollIntoView   string
}

// JSEventNames represent the event name tested with the js event handler in this package.
type JSEventNames struct {
	Click   string
	OnInput string
	Keyup   string
}

var (
	// HTMLTag is export from this package as a global variable and represents a property for each support HTMLTagName
	HTMLTag HTMLTagNames

	// JSMethod is export from this package as a global variable and represents a property for each support Javescript Method
	JSMethod JSMethodNames

	// JSEvent is export from this package as a global variable and represents a property for each support Javescript Event
	JSEvent JSEventNames
)

func init() {
	HTMLTag = HTMLTagNames{
		Body:    "body",
		Div:     "div",
		Script:  "string",
		Header:  "header",
		Footer:  "footer",
		Input:   "input",
		Img:     "img",
		Head:    "head",
		Span:    "span",
		Style:   "style",
		Button:  "button",
		Section: "section",
		Article: "article",
		Pre:     "pre",
		Code:    "code",
	}

	JSMethod = JSMethodNames{
		CreateElement:    "createElement",
		AppendChild:      "appendChild",
		GetElementByID:   "getElementById",
		SetAttribute:     "setAttribute",
		AddEventListener: "addEventListener",
		Toggle:           "toggle",
		Contains:         "contains",
		AddClass:         "add",
		ScrollIntoView:   "scrollIntoView",
	}

	JSEvent = JSEventNames{
		Click:   "click",
		OnInput: "onInput",
		Keyup:   "keyUp",
	}

}
