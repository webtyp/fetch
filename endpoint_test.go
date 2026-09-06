package fetch_test

import (
	"os"
	"testing"

	"webtyp.com/fetch"
)

type MockUser struct {
	ID string
}

func (m MockUser) HandlerName() string {
	return "/post_json"
}

func TestEndpointResolution(t *testing.T) {
	t.Run("String endpoint", func(t *testing.T) {
		req := fetch.Get("/users")
		if req == nil {
			t.Fatal("Request should not be nil")
		}
	})

	t.Run("EndpointProvider interface", func(t *testing.T) {
		user := MockUser{ID: "42"}
		req := fetch.Get(user)
		if req == nil {
			t.Fatal("Request should not be nil")
		}
	})
}

func TestIntegration_BaseURL(t *testing.T) {
	urlBytes, err := os.ReadFile(".test_server_url")
	if err != nil {
		t.Skip("Test server URL not found, skipping integration test")
	}
	baseURL := string(urlBytes)

	t.Run("Global BaseURL", func(t *testing.T) {
		fetch.SetBaseURL(baseURL)
		defer fetch.SetBaseURL("")

		done := make(chan bool)
		var responseBody string
		var responseErr error

		fetch.Get("/get").Send(func(resp *fetch.Response, err error) {
			if err != nil {
				responseErr = err
			} else {
				responseBody = resp.Text()
			}
			done <- true
		})

		<-done

		if responseErr != nil {
			t.Fatalf("Expected no error, got %v", responseErr)
		}
		if responseBody != "get success" {
			t.Errorf("Expected 'get success', got '%s'", responseBody)
		}
	})

	t.Run("Per-request BaseURL override", func(t *testing.T) {
		fetch.SetBaseURL("https://invalid.url")
		defer fetch.SetBaseURL("")

		done := make(chan bool)
		var responseBody string
		var responseErr error

		fetch.Get("/get").
			BaseURL(baseURL).
			Send(func(resp *fetch.Response, err error) {
				if err != nil {
					responseErr = err
				} else {
					responseBody = resp.Text()
				}
				done <- true
			})

		<-done

		if responseErr != nil {
			t.Fatalf("Expected no error, got %v", responseErr)
		}
		if responseBody != "get success" {
			t.Errorf("Expected 'get success', got '%s'", responseBody)
		}
	})

	t.Run("EndpointProvider integration", func(t *testing.T) {
		fetch.SetBaseURL(baseURL)
		defer fetch.SetBaseURL("")

		user := MockUser{ID: "42"}
		done := make(chan bool)
		var status int

		fetch.Post(user).
			ContentTypeJSON().
			Body([]byte(`{"message":"hello"}`)).
			Send(func(resp *fetch.Response, err error) {
				if err == nil {
					status = resp.Status
				}
				done <- true
			})

		<-done
		if status != 200 {
			t.Errorf("Expected status 200, got %d", status)
		}
	})
}

func TestFetchEdgeCases(t *testing.T) {
	t.Run("GetBaseURL", func(t *testing.T) {
		fetch.SetBaseURL("test")
		if fetch.GetBaseURL() != "test" {
			t.Error("GetBaseURL failed")
		}
		fetch.SetBaseURL("")
	})

	t.Run("Response Body getter", func(t *testing.T) {
		resp := &fetch.Response{}
		if len(resp.Body()) != 0 {
			t.Error("Empty body expected")
		}
	})

	t.Run("Dispatch no handler error", func(t *testing.T) {
		fetch.SetHandler(nil)
		fetch.Get("/").Dispatch() // Should just log and return
	})

	t.Run("resolveEndpoint errors", func(t *testing.T) {
		// Nil endpoint
		req := fetch.Get(nil)
		req.Send(func(resp *fetch.Response, err error) {
			if err == nil {
				t.Error("Expected error for nil endpoint")
			}
		})

		// Invalid type
		req = fetch.Get(123)
		req.Send(func(resp *fetch.Response, err error) {
			if err == nil {
				t.Error("Expected error for invalid endpoint type")
			}
		})
	})

	t.Run("joinURLPath empty path", func(t *testing.T) {
		// Since we can't call internal functions, we test via buildFullURL if it was exposed
		// or just trust that usual paths cover it if we use empty string endpoint
		fetch.SetBaseURL("http://api.com")
		defer fetch.SetBaseURL("")
		fetch.Get("").Send(func(resp *fetch.Response, err error) {
			// buildURL should return error for empty endpoint
			if err == nil {
				t.Error("Expected error for empty endpoint")
			}
		})
	})

	t.Run("Logging", func(t *testing.T) {
		var logged bool
		fetch.SetLog(func(args ...any) {
			logged = true
		})
		fetch.SetHandler(nil)
		fetch.Get("/").Dispatch() // This should trigger a log message
		if !logged {
			t.Error("Logger was not called")
		}
		fetch.SetLog(nil)
	})
}
