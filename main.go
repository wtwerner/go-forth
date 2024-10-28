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

// Declare global variables for HTTP client
var client *http.Client

var (
	keyStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("205")) // Light pink for keys
	stringStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("34"))  // Green for strings
	numberStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("178")) // Yellow for numbers
	boolStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("207")) // Purple for booleans
	nullStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("240")) // Gray for null
	indentation = "  "                                                  // Two spaces for indentation
	respStyle   = lipgloss.NewStyle().
			Foreground(lipgloss.Color("205")).
			Background(lipgloss.Color("236")).
			Padding(1).
			Margin(1).
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("63"))
)

func initialModel() model {
	ti := textinput.New()
	ti.Placeholder = "Enter a valid URL"
	ti.Focus() // Set focus to the input field
	ti.CharLimit = 256
	ti.Width = 30

	return model{
		text:      "",
		textInput: ti,
		quitting:  false,
	}
}

type model struct {
	text      string
	textInput textinput.Model
	quitting  bool
	err       error
}

func (m model) Init() tea.Cmd {
	client = internalHTTPClient() // Initialize HTTP client
	return nil
}

func (m model) View() string {
	if m.quitting {
		return "Thanks for using go-forth!\n"
	}

	if m.err != nil {
		return fmt.Sprintf("Error: %v\n\n%s", m.err, m.textInput.View())
	}

	return fmt.Sprintf(
		"\nPlease enter a URL where you'd like to send a GET request:\n\n%s\n\n%s\n\n%s\n",
		m.textInput.View(),
		respStyle.Render(m.text), // Display the fetched data with styling
		"Press ctrl+c to exit",
	)
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "enter": // Validate on "enter"
			input := m.textInput.Value()
			if _, err := url.ParseRequestURI(input); err != nil {
				m.text = `{ "error": "invalid URL, please try again" }`
				return m, nil
			}

			data, err := FetchDataWithClient(input, client)
			if err != nil {
				m.text = data // Already formatted JSON error string from FetchDataWithClient
				return m, nil
			}

			prettyData, err := prettyPrintJSON(data)
			if err != nil {
				m.text = fmt.Sprintf(`{ "error": "error formatting JSON", "details": "%v" }`, err)
				return m, nil
			}

			m.text = prettyData

		case "ctrl+c":
			m.quitting = true
			return m, tea.Quit
		}
	}

	var cmd tea.Cmd
	m.textInput, cmd = m.textInput.Update(msg)
	return m, cmd
}

func FetchDataWithClient(url string, client *http.Client) (string, error) {
	resp, err := client.Get(url)
	if err != nil {
		return `{ "error": "failed to make the request", "details": "` + err.Error() + `" }`, nil
	}
	defer resp.Body.Close()

	// Check if the response status code is 200 OK
	if resp.StatusCode != http.StatusOK {
		return fmt.Sprintf(`{ "error": "received non-200 response code", "code": %d }`, resp.StatusCode), nil
	}

	// Ensure the response Content-Type contains "application/json"
	contentType := resp.Header.Get("Content-Type")
	if contentType == "" || !containsJSON(contentType) {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Sprintf(`{ "error": "response is not JSON", "content_type": "%s", "body": "%s" }`, contentType, truncateString(string(body), 100)), nil
	}

	// Read and attempt to parse the JSON response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return `{ "error": "failed to read the response body", "details": "` + err.Error() + `" }`, nil
	}

	// Validate JSON structure by unmarshaling
	var jsonCheck map[string]interface{}
	if err := json.Unmarshal(body, &jsonCheck); err != nil {
		return `{ "error": "invalid JSON format", "details": "` + err.Error() + `" }`, nil
	}

	return string(body), nil
}

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
		buf.Truncate(buf.Len() - 2) // Remove trailing comma and newline
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

// Helper function to check if Content-Type is JSON
func containsJSON(contentType string) bool {
	return strings.Contains(contentType, "application/json")
}

// Helper function to truncate long strings to a specific length
func truncateString(str string, num int) string {
	if len(str) <= num {
		return str
	}
	return str[:num] + "..."
}

func internalHTTPClient() *http.Client {
	return &http.Client{
		Timeout: 10 * time.Second,
	}
}

func main() {
	p := tea.NewProgram(
		initialModel(),
	)
	if err := p.Start(); err != nil {
		panic(err)
	}
}
