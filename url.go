package fetch

import (
	. "webtyp.com/fmt"
)

var defaultBaseURL string

// SetBaseURL sets the global base URL for all requests.
func SetBaseURL(url string) {
	defaultBaseURL = url
}

// GetBaseURL returns the current global base URL.
func GetBaseURL() string {
	return defaultBaseURL
}

// buildURL constructs the full request URL using the new resolution logic.
func buildURL(r *Request) (string, error) {
	endpoint, err := resolveEndpoint(r.endpoint)
	if err != nil {
		return "", err
	}

	if endpoint == "" {
		return "", Err("endpoint cannot be empty")
	}

	return buildFullURL(endpoint, r.baseURL)
}
