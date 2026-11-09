//go:build wasm

package fetch

import (
	"syscall/js"

	. "webtyp.com/fmt"
)

// doRequest is the WASM implementation for making an HTTP request using the browser's fetch API.
func doRequest(r *Request, callback func(*Response, error)) {
	// 1. Build the full URL.
	fullURL, err := buildURL(r)
	if err != nil {
		callback(nil, err)
		return
	}

	// 2. Prepare request body.
	var jsBody js.Value
	if len(r.body) > 0 {
		// Convert Go byte slice to a JS Uint8Array's buffer.
		uint8Array := js.Global().Get("Uint8Array").New(len(r.body))
		js.CopyBytesToJS(uint8Array, r.body)
		jsBody = uint8Array.Get("buffer")
	}

	// 3. Prepare headers object for the fetch call.
	jsHeaders := js.Global().Get("Headers").New()
	for _, h := range r.headers {
		jsHeaders.Call("append", h.Key, h.Value)
	}

	// 4. Prepare the main options object for fetch.
	options := js.Global().Get("Object").New()
	options.Set("method", r.method)
	options.Set("headers", jsHeaders)
	if !jsBody.IsUndefined() && r.method != "GET" && r.method != "HEAD" {
		options.Set("body", jsBody)
	}

	// 5. Handle timeout with AbortController.
	if r.timeout > 0 {
		controller := js.Global().Get("AbortController").New()
		options.Set("signal", controller.Get("signal"))
		js.Global().Call("setTimeout", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
			controller.Call("abort")
			return nil
		}), r.timeout)
	}

	// 6. Define promise handlers to bridge async JS to sync Go.
	var success, failure, responseHandler js.Func

	// cleanup releases the JS functions when the request is complete.
	cleanup := func() {
		success.Release()
		failure.Release()
		responseHandler.Release()
	}

	// success handles the final response body (as an ArrayBuffer).
	success = js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		// Convert the JS ArrayBuffer back to a Go []byte.
		uint8Array := js.Global().Get("Uint8Array").New(args[0])
		goBytes := make([]byte, uint8Array.Get("length").Int())
		js.CopyBytesToGo(goBytes, uint8Array)

		// The response object was partially built in responseHandler, we need to pass it here?
		// No, the promise chain is tricky with how we want to return both headers/status AND body.
		// So we will change how we handle this.
		// Ideally, we want to capture the response object in the first promise,
		// and then combine it with the body in the second promise.

		// However, the callback structure expects *Response.
		// We'll use a closure variable to hold the partial response.
		return nil
	})

	// Re-implementing logic to capture response details properly.

	// failure handles any error in the promise chain.
	failure = js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		var errMsg string
		if len(args) > 0 && !args[0].IsUndefined() && !args[0].IsNull() {
			jsErr := args[0]
			// Try to get error message
			if jsErr.Type() == js.TypeString {
				errMsg = jsErr.String()
			} else if jsErr.Get("message").Type() == js.TypeString {
				errMsg = jsErr.Get("message").String()
			} else {
				errMsg = jsErr.Call("toString").String()
			}
		}
		if errMsg == "" {
			errMsg = "unknown network error (possibly CORS, network unavailable, or invalid URL)"
		}
		err := Errf("fetch failed: %s (URL: %s)", errMsg, fullURL)

		callback(nil, err)
		cleanup()
		return nil
	})

	var partialResponse *Response

	// responseHandler handles the initial Response object from fetch.
	responseHandler = js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		jsResp := args[0]

		status := jsResp.Get("status").Int()

		// Extract headers
		var headers []Header
		// Headers iterator
		jsHeaders := jsResp.Get("headers")
		iterator := jsHeaders.Call("entries")

		for {
			entry := iterator.Call("next")
			if entry.Get("done").Bool() {
				break
			}
			pair := entry.Get("value")
			headers = append(headers, Header{
				Key:   pair.Index(0).String(),
				Value: pair.Index(1).String(),
			})
		}

		partialResponse = &Response{
			Status:     status,
			Headers:    headers,
			RequestURL: fullURL,
			Method:     r.method,
		}

		// Always read the body as ArrayBuffer, regardless of status.
		// The user is responsible for checking status code.
		return jsResp.Call("arrayBuffer")
	})

	// successBody handles the ArrayBuffer from the response body.
	successBody := js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		uint8Array := js.Global().Get("Uint8Array").New(args[0])
		goBytes := make([]byte, uint8Array.Get("length").Int())
		js.CopyBytesToGo(goBytes, uint8Array)

		if partialResponse != nil {
			partialResponse.body = goBytes
			callback(partialResponse, nil)
		} else {
			// Should not happen
			callback(nil, Errf("internal error: response missing"))
		}

		cleanup()
		return nil
	})

	// Re-assign success to successBody for clarity in cleanup if I used the previous name
	// But I defined successBody separately. So I need to update cleanup.

	cleanup = func() {
		failure.Release()
		responseHandler.Release()
		successBody.Release()
	}

	js.Global().Call("fetch", fullURL, options).
		Call("then", responseHandler).
		Call("then", successBody).
		Call("catch", failure)
}

func getOrigin() string {
	return js.Global().Get("location").Get("origin").String()
}
