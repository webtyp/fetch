//go:build !wasm

package fetch

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"time"

	. "webtyp.com/fmt"
)

// doRequest is the standard library implementation for making an HTTP request.
func doRequest(r *Request, callback func(*Response, error)) {
	go func() {
		// 1. Build the full URL.
		fullURL, err := buildURL(r)
		if err != nil {
			callback(nil, err)
			return
		}

		// 2. Prepare body reader.
		var bodyReader io.Reader
		if len(r.body) > 0 {
			bodyReader = bytes.NewReader(r.body)
		}

		// 3. Set up the request context with timeout.
		ctx := context.Background()
		if r.timeout > 0 {
			var cancel context.CancelFunc
			ctx, cancel = context.WithTimeout(ctx, time.Duration(r.timeout)*time.Millisecond)
			defer cancel()
		}

		// 4. Create the HTTP request.
		req, err := http.NewRequestWithContext(ctx, r.method, fullURL, bodyReader)
		if err != nil {
			callback(nil, Errf("failed to create request: %s", err.Error()))
			return
		}

		// 5. Add headers to the request.
		for _, h := range r.headers {
			req.Header.Add(h.Key, h.Value)
		}

		// 6. Execute the request.
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			callback(nil, Errf("request failed: %s", err.Error()))
			return
		}
		defer resp.Body.Close()

		// 7. Read the response body.
		responseBody, err := io.ReadAll(resp.Body)
		if err != nil {
			callback(nil, Errf("failed to read response body: %s", err.Error()))
			return
		}

		// 8. Construct the Response object.
		var headers []Header
		for k, v := range resp.Header {
			for _, val := range v {
				headers = append(headers, Header{Key: k, Value: val})
			}
		}

		response := &Response{
			Status:     resp.StatusCode,
			Headers:    headers,
			RequestURL: fullURL,
			Method:     r.method,
			body:       responseBody,
		}

		callback(response, nil)
	}()
}

func getOrigin() string {
	return ""
}
