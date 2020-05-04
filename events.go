package gouidom

import (
	"syscall/js"
)

type jsCallbackData struct {
	this js.Value
	args []js.Value
}

type domEvent struct {
	dispatch  js.Func
	eventChan chan jsCallbackData
}

func (de *domEvent) eventRouter() {
	//TODO: Figure out how to get from eventData to the js.Values callback function.
	for eventData := range de.eventChan {
		CLog("eventData is: %+v", eventData)
	}
}

func (v *VDOM) newVDOMEvents() {
	// Sets up the core dom event router
	v.events = &domEvent{
		eventChan: make(chan jsCallbackData),
		dispatch: js.FuncOf(func(this js.Value, args []js.Value) interface{} {
			return jsCallbackData{
				this: this,
				args: args,
			}
		}),
	}

	// All event router
	go v.events.eventRouter()

}
