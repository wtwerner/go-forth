package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"time"

	"github.com/charmbracelet/lipgloss"
)

// FetchData makes an HTTP GET request to the given URL and returns the response body as a string.
func FetchData(url string) (string, error) {
	resp, err := http.Get(url)
	if err != nil {
		return "", fmt.Errorf("failed to make the request: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read the response body: %v", err)
	}

	return string(body), nil
}

// internalHTTPClient returns a custom HTTP client with a timeout.
func internalHTTPClient() *http.Client {
	return &http.Client{
		Timeout: 10 * time.Second,
	}
}

// FetchDataWithClient makes an HTTP GET request to the given URL using the provided HTTP client and returns the response body as a string.
func FetchDataWithClient(url string, client *http.Client) (string, error) {
	resp, err := client.Get(url)
	if err != nil {
		return "", fmt.Errorf("failed to make the request to %s: %v", url, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("received non-200 response code: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read the response body from %s: %v", url, err)
	}

	return string(body), nil
}

// isValidURL validates the given URL.
func isValidURL(urlStr string) bool {
	parsedURL, err := url.ParseRequestURI(urlStr)
	if err != nil || parsedURL.Scheme == "" || parsedURL.Host == "" {
		return false
	}
	return true
}

func main() {
	url := os.Getenv("FETCH_URL")
	if url == "" {
		url = "https://api.github.com"
	}

	if !isValidURL(url) {
		log.Fatalf("Error: Invalid URL")
	}
	client := internalHTTPClient() // Get internal HTTP client
	data, err := FetchDataWithClient(url, client)
	if err != nil {
		log.Fatalf("Error: %v", err)
	}

	// Pretty-print the JSON data
	var prettyJSON bytes.Buffer
	err = json.Indent(&prettyJSON, []byte(data), "", "  ")
	if err != nil {
		log.Fatalf("Error formatting JSON: %v", err)
	}

	// Style the output using lipgloss
	style := lipgloss.NewStyle().
		Foreground(lipgloss.Color("205")).
		Background(lipgloss.Color("236")).
		Padding(1).
		Margin(1).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("63"))

	fmt.Println(style.Render(prettyJSON.String()))
}
