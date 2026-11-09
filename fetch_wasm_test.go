//go:build wasm

package fetch_test

import (
	"net/http"
	"os"
	"testing"
)

func TestWasm(t *testing.T) {
	// Read server URL from file (written by test server on startup)
	urlBytes, err := os.ReadFile(".test_server_url")
	if err != nil {
		t.Fatalf("Failed to read server URL from .test_server_url: %v. Make sure test server is running.", err)
	}
	serverURL := string(urlBytes)

	// No Client initialization needed for the simplified API

	// Shutdown server after tests
	defer func() {
		http.Get(serverURL + "/shutdown")
	}()

	t.Run("Get", func(t *testing.T) { SendRequest_GetShared(t, serverURL) })
	t.Run("PostJSON", func(t *testing.T) { SendRequest_PostJSONShared(t, serverURL) })
	t.Run("TimeoutSuccess", func(t *testing.T) { SendRequest_TimeoutSuccessShared(t, serverURL) })
	t.Run("TimeoutFailure", func(t *testing.T) { SendRequest_TimeoutFailureShared(t, serverURL) })
	t.Run("ServerError", func(t *testing.T) { SendRequest_ServerErrorShared(t, serverURL) })
	t.Run("PostFile", func(t *testing.T) { SendRequest_PostFileShared(t, serverURL) })
	t.Run("PutDelete", func(t *testing.T) { SendRequest_PutDeleteShared(t, serverURL) })
	t.Run("Patch", func(t *testing.T) { SendRequest_PatchShared(t, serverURL) })
	t.Run("DeleteWithBody", func(t *testing.T) { SendRequest_DeleteWithBodyShared(t, serverURL) })
	t.Run("Headers", func(t *testing.T) { SendRequest_HeadersShared(t, serverURL) })
	t.Run("ContentTypes", func(t *testing.T) { SendRequest_ContentTypesShared(t, serverURL) })
	t.Run("Dispatch", func(t *testing.T) { SendRequest_DispatchShared(t, serverURL) })
}
