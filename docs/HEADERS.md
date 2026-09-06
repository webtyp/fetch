# HTTP Headers

`webtyp/fetch` provides easy ways to set request headers and retrieve response headers.

## Request Headers

### Custom Headers
You can set any custom header using the `.Header(key, value)` method:

```go
fetch.Get("/api/settings").
    Header("Authorization", "Bearer token123").
    Header("X-Custom-Header", "Value").
    Send(...)
```

### Content-Type Helpers
The library includes declarative helpers for the most common `Content-Type` headers:

| Helper | Content-Type |
| --- | --- |
| `.ContentTypeJSON()` | `application/json` |
| `.ContentTypeBinary()` | `application/octet-stream` |
| `.ContentTypeForm()` | `application/x-www-form-urlencoded` |
| `.ContentTypeText()` | `text/plain` |
| `.ContentTypeHTML()` | `text/html` |

Example usage:
```go
fetch.Post("/users").
    ContentTypeJSON().
    Body([]byte(`{"name":"Alice"}`)).
    Send(...)
```

## Response Headers

You can retrieve headers from the `Response` object using the `GetHeader(key)` method. This method is **case-insensitive**.

```go
fetch.Get("/data").Send(func(resp *fetch.Response, err error) {
    if err != nil {
        return
    }

    // Get a specific header (case-insensitive)
    contentType := resp.GetHeader("content-type")
    server := resp.GetHeader("Server")
    
    println("Content-Type:", contentType)
    println("Server:", server)
})
```

### Accessing All Headers
You can also iterate over all headers if needed:

```go
for _, h := range resp.Headers {
    println(h.Key, ":", h.Value)
}
```
