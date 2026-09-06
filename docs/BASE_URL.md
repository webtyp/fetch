# Base URL and Endpoint Resolution

`webtyp/fetch` allows you to simplify your requests by using a base URL and providing only the endpoint.

## Priority Resolution

When building the final URL, the following priority is used:

1.  **Absolute URL**: If the endpoint passed to `Get`, `Post`, etc., starts with `http://` or `https://`, it is used directly.
2.  **Request BaseURL**: If `.BaseURL("...")` is called on the request builder.
3.  **Global BaseURL**: If `fetch.SetBaseURL("...")` was called previously.
4.  **WASM Origin**: In WebAssembly environments (browsers), it defaults to `location.origin`.
5.  **Error**: If none of the above are available, the request will fail with an error.

## Global Base URL

You can set a global base URL once at the start of your application:

```go
import "webtyp.com/fetch"

func check() {
    fetch.SetBaseURL("https://api.example.com")

    // Request to https://api.example.com/users
    fetch.Get("/users").Send(...)
}
```

## Per-Request Base URL

You can override the global base URL for a specific request:

```go
fetch.Get("/data").
    BaseURL("https://other-api.com").
    Send(...)
```

## Automatic WASM Origin

In a browser, if you don't set a base URL, it automatically uses the current origin:

```go
// In browser at https://myapp.com
// Request to https://myapp.com/api/v1/status
fetch.Get("/api/v1/status").Send(...)
```

## Endpoint Provider Interface

You can pass structures that implement the `EndpointProvider` interface directly:

```go
type User struct {
    ID string
}

func (u User) HandlerName() string {
    return "/users"
}

func main() {
    user := User{ID: "123"}
    
    // Automatically uses "/users" as the endpoint
    fetch.Get(user).Send(...)
}
```

This is useful for creating type-safe API clients where the structure itself knows its endpoint.
