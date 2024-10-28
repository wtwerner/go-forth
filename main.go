package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// HTTP client configuration
var client = &http.Client{Timeout: 10 * time.Second}

// Lip Gloss styles for JSON components and response display
var (
	keyStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))
	stringStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("34"))
	numberStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("178"))
	boolStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("207"))
	nullStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
	respStyle   = lipgloss.NewStyle().
			Foreground(lipgloss.Color("205")).
			Background(lipgloss.Color("236")).
			Padding(1).
			Margin(1).
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("63"))
)

const indentation = "  "

// Model definition and initialization
type model struct {
	text      string
	textInput textinput.Model
	quitting  bool
}

func initialModel() model {
	ti := textinput.New()
	ti.Placeholder = "Enter a valid URL"
	ti.Focus()
	ti.CharLimit = 256
	ti.Width = 30

	return model{text: "", textInput: ti}
}

// Bubble Tea program functions
func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "enter":
			input := m.textInput.Value()
			if !isValidURL(input) {
				m.text = `{ "error": "invalid URL, please try again" }`
				return m, nil
			}

			data, err := FetchData(input)
			if err != nil || strings.Contains(data, `"error"`) {
				// Directly assign the error message to avoid double-formatting
				m.text = data
				return m, nil
			}

			// Pretty-print the JSON response only if it's valid JSON data
			m.text, err = prettyPrintJSON(data)
			if err != nil {
				m.text = fmt.Sprintf(`{ "error": "error formatting JSON", "details": "%v" }`, err)
			}

		case "ctrl+c":
			m.quitting = true
			return m, tea.Quit
		}
	}

	var cmd tea.Cmd
	m.textInput, cmd = m.textInput.Update(msg)
	return m, cmd
}

func (m model) View() string {
	if m.quitting {
		return "Thanks for using go-forth!\n"
	}

	content := respStyle.Render(m.text)
	return fmt.Sprintf(
		"\nPlease enter a URL for a GET request:\n\n%s\n\n%s\n\n%s\n",
		m.textInput.View(),
		content,
		"Press ctrl+c to exit",
	)
}

// Fetch data and format error responses as JSON
func FetchData(url string) (string, error) {
	resp, err := client.Get(url)
	if err != nil {
		return formatJSONError("failed to make the request", err.Error()), nil
	}
	defer resp.Body.Close()

	// First, check if the response status code is not 200 OK
	if resp.StatusCode != http.StatusOK {
		return formatJSONError("received non-200 response code", fmt.Sprintf("%d", resp.StatusCode)), nil
	}

	// Check if content type contains "application/json"
	contentType := resp.Header.Get("Content-Type")
	if !strings.Contains(contentType, "application/json") {
		body, _ := io.ReadAll(resp.Body)
		return formatJSONError("response is not JSON", truncateString(string(body), 100)), nil
	}

	// Read the JSON body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return formatJSONError("failed to read the response body", err.Error()), nil
	}

	// Ensure the body is valid JSON
	if !isJSON(body) {
		return formatJSONError("invalid JSON format", string(body)), nil
	}

	return string(body), nil
}

// Helper functions for data validation and formatting
func isValidURL(input string) bool {
	parsedURL, err := url.ParseRequestURI(input)
	return err == nil && parsedURL.Scheme != "" && parsedURL.Host != ""
}

func isJSON(data []byte) bool {
	var js json.RawMessage
	return json.Unmarshal(data, &js) == nil
}

func formatJSONError(message, details string) string {
	return fmt.Sprintf(`{ "error": "%s", "details": "%s" }`, message, details)
}

func truncateString(str string, length int) string {
	if len(str) <= length {
		return str
	}
	return str[:length] + "..."
}

// JSON Pretty-printing with color
func prettyPrintJSON(data string) (string, error) {
	var jsonData interface{}
	if err := json.Unmarshal([]byte(data), &jsonData); err != nil {
		return "", fmt.Errorf(`{ "error": "invalid JSON format", "details": "%v" }`, err)
	}
	return renderJSON(jsonData, 0), nil
}

func renderJSON(data interface{}, level int) string {
	var buf bytes.Buffer
	indent := strings.Repeat(indentation, level)

	switch v := data.(type) {
	case map[string]interface{}:
		buf.WriteString("{\n")
		for key, value := range v {
			buf.WriteString(indent + indentation)
			buf.WriteString(keyStyle.Render(fmt.Sprintf(`"%s"`, key)) + ": ")
			buf.WriteString(renderJSON(value, level+1))
			buf.WriteString(",\n")
		}
		buf.Truncate(buf.Len() - 2)
		buf.WriteString("\n" + indent + "}")

	case []interface{}:
		buf.WriteString("[\n")
		for _, item := range v {
			buf.WriteString(indent + indentation)
			buf.WriteString(renderJSON(item, level+1))
			buf.WriteString(",\n")
		}
		buf.Truncate(buf.Len() - 2)
		buf.WriteString("\n" + indent + "]")

	case string:
		buf.WriteString(stringStyle.Render(fmt.Sprintf(`"%s"`, v)))
	case float64:
		buf.WriteString(numberStyle.Render(fmt.Sprintf("%v", v)))
	case bool:
		buf.WriteString(boolStyle.Render(fmt.Sprintf("%v", v)))
	case nil:
		buf.WriteString(nullStyle.Render("null"))
	}

	return buf.String()
}

func main() {
	p := tea.NewProgram(initialModel())
	_, err := p.Run()
	if err != nil {
		fmt.Printf("Error running program: %v\n", err)
	}
}
