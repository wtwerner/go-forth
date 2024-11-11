package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/stretchr/testify/assert"
)

// Test helper for checking valid and invalid URLs
func TestIsValidURL(t *testing.T) {
	assert.True(t, isValidURL("https://example.com"))
	assert.True(t, isValidURL("http://example.com"))
	assert.True(t, isValidURL("ftp://example.com"))
	assert.False(t, isValidURL("not-a-url"))
}

// Test helper for validating HTTP methods
func TestIsValidHTTPMethod(t *testing.T) {
	validMethods := []string{"GET", "POST", "PUT", "DELETE", "PATCH", "get", "post", "put", "delete", "patch"}
	invalidMethods := []string{"FETCH", "POSTS", ""}

	for _, method := range validMethods {
		assert.True(t, httpMethods[strings.ToUpper(method)])
	}

	for _, method := range invalidMethods {
		assert.False(t, httpMethods[strings.ToUpper(method)], "Expected invalid method: %s", method)
	}
}

// Mock server to simulate JSON and non-JSON responses
func mockServer() *httptest.Server {
	mux := http.NewServeMux()
	mux.HandleFunc("/json", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"message": "Hello, JSON!"})
	})
	mux.HandleFunc("/plain", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		w.Write([]byte("Hello, plain text!"))
	})
	return httptest.NewServer(mux)
}

// Test FetchData function for both JSON and plain text responses
func TestFetchData(t *testing.T) {
	server := mockServer()
	defer server.Close()

	// JSON Response
	resp, err := FetchData(server.URL+"/json", "GET")
	assert.NoError(t, err)
	assert.Contains(t, resp, `"message": "Hello, JSON!"`)

	// Plain Text Response
	resp, err = FetchData(server.URL+"/plain", "GET")
	assert.NoError(t, err)
	assert.Contains(t, resp, "Hello, plain text!")
}

func TestPrettyPrintText(t *testing.T) {
	rawText := "<html><body><p>Hello, World!</p></body></html>"
	expectedIndented := "<html>\n  <body>\n    <p>Hello, World!</p>\n  </body>\n</html>\n"
	formattedText, err := formatHTMLText(rawText)
	assert.NoError(t, err)
	assert.Equal(t, expectedIndented, formattedText)
}

// Test for pretty printing JSON
func TestPrettyPrintJSON(t *testing.T) {
	rawJSON := `{"message": "Hello, JSON!"}`
	formattedJSON, err := prettyPrintJSON(rawJSON)
	assert.NoError(t, err)
	assert.Contains(t, formattedJSON, `"message": "Hello, JSON!"`)
}

// Test for Update function to verify key behaviors
func TestUpdateFunction(t *testing.T) {
	m := initialModel()

	// Test valid URL and method entry
	m.urlInput.SetValue("https://example.com")
	m.methodInput.SetValue("GET")
	msg := tea.KeyMsg{Type: tea.KeyEnter}
	updatedModel, _ := m.Update(msg)
	m = updatedModel.(model)
	assert.NotContains(t, m.text, "error", "Expected no error in valid input")

	// Test invalid URL entry
	m.urlInput.SetValue("not-a-url")
	updatedModel, _ = m.Update(msg)
	m = updatedModel.(model)
	assert.Contains(t, m.text, "invalid URL", "Expected invalid URL error")

	// Test invalid HTTP method
	m.urlInput.SetValue("https://example.com")
	m.methodInput.SetValue("FETCH")
	updatedModel, _ = m.Update(msg)
	m = updatedModel.(model)
	assert.Contains(t, m.text, "invalid HTTP method", "Expected invalid HTTP method error")
}
