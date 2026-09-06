package fetch

import (
	. "webtyp.com/fmt"
)

// EndpointProvider allows structs to define their API endpoint
type EndpointProvider interface {
	HandlerName() string
}

// resolveEndpoint extracts endpoint string from any (string or EndpointProvider)
func resolveEndpoint(endpoint any) (string, error) {
	if endpoint == nil {
		return "", Err("endpoint cannot be nil")
	}

	switch v := endpoint.(type) {
	case string:
		return v, nil
	case EndpointProvider:
		return v.HandlerName(), nil
	default:
		return "", Errf("unsupported endpoint type: %T", endpoint)
	}
}

// buildFullURL constructs final URL using BaseURL + endpoint
func buildFullURL(endpoint string, requestBaseURL string) (string, error) {
	if isAbsoluteURL(endpoint) {
		return endpoint, nil
	}

	var base string
	if requestBaseURL != "" {
		base = requestBaseURL
	} else if defaultBaseURL != "" {
		base = defaultBaseURL
	} else {
		base = getOrigin()
	}

	if base == "" {
		return "", Err("BaseURL not set, provide absolute URL or call SetBaseURL()")
	}

	return joinURLPath(base, endpoint), nil
}

// isAbsoluteURL checks if URL starts with "http://" or "https://"
func isAbsoluteURL(url string) bool {
	return HasPrefix(url, "http://") || HasPrefix(url, "https://")
}

// joinURLPath joins base and path, normalizing slashes
func joinURLPath(base, path string) string {
	if path == "" {
		return base
	}
	// Use PathJoin for normalization. PathJoin is designed for file paths but
	// since URLs use '/' it works well for simple joins.
	return PathJoin(base, path).String()
}
