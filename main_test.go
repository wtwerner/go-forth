package main

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

// Fetch data from a valid URL
func TestFetchDataSuccess(t *testing.T) {
	expectedBody := "Example Domain"

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, expectedBody)
	}))

	defer server.Close()

	result, err := FetchData(server.URL)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if result != expectedBody {
		t.Errorf("expected %v, got %v", expectedBody, result)
	}
}

// Return error from an invalid URL
func TestFetchDataInvalidURL(t *testing.T) {
	invalidURL := "://invalid-url"

	_, err := FetchData(invalidURL)
	if err == nil {
		t.Fatal("expected an error, got none")
	}

	expectedErrorMsg := "failed to make the request"
	if !strings.Contains(err.Error(), expectedErrorMsg) {
		t.Errorf("expected error message to contain %v, got %v", expectedErrorMsg, err.Error())
	}
}

// Make an HTTP client with a 10-second timeout
func TestReturnsHTTPClientWith10SecondTimeout(t *testing.T) {
	client := internalHTTPClient()

	if client.Timeout != 10*time.Second {
		t.Errorf("Expected timeout to be 10 seconds, got %v", client.Timeout)
	}
}

// Fetches data successfully from a valid URL
func TestFetchDataWithClient_Success(t *testing.T) {
	validURL := "http://example.com"
	expectedBody := "Hello, World!"

	client := &http.Client{
		Transport: roundTripperFunc(func(req *http.Request) *http.Response {
			return &http.Response{
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(strings.NewReader(expectedBody)),
			}
		}),
	}

	body, err := FetchDataWithClient(validURL, client)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if body != expectedBody {
		t.Fatalf("expected body %q, got %q", expectedBody, body)
	}
}

type roundTripperFunc func(req *http.Request) *http.Response

func (f roundTripperFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req), nil
}

// Handles invalid URL format
func TestFetchDataWithClient_InvalidURL(t *testing.T) {
	invalidURL := "://invalid-url"
	client := &http.Client{}

	_, err := FetchDataWithClient(invalidURL, client)
	if err == nil {
		t.Fatal("expected an error, got none")
	}

	expectedErrorMsg := "failed to make the request"
	if !strings.Contains(err.Error(), expectedErrorMsg) {
		t.Fatalf("expected error message to contain %q, got %q", expectedErrorMsg, err.Error())
	}
}

// Valid URL with http scheme returns true
func TestValidURLWithHTTPScheme(t *testing.T) {
	urlStr := "http://example.com"
	result := isValidURL(urlStr)
	if !result {
		t.Errorf("Expected true, got %v", result)
	}
}

// Empty string returns false
func TestEmptyStringReturnsFalse(t *testing.T) {
	urlStr := ""
	result := isValidURL(urlStr)
	if result {
		t.Errorf("Expected false, got %v", result)
	}
}
