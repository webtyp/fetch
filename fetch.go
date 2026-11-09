package fetch

import (
	. "webtyp.com/fmt"
)

// Header represents a single HTTP header key-value pair.
type Header struct {
	Key   string
	Value string
}

// Request represents an HTTP request builder.
type Request struct {
	method   string
	endpoint any
	baseURL  string // per-request override
	headers  []Header
	body     []byte
	timeout  int
}

// Response represents an HTTP response.
type Response struct {
	Status     int
	Headers    []Header
	RequestURL string
	Method     string
	body       []byte
}

// NewResponse constructs a usable Response (e.g. from a Service Worker
// handler responding to an intercepted FetchEvent).
//
//	status  - HTTP status code (200, 404, ...).
//	headers - Response headers ({Key, Value}).
//	body    - Raw serialized body.
func NewResponse(status int, headers []Header, body []byte) *Response {
	return &Response{
		Status:  status,
		Headers: headers,
		body:    body,
	}
}

// Get creates a new GET request.
func Get(endpoint any) *Request {
	return &Request{method: "GET", endpoint: endpoint}
}

// Post creates a new POST request.
func Post(endpoint any) *Request {
	return &Request{method: "POST", endpoint: endpoint}
}

// Put creates a new PUT request.
func Put(endpoint any) *Request {
	return &Request{method: "PUT", endpoint: endpoint}
}

// Patch creates a new PATCH request: a partial update.
//
//	fetch.Patch(url).ContentTypeJSON().Body(body).Send(done)
func Patch(endpoint any) *Request {
	return &Request{method: "PATCH", endpoint: endpoint}
}

// Delete creates a new DELETE request.
func Delete(endpoint any) *Request {
	return &Request{method: "DELETE", endpoint: endpoint}
}

// BaseURL sets a per-request base URL override.
func (r *Request) BaseURL(url string) *Request {
	r.baseURL = url
	return r
}

// Header adds a header to the request.
func (r *Request) Header(key, value string) *Request {
	r.headers = append(r.headers, Header{Key: key, Value: value})
	return r
}

// ContentTypeJSON sets Content-Type to application/json
func (r *Request) ContentTypeJSON() *Request {
	return r.Header("Content-Type", "application/json")
}

// ContentTypeBinary sets Content-Type to application/octet-stream
func (r *Request) ContentTypeBinary() *Request {
	return r.Header("Content-Type", "application/octet-stream")
}

// ContentTypeForm sets Content-Type to application/x-www-form-urlencoded
func (r *Request) ContentTypeForm() *Request {
	return r.Header("Content-Type", "application/x-www-form-urlencoded")
}

// ContentTypeText sets Content-Type to text/plain
func (r *Request) ContentTypeText() *Request {
	return r.Header("Content-Type", "text/plain")
}

// ContentTypeHTML sets Content-Type to text/html
func (r *Request) ContentTypeHTML() *Request {
	return r.Header("Content-Type", "text/html")
}

// Body sets the request body.
func (r *Request) Body(data []byte) *Request {
	r.body = data
	return r
}

// Timeout sets the request timeout in milliseconds.
func (r *Request) Timeout(ms int) *Request {
	r.timeout = ms
	return r
}

// Send executes the request and calls the callback with the response.
func (r *Request) Send(callback func(*Response, error)) {
	doRequest(r, callback)
}

// Dispatch executes the request and sends the response to the global handler.
// This is a fire-and-forget method.
func (r *Request) Dispatch() {
	if globalHandler == nil {
		log("Dispatch called but no global handler set")
		return
	}
	doRequest(r, func(resp *Response, err error) {
		if err != nil {
			log("Dispatch error:", err)
			return
		}
		globalHandler(resp)
	})
}

// Body returns the response body as a byte slice.
func (r *Response) Body() []byte {
	return r.body
}

// Text returns the response body as a string.
func (r *Response) Text() string {
	return string(r.body)
}

// GetHeader returns the value of the specified header.
// It is case-insensitive.
func (r *Response) GetHeader(key string) string {
	key = Convert(key).ToLower().String()
	for _, h := range r.Headers {
		if Convert(h.Key).ToLower().String() == key {
			return h.Value
		}
	}
	return ""
}
