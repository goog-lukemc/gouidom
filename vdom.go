package gouidom

import (
	"fmt"
	"path"
	"sort"
	"strings"
	"sync"
	"syscall/js"
)

//VDOM is a virtual DOM representation of the DOM hosting the WASM
type VDOM struct {
	mux    sync.Mutex
	vd     map[string]*Element
	events *domEvent
	window *js.Value
}

func (v *VDOM) GetCurrentPath() string {
	return js.Global().Get("window").Get("location").Get("pathname").String()
}

func (v *VDOM) elementExists(path string) error {

	if val, ok := v.vd[path]; ok {
		if val.jsValue != nil {
			// The element is in the vdom and also has a non nil jsValue in the actual dom
			return nil
		}
		return fmt.Errorf("errJSValueNotFound:%s", path)
	}
	return fmt.Errorf("errNotFound:%s", path)
}

func (v *VDOM) AddElement(eles ...*Element) {
	for _, ele := range eles {

		// Get the desired parent element from the VDOM
		p := v.vd[ele.Parent]

		// Lock  to do some work on the dom and vdom
		v.mux.Lock()

		// Logic that create and maintain the Id intergerty of the dom and vdom
		vPath, err := v.verifyElementId(ele)
		if err != nil {
			v.DumpVDOM()
			CLog("errIDVerification:%s", err.Error())
		}

		// Add the element to the vdom
		v.vd[vPath] = ele

		if !(ele.ID == "head") && !(ele.ID == "body") {
			// Append to the actual DOM
			if err := p.AppendChild(ele); err != nil {
				CLog("errAppendElement:%s", err.Error())
			}

		}

		// Unlock the vdom map
		v.mux.Unlock()
	}
}

func (v *VDOM) verifyElementId(ele *Element) (string, error) {

	// The parent property is required and exist in both DOMs
	if err := v.elementExists(ele.Parent); err != nil {
		return "", err
	}

	if ele.ID == "" {
		// Create a unique id. Hash the parent path + number of current elements in the parent
		// get the current number of children in the parentNode

		cn := v.vd[ele.Parent].nodeCount()

		ele.ID = newElementIDFromDOMPath(fmt.Sprintf("%s-%d", ele.Parent, cn))
		//CLog("node:%v nodeCount:%v ID:%s", ele.Parent, cn, ele.ID)
	}

	return path.Join(ele.Parent, ele.ID), nil
}

func (v *VDOM) DumpVDOM() {
	for k, v := range v.vd {
		CLog("Key:%s Parent:%s ID:%s DOM:%v", k, v.Parent, v.ID, v.jsValue)
	}
}

func (v *VDOM) GenStyleTemplate() {
	sb := strings.Builder{}
	keys := []string{}
	classes := map[string]string{}
	htmltags := map[string]interface{}{}

	for k, v := range v.vd {
		htmltags[v.Typ] = nil
		l := v.jsValue.Get("classList")
		if l.Type() != js.TypeUndefined {
			cnt := l.Get("length").Int()
			if cnt > 0 {
				csn := []string{}
				for i := 0; i < cnt; i++ {
					csn = append(csn, l.Call("item", i).String())
				}
				classes[k] = strings.Join(csn, ",")
			}
		}
		keys = append(keys, k)
	}

	cm := map[string]string{}
	for k, v := range classes {
		for _, rty := range strings.Split(v, ",") {
			if _, ok := cm[rty]; ok {
				cm[rty] = cm[rty] + fmt.Sprintf("/*%s*/\n", k)
				continue
			}
			cm[rty] = fmt.Sprintf("/*%s*/\n", cm[rty])
		}

	}

	for k := range htmltags {
		if k == "<undefined>" {
			continue
		}
		sb.WriteString(TagCSSBlock(k))
	}

	for c, h := range cm {
		sb.WriteString(ClassCSSBlock(c, h))
	}

	sort.Strings(keys)
	for _, k := range keys {
		sb.WriteString(IdCSSBlock(v.vd[k].ID, k))
	}

	CLog("%s", sb.String())
	CLog("%s", v.vd["html"].jsValue.Get("outerHTML").String())
}

func TagCSSBlock(tagName string) string {
	return fmt.Sprintf(`#%s{

}

`, tagName)
}

func ClassCSSBlock(name string, path string) string {
	return fmt.Sprintf(`%s.%s{

}

`, path, name)
}

func IdCSSBlock(id, path string) string {
	return fmt.Sprintf(`/* %s */
[id='%s']{

}

`, path, id)
}

// NewApp Creates a new grid on a page
func NewApp(appTitle string) (*VDOM, error) {

	// Set the title of the appear
	if err := SetAppTitle(appTitle); err != nil {
		return nil, err
	}

	// Creating a virtual dom
	v := VDOM{
		vd: make(map[string]*Element),
	}

	// Manually adding the app root html element
	gd := js.Global().Get("document")
	html := &Element{
		ID:      "html",
		Typ:     "html",
		jsValue: &gd,
	}
	v.vd["html"] = html

	// Spin up the event manager for domEvent which starts a go routine to monitor events.
	v.newVDOMEvents()

	// Create a callback function for a navigatoin hashchange
	html.Fulfillment = func(this js.Value, args []js.Value) interface{} {
		CLog("%+v", args[0])
		return nil
	}

	head, err := GetElementByID(HTMLTag.Head)
	if err != nil {
		return nil, err
	}
	head.Parent = "html"

	body, err := GetElementByID(HTMLTag.Body)
	if err != nil {
		return nil, err
	}
	body.Parent = "html"

	v.AddElement(head, body)

	w := js.Global().Get("window")
	v.window = &w

	return &v, nil
}

// SetAppTitle set the title of the app on page. In most browsers this will appear at the tab.
func SetAppTitle(appTitle string) error {
	// Set the title of the append
	title, err := GetElementByID("title")
	if err != nil {
		return err
	}
	// Set the title of the all
	title.SetInnerHTML(appTitle)
	return nil

}
