# API Reference

## Top-Level Functions

### `func Get(url string) *Request`
Creates a new GET request.

### `func Post(url string) *Request`
Creates a new POST request.

### `func Put(url string) *Request`
Creates a new PUT request.

### `func Patch(url string) *Request`
Creates a new PATCH request: a partial update.

### `func Delete(url string) *Request`
Creates a new DELETE request.

### `func SetLog(fn func(...any))`
Sets a logger function for debugging.

### `func SetHandler(fn func(*Response))`
Sets the global handler for `Dispatch()` requests.

## Request

### `func (r *Request) Header(key, value string) *Request`
Adds a header to the request.

### `func (r *Request) Body(data []byte) *Request`
Sets the request body.

### `func (r *Request) Timeout(ms int) *Request`
Sets the request timeout in milliseconds.

### `func (r *Request) Send(callback func(*Response, error))`
Executes the request and calls the callback with the response.

### `func (r *Request) Dispatch()`
Executes the request and sends the response to the global handler.

## Response

### `func NewResponse(status int, headers []Header, body []byte) *Response`
Constructs a usable Response.

### `type Response struct`
- `Status int`: HTTP status code
- `Headers []Header`: Response headers
- `RequestURL string`: The URL requested
- `Method string`: The HTTP method used

### `func (r *Response) Body() []byte`
Returns the response body as a byte slice.

### `func (r *Response) Text() string`
Returns the response body as a string.

### `func (r *Response) GetHeader(key string) string`
Returns the value of the specified header (case-insensitive).
