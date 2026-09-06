# CORS (Cross-Origin Resource Sharing)

When using `webtyp/fetch` in a browser environment (WASM), you may encounter CORS errors. It is important to understand that CORS is a **security feature implemented by the browser**, not a limitation of this library.

## What is CORS?

CORS is a mechanism that uses additional HTTP headers to tell browsers to give a web application running at one origin access to selected resources from a different origin.

If your WASM app is running on `https://myapp.com` and tries to fetch from `https://api.example.com`, the browser will block the request unless the server explicitly allows it.

## Common Error Symptoms

In the browser console, you might see:
- `Access to fetch at '...' from origin '...' has been blocked by CORS policy`
- `Failed to fetch` (with no further detail in Go, but detail in browser console)

## How to Fix It (Server-Side)

The server receiving the request must return specific headers. If you are using Go for your backend, you need to add a middleware that adds these headers:

```go
func corsMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        // Allow the origin of your WASM app
        w.Header().Set("Access-Control-Allow-Origin", "*") // Or your specific domain
        w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
        w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

        // Handle preflight (OPTIONS) requests
        if r.Method == http.MethodOptions {
            w.WriteHeader(http.StatusOK)
            return
        }

        next.ServeHTTP(w, r)
    })
}
```

## Special Case: Custom Headers

If you use custom headers (e.g., via `.Header("X-My-Header", "value")`), you must:
1. Allow them in `Access-Control-Allow-Headers`.
2. Expose them in `Access-Control-Expose-Headers` if you want to read them from the response in your WASM code.

## Resources

- [MDN Web Docs: CORS](https://developer.mozilla.org/en-US/docs/Web/HTTP/CORS)
