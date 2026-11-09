# webtyp/fetch
<img src="docs/img/badges.svg">

A minimal, zero-dependency, WASM-compatible HTTP client for Go.

## Features

- **Tiny API**: `Get`, `Post`, `Put`, `Patch`, `Delete`
- **Declarative Headers**: `ContentTypeJSON()`, `ContentTypeBinary()`, etc.
- **WASM Compatible**: Works in browsers using `fetch` API and in standard Go
- **Async Support**: `Send` (callback) and `Dispatch` (fire-and-forget)

## Installation

```bash
go get webtyp.com/fetch
```

## Usage

```go
package main

import (
	"webtyp.com/fetch"
)

func main() {
	// Simple GET request using relative path (auto-detects origin in WASM)
	fetch.Get("/data").
		Send(func(resp *fetch.Response, err error) {
			if err != nil {
				println("Error:", err.Error())
				return
			}
			println("Status:", resp.Status)
			println("Body:", resp.Text())
		})

	// POST request with JSON body using custom BaseURL
	body := []byte(`{"name":"Alice"}`)
	fetch.Post("/users").
		BaseURL("https://api.example.com").
		ContentTypeJSON().
		Body(body).
		Timeout(5000).
		Send(func(resp *fetch.Response, err error) {
			if err != nil {
				return
			}
			println("Created:", resp.Text())
		})
}
```

## Documentation

- [Base URL & Endpoint Resolution](docs/BASE_URL.md) - How URLs are resolved automatically
- [HTTP Headers](docs/HEADERS.md) - How to set and retrieve headers
- [CORS Troubleshooting](docs/CORS.md) - Common issues in WASM environments

## Content-Type Helpers

The following declarative methods are available on `Request`:
- `ContentTypeJSON()` -> `application/json`
- `ContentTypeBinary()` -> `application/octet-stream`
- `ContentTypeForm()` -> `application/x-www-form-urlencoded`
- `ContentTypeText()` -> `text/plain`
- `ContentTypeHTML()` -> `text/html`

## Global Handler (Dispatch)

For fire-and-forget requests or centralized error handling:

```go
func main() {
	// Set a global handler
	fetch.SetHandler(func(resp *fetch.Response) {
		if resp.Status >= 400 {
			println("Request to", resp.RequestURL, "failed with", resp.Status)
		}
	})

	// Fire and forget
	fetch.Post("/analytics").
		ContentTypeText().
		Body([]byte("event=click")).
		Dispatch()
}
```

