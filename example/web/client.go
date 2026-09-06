//go:build wasm

package main

import (
	"syscall/js"

	"webtyp.com/fetch"
)

func main() {
	// Get the current window location origin to use as base URL
	origin := js.Global().Get("location").Get("origin").String()

	// Setup UI
	document := js.Global().Get("document")
	body := document.Get("body")

	// Title
	h1 := document.Call("createElement", "h1")
	h1.Set("innerHTML", "Fetch Text Example")
	body.Call("appendChild", h1)

	// Info
	p := document.Call("createElement", "p")
	p.Set("innerHTML", "Click the button to send text 'Alice' to the server.")
	body.Call("appendChild", p)

	// Button
	btn := document.Call("createElement", "button")
	btn.Set("innerText", "Send Text Data")
	body.Call("appendChild", btn)

	// Result container
	resultDiv := document.Call("createElement", "div")
	resultDiv.Set("style", "margin-top: 20px; padding: 10px; border: 1px solid #ccc;")
	body.Call("appendChild", resultDiv)

	// Click handler
	cb := js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		resultDiv.Set("innerText", "Sending...")

		textData := "Alice"

		// Send POST request with plain text
		fetch.Post(origin + "/api/user").
			ContentTypeText().
			Body([]byte(textData)).
			Timeout(5000).
			Send(func(resp *fetch.Response, err error) {
				if err != nil {
					resultDiv.Set("innerText", "Error: "+err.Error())
					return
				}
				resultDiv.Set("innerText", "Response: "+resp.Text())
			})

		return nil
	})
	btn.Call("addEventListener", "click", cb)

	// Keep the program running
	select {}
}
