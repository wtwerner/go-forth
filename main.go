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

	tea "github.com/charmbracelet/bubbletea"
	gloss "github.com/charmbracelet/lipgloss"
)

type simplePage struct{ text string }

func newSimplePage(text string) simplePage {
	return simplePage{text: text}
}

func (s simplePage) Init() tea.Cmd { return nil }

func (s simplePage) View() string {
	return fmt.Sprintf(
		"%s\n\nPress Ctrl+C to exit",
		s.text,
	)
}

func (s simplePage) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if msg.String() == "ctrl+c" {
			return s, tea.Quit
		}
	}
	return s, nil
}

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

func internalHTTPClient() *http.Client {
	return &http.Client{
		Timeout: 10 * time.Second,
	}
}

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

	var prettyJSON bytes.Buffer
	err = json.Indent(&prettyJSON, []byte(data), "", "  ")
	if err != nil {
		log.Fatalf("Error formatting JSON: %v", err)
	}

	// Style the output using lipgloss
	style := gloss.NewStyle().
		Foreground(gloss.Color("205")).
		Background(gloss.Color("236")).
		Padding(1).
		Margin(1).
		Border(gloss.RoundedBorder()).
		BorderForeground(gloss.Color("63"))

	p := tea.NewProgram(
		newSimplePage(style.Render(prettyJSON.String())),
	)
	if err := p.Start(); err != nil {
		panic(err)
	}
}
