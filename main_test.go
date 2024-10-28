package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"reflect"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

// Helper function to compare JSON objects
func jsonEqual(a, b string) bool {
	var objA, objB map[string]interface{}
	if err := json.Unmarshal([]byte(a), &objA); err != nil {
		return false
	}
	if err := json.Unmarshal([]byte(b), &objB); err != nil {
		return false
	}
	return reflect.DeepEqual(objA, objB)
}

func TestIsValidURL(t *testing.T) {
	tests := []struct {
		input    string
		expected bool
	}{
		{"https://www.example.com", true},
		{"ftp://invalid-protocol.com", true},
		{"not-a-url", false},
		{"http://", false},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			if got := isValidURL(tt.input); got != tt.expected {
				t.Errorf("isValidURL(%s) = %v; want %v", tt.input, got, tt.expected)
			}
		})
	}
}

func TestTruncateString(t *testing.T) {
	tests := []struct {
		input    string
		length   int
		expected string
	}{
		{"short", 10, "short"},
		{"this is a long string", 10, "this is a ..."},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			if got := truncateString(tt.input, tt.length); got != tt.expected {
				t.Errorf("truncateString(%s, %d) = %v; want %v", tt.input, tt.length, got, tt.expected)
			}
		})
	}
}

func TestFormatJSONError(t *testing.T) {
	tests := []struct {
		message  string
		details  string
		expected string
	}{
		{"error", "details", `{ "error": "error", "details": "details" }`},
	}

	for _, tt := range tests {
		t.Run(tt.message, func(t *testing.T) {
			if got := formatJSONError(tt.message, tt.details); got != tt.expected {
				t.Errorf("formatJSONError(%s, %s) = %v; want %v", tt.message, tt.details, got, tt.expected)
			}
		})
	}
}

func TestPrettyPrintJSON(t *testing.T) {
	validJSON := `{"key": "value", "number": 123, "bool": true}`
	invalidJSON := `{invalid json}`

	tests := []struct {
		input    string
		expected string
		hasError bool
	}{
		{validJSON, `{`, false},
		{invalidJSON, `{ "error": "invalid JSON format"`, true},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result, err := prettyPrintJSON(tt.input)
			if (err != nil) != tt.hasError {
				t.Fatalf("prettyPrintJSON error = %v, wantError %v", err, tt.hasError)
			}
			if !tt.hasError && !strings.HasPrefix(result, tt.expected) {
				t.Errorf("prettyPrintJSON(%s) = %v; want prefix %v", tt.input, result, tt.expected)
			}
		})
	}
}

func TestFetchDataWithMockClient(t *testing.T) {
	tests := []struct {
		name        string
		statusCode  int
		body        string
		contentType string
		expected    string
	}{
		{"Valid JSON", http.StatusOK, `{"key": "value"}`, "application/json", `{"key": "value"}`},
		{"Non-200 Status", http.StatusNotFound, ``, "application/json", `{ "error": "received non-200 response code"`},
		{"Invalid JSON Content Type", http.StatusOK, "<html>not json</html>", "text/html", `{ "error": "response is not JSON"`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", tt.contentType)
				w.WriteHeader(tt.statusCode)
				w.Write([]byte(tt.body))
			}))
			defer mockServer.Close()

			client = mockServer.Client() // Use the mock client for testing
			data, _ := FetchData(mockServer.URL)
			if !strings.HasPrefix(data, tt.expected) {
				t.Errorf("FetchData() = %v; want prefix %v", data, tt.expected)
			}
		})
	}
}

func TestUpdateFunction(t *testing.T) {
	// Enable test mode to skip pretty-printing
	os.Setenv("TEST_MODE", "true")
	defer os.Unsetenv("TEST_MODE")

	// Create a mock server with a handler to simulate different responses
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.Path, "notfound") {
			w.WriteHeader(http.StatusNotFound) // Simulate 404 response
			return
		}
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"key": "value"}`)) // Simulate valid JSON response
	}))
	defer mockServer.Close()

	tests := []struct {
		name        string
		input       string
		expectedMsg string
	}{
		{"Invalid URL", "invalid-url", `{ "error": "invalid URL, please try again" }`},
		{"Valid URL but non-200", mockServer.URL + "/notfound", `{ "error": "received non-200 response code", "details": "404" }`},
		{"Valid URL with JSON", mockServer.URL, `{"key": "value"}`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := initialModel()
			m.textInput.SetValue(tt.input)

			msg := tea.KeyMsg{Type: tea.KeyEnter}
			updatedModel, _ := m.Update(msg)
			got := updatedModel.(model).text

			// Use jsonEqual for JSON comparison or prefix match for error messages
			if strings.Contains(got, `"error"`) {
				if !strings.HasPrefix(got, tt.expectedMsg) {
					t.Errorf("Update() = %v; want prefix %v", got, tt.expectedMsg)
				}
			} else {
				if !jsonEqual(got, tt.expectedMsg) {
					t.Errorf("Update() JSON = %v; want JSON %v", got, tt.expectedMsg)
				}
			}
		})
	}
}
