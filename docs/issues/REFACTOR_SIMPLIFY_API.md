# Refactor: Simplify Fetch API

Refactor the `webtyp/fetch` library to create a minimal, JavaScript-style API. Remove all unnecessary constructors and dependencies. The user controls encoding/decoding.

## Goals

1. Remove `New()` constructor and `NewClient()` pattern
2. Remove dependencies on `webtyp/json` and `webtyp/gobin`
3. Create simple top-level functions: `Get()`, `Post()`, `Put()`, `Delete()`
4. Add two async send methods: `Send(callback)` and `Dispatch()`
5. Use `[]Header` slice instead of `map[string]string` (TinyGo compatibility)
6. Add `SetLog()` for debugging and `SetHandler()` for global response handling
7. Remove all existing documentation referring to the old API and replace it with the new one.

## API Specification

### Top-Level Functions

```go
func Get(url string) *Request
func Post(url string) *Request  
func Put(url string) *Request
func Delete(url string) *Request

func SetLog(fn func(...any))           // debugging
func SetHandler(fn func(*Response))    // global handler for Dispatch
```

### Header Struct (TinyGo friendly, no maps)

```go
type Header struct {
    Key   string
    Value string
}
```

### Request Builder

```go
type Request struct {
    method   string
    url      string
    headers  []Header
    body     []byte
    timeout  int
}

func (r *Request) Header(key, value string) *Request
func (r *Request) Body(data []byte) *Request
func (r *Request) Timeout(ms int) *Request
func (r *Request) Send(callback func(*Response, error))  // immediate callback
func (r *Request) Dispatch()                              // fire and forget to SetHandler
```

### Response Struct

```go
type Response struct {
    Status     int
    Headers    []Header
    RequestURL string
    Method     string
    body       []byte
}

func (r *Response) Body() []byte
func (r *Response) Text() string
func (r *Response) GetHeader(key string) string
```

## Files to Modify

### Delete

- `interfaces.go` - Remove `Client` interface
- `fetchgo.go` - Remove `Fetch` struct and `New()` constructor

### Create

- `log.go` - Add `SetLog()`, `SetHandler()` and private variables

### Modify

- `client.go` - Rewrite with `Header`, `Request`, `Response` structs and `Get/Post/Put/Delete` functions
- `client_wasm.go` - Adapt `doRequest` for `Send()` and `Dispatch()` with `[]Header`
- `client_stdlib.go` - Adapt `doRequest` for `Send()` and `Dispatch()` with `[]Header`
- `go.mod` - Remove `webtyp/json` and `webtyp/gobin` dependencies
- `README.md` - Update with new API examples
- `docs/API.md` - Update API reference

### Tests

Adapt all existing tests to new API.

## Example Usage

```go
import (
    "webtyp.com/fetch"
    "webtyp.com/json"
)

func main() {
    fetch.SetLog(println)
    fetch.SetHandler(func(resp *fetch.Response) {
        log(resp.RequestURL, resp.Status)
    })

    var body []byte
    json.Encode(User{Name: "Alice"}, &body)

    // Send: immediate response
    fetch.Post("https://api.example.com/users").
        Header("Content-Type", "application/json").
        Body(body).
        Timeout(5000).
        Send(func(resp *fetch.Response, err error) {
            if err != nil { return }
            json.Decode(resp.Body(), &user)
        })
    
    // Dispatch: fire and forget to global handler
    fetch.Post("https://api.example.com/analytics").
        Body(data).
        Dispatch()
}
```

## Testing Instructions

Install the test runner:

```bash
go install webtyp.com/devflow/cmd/gotest@latest
```

Run tests:

```bash
gotest           # Quiet mode (default)
gotest -v        # Verbose mode
```

### What gotest does

1. Runs `go vet ./...`
2. Runs `go test ./...`
3. Runs `go test -race ./...` (stdlib tests only)
4. Calculates coverage
5. Auto-detects and runs WASM tests if found (`*Wasm*_test.go`)
6. Updates README badges

### Expected Output

**Quiet mode:**
```
✅ vet ok, ✅ tests stdlib ok, ✅ race detection ok, ✅ coverage: 71%
```

**With WASM tests:**
```
✅ vet ok, ✅ tests stdlib ok, ✅ race detection ok, ✅ coverage: 71%, ✅ tests wasm ok
```

### Exit Codes

- `0` - All tests passed
- `1` - Tests failed, vet issues, or race conditions detected

### WASM Tests Note

WASM tests (`*Wasm*_test.go`) may fail during execution because they run in a real browser environment using `chromedp`. 
- **Instruction**: You must implement these tests aiming for 100% coverage. 
- **Failure Handling**: If they fail due to browser connectivity or protocol issues, acknowledge the failure but proceed with the refactor. We will fix stability issues in a later stage.

## Breaking Changes

This is a complete API rewrite. All existing code using this library must be updated:

- Remove `fetch.New()` calls
- Remove `client := fg.NewClient(...)` pattern
- Replace `client.SendJSON()` with `fetch.Post().Body(encodedData).Send()`
- Replace `client.SendBinary()` with `fetch.Post().Body(encodedData).Send()`
- User must encode/decode data manually using `webtyp/json` or `webtyp/binary`
