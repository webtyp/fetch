# Architecture Refactor: Automatic Codec System

## 1. Philosophy: Simplicity and Automation

The new architecture is based on a simple principle: **explicit encoding methods prevent confusion and ensure correct codec usage.**

- **Explicit Methods**: `SendJSON()` and `SendBinary()` make it clear which codec is used.
- **No Guessing**: Users explicitly choose JSON (for compatibility) or TinyBin (for performance).
- **Platform-Aware**: The library automatically uses the best implementation for the target platform (stdlib vs. WASM).
- **Domain Agnostic**: A single client can make requests to multiple domains (essential for microservices, third-party APIs, and CRUDP integration).

## 2. Core Components

### `Fetch` Struct: The Codec Manager

The `Fetch` struct is the central point of the library. It is created once and manages all shared resources, including the TinyBin instance and CORS configuration.

```go
// fetchgo.go
package fetchgo

import "webtyp.com/tinybin"

// Fetch manages HTTP clients with explicit codec methods.
type Fetch struct {
    tb              *tinybin.TinyBin
    corsMode        string
    corsCredentials bool
}

// New creates a new Fetch instance with sensible defaults.
func New() *Fetch {
    return &Fetch{
        tb:              tinybin.New(),
        corsMode:        "cors",
        corsCredentials: false,
    }
}

// SetCORS configures CORS behavior for WASM/browser requests.
func (f *Fetch) SetCORS(mode string, credentials bool) *Fetch {
    f.corsMode = mode
    f.corsCredentials = credentials
    return f // Chainable
}

// NewClient creates a configured HTTP client.
func (f *Fetch) NewClient(timeoutMS int) Client {
    return &client{
        timeoutMS:      timeoutMS,
        fetchgo:        f, // Reference to parent
        defaultHeaders: make(map[string]string),
    }
}
```

### `Client` Interface: The Public API

The user interacts with a `Client` interface, which defines the available methods for making requests. This prevents direct instantiation of the client struct and ensures all clients are created correctly via the `Fetch` manager.

**Key Design Decision:** No `baseURL` field. All URLs must be absolute. This allows a single client to make requests to multiple domains.

```go
// types.go or interfaces.go
package fetchgo

// Client defines the public interface for an HTTP client.
type Client interface {
    // SendJSON performs an HTTP request with JSON encoding.
    // url MUST be absolute (e.g., "https://api.example.com/users")
    // body is encoded to JSON, Content-Type: application/json
    // The callback receives the raw response body as []byte.
    // The user is responsible for decoding it using json.Unmarshal.
    SendJSON(method, url string, body any, callback func([]byte, error))
    
    // SendBinary performs an HTTP request with TinyBin encoding.
    // url MUST be absolute (e.g., "https://api.example.com/users")
    // body is encoded with TinyBin, Content-Type: application/octet-stream
    // The callback receives the raw response body as []byte.
    // The user is responsible for decoding it using tinybin.Decode.
    SendBinary(method, url string, body any, callback func([]byte, error))
    
    // SetHeader sets a default header for all requests from this client.
    SetHeader(key, value string)
}
```

### `client` struct: The (Private) Implementation

The `client` struct is the private implementation that holds request-specific configuration and a reference to its parent `Fetch` instance.

```go
// client.go
// client is the private implementation of the Client interface.
type client struct {
    defaultHeaders map[string]string
    timeoutMS      int
    fetchgo        *Fetch // Reference to parent for codec/config access
}

// SendJSON encodes body as JSON and sends HTTP request.
func (c *client) SendJSON(method, url string, body any, callback func([]byte, error)) {
    go func() {
        encoder := c.fetchgo.getJSONEncoder()
        result, err := c.doRequest(method, url, "application/json", encoder, body)
        callback(result, err)
    }()
}

// SendBinary encodes body with TinyBin and sends HTTP request.
func (c *client) SendBinary(method, url string, body any, callback func([]byte, error)) {
    go func() {
        encoder := c.fetchgo.getTinyBinEncoder()
        result, err := c.doRequest(method, url, "application/octet-stream", encoder, body)
        callback(result, err)
    }()
}

// SetHeader adds or updates a default header for all requests from this client.
func (c *client) SetHeader(key, value string) {
    c.defaultHeaders[key] = value
}
```

## 3. Automatic Codec Selection

The `Fetch` instance is responsible for providing the correct encoder/decoder based on the `Content-Type` header.

```go
// fetchgo.go (internal methods)

// getJSONEncoder returns platform-specific JSON encoder
func (f *Fetch) getJSONEncoder() encoder {
    return f.getJSONEncoder() // Platform-specific (stdlib or WASM)
}

// getTinyBinEncoder returns TinyBin encoder (cross-platform)
func (f *Fetch) getTinyBinEncoder() encoder {
    return &tinyBinEncoder{tb: f.tb}
}

// Note: No decoder is needed. The response is returned as raw []byte.
// The user is responsible for decoding using json.Unmarshal, tinybin.Decode, etc.
```

## 4. Platform-Specific Codecs

This is the key to cross-platform compatibility. We use build tags to provide different implementations of `getJSONEncoder()` and `getJSONDecoder()`.

### Backend (stdlib)

Uses the standard `encoding/json` package.

```go
// codecs_json_stdlib.go
//go:build !wasm

package fetchgo

import "encoding/json"

func (f *Fetch) getJSONEncoder() encoder { return &stdlibJSONEncoder{} }

// stdlibJSONEncoder encodes Go values to JSON []byte
type stdlibJSONEncoder struct{}

func (e *stdlibJSONEncoder) Encode(data any) ([]byte, error) {
    return json.Marshal(data)
}
```

### Frontend (WASM)

Uses the browser's native `JSON.stringify` and `JSON.parse` via `syscall/js`.

```go
// codecs_json_wasm.go
//go:build wasm

package fetchgo

import "syscall/js"

func (f *Fetch) getJSONEncoder() encoder { return &wasmJSONEncoder{} }

// wasmJSONEncoder uses browser's JSON.stringify
type wasmJSONEncoder struct{}

func (e *wasmJSONEncoder) Encode(data any) ([]byte, error) {
    // Convert Go value to JS, stringify, return as []byte
    jsValue := convertGoToJS(data)
    jsonString := js.Global().Get("JSON").Call("stringify", jsValue).String()
    return []byte(jsonString), nil
}
```

### Cross-Platform Codecs

The `tinyBinEncoder` works on both platforms without build tags.

```go
// codecs_shared.go
package fetchgo

import "webtyp.com/tinybin"

// tinyBinEncoder encodes data using TinyBin
type tinyBinEncoder struct {
    tb *tinybin.TinyBin
}

func (e *tinyBinEncoder) Encode(data any) ([]byte, error) {
    return e.tb.Encode(data)
}
```

## 5. CORS Configuration

CORS settings are stored in the `Fetch` instance and applied in `client_wasm.go` when making the `fetch` call.

```go
// client_wasm.go
func (c *client) doRequest(method, url, contentType string, encoder encoder, body any) ([]byte, error) {
    // ...
    options := js.Global().Get("Object").New()
    options.Set("method", method)
    
    // Apply CORS settings from parent
    if c.fetchgo != nil {
        options.Set("mode", c.fetchgo.corsMode)
        options.Set("credentials", boolToString(c.fetchgo.corsCredentials))
    }
    
    // Set Content-Type header
    headers := js.Global().Get("Object").New()
    headers.Set("Content-Type", contentType)
    
    // Add default headers
    for key, value := range c.defaultHeaders {
        headers.Set(key, value)
    }
    options.Set("headers", headers)
    
    // ...
    promise := js.Global().Call("fetch", url, options) // url is absolute
    // ...
}
```

## 6. Usage Examples

### Basic Usage (Single Domain)

```go
// Create the manager once
http := fetchgo.New().SetCORS("cors", true)

// Create a client with 5 second timeout
client := http.NewClient(5000)

// Configure authentication
client.SetHeader("Authorization", "Bearer token123")

// Send JSON request (explicitly using JSON)
var userData = User{Name: "John", Email: "john@example.com"}
client.SendJSON("POST", "https://api.example.com/users", 
    userData, func(body []byte, err error) {
        if err != nil {
            // Handle error
            return
        }
        // Decode the JSON response
        var user User
        if err := json.Unmarshal(body, &user); err != nil {
            // Handle decode error
            return
        }
        // Use user...
    })
```

### Advanced Usage (Multiple Domains)

```go
// Single client for all requests
http := fetchgo.New()
client := http.NewClient(5000)

// Request to third-party API using JSON (compatibility)
client.SendJSON("POST", "https://api.stripe.com/charges", 
    orderData, func(body []byte, err error) {
        if err == nil {
            var order Order
            json.Unmarshal(body, &order)
        }
    })

// Request to your microservice using TinyBin (performance)
client.SendBinary("POST", "https://payments.yourapp.com/charge", 
    paymentData, func(body []byte, err error) {
        if err == nil {
            var response PaymentResponse
            // Decode using TinyBin
            tinybin.New().Decode(body, &response)
        }
    })

// Another TinyBin request to different domain
client.SendBinary("GET", "https://api.yourapp.com/orders", 
    nil, func(body []byte, err error) {
        if err == nil {
            var orders []Order
            tinybin.New().Decode(body, &orders)
        }
    })
```

### Integration with CRUDP

`fetchgo` serves as the HTTP transport layer for CRUDP. Here's how CRUDP would use it:

```go
// pkg/crudp/client.go
package crudp

import "webtyp.com/fetch"

type Client struct {
    http      fetchgo.Client
    syncURL   string
    sseURL    string
    queue     []Packet
    pending   map[string]ResponseCallback
}

func NewCRUDPClient(baseURL string) *Client {
    // Create fetchgo client for HTTP transport
    httpClient := fetchgo.New().NewClient(5000)
    
    return &Client{
        http:    httpClient,
        syncURL: baseURL + "/sync",
        sseURL:  baseURL + "/events",
        queue:   make([]Packet, 0),
        pending: make(map[string]ResponseCallback),
    }
}

// Sync sends batch using fetchgo with TinyBin for performance
func (c *Client) Sync() error {
    batch := BatchRequest{Packets: c.queue}
    c.queue = nil
    
    // Use TinyBin encoding for efficient binary protocol
    c.http.SendBinary("POST", c.syncURL, batch, func(body []byte, err error) {
        if err != nil {
            // Handle error or retry
            return
        }
        
        // Decode TinyBin response
        var response BatchResponse
        if err := c.tb.Decode(body, &response); err != nil {
            // Handle decode error
            return
        }
        
        // Process response packets
        c.processBatchResponse(response)
    })
    
    return nil
}
```

This architecture provides a simple, powerful, and flexible API that:
- Hides the complexity of cross-platform development
- Supports multiple domains from a single client
- Serves as a solid foundation for higher-level protocols like CRUDP