package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"golang.org/x/net/html"
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

type component int

const (
	textInputFocus component = iota
	methodListFocus
)

// Model definition and initialization
type model struct {
	text             string
	urlInput         textinput.Model
	methodInput      textinput.Model
	focusedComponent component
	quitting         bool
}

func initialModel() model {
	const defaultWidth = 40

	url := textinput.New()
	url.Placeholder = "Enter a valid URL"
	url.Focus()
	url.CharLimit = 256
	url.Width = defaultWidth

	method := textinput.New()
	method.Placeholder = "HTTP Method"
	method.CharLimit = 6

	return model{text: "", urlInput: url, methodInput: method, focusedComponent: textInputFocus}
}

// Bubble Tea program functions
func (m model) Init() tea.Cmd {
	return nil
}

var httpMethods = map[string]bool{
	"GET":    true,
	"POST":   true,
	"PUT":    true,
	"DELETE": true,
	"PATCH":  true,
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			m.quitting = true
			return m, tea.Quit

		case "enter":
			// Validate URL
			input := m.urlInput.Value()
			if !isValidURL(input) {
				m.text = `{ "error": "invalid URL, please try again" }`
				return m, nil
			}

			// Validate HTTP Method
			method := strings.ToUpper(m.methodInput.Value())
			if !httpMethods[method] {
				m.text = `{ "error": "invalid HTTP method, please enter GET, POST, PUT, DELETE, or PATCH" }`
				return m, nil
			}

			// Fetch and format data with the validated method
			data, err := FetchData(input, method)
			if err != nil {
				m.text = data
			} else {
				m.text = data
			}
			return m, nil

		case "down", "j":
			if m.focusedComponent == textInputFocus {
				m.focusedComponent = methodListFocus
				m.urlInput.Blur()
				m.methodInput.Focus()
			}
			return m, nil

		case "up", "k":
			if m.focusedComponent == methodListFocus {
				m.focusedComponent = textInputFocus
				m.methodInput.Blur()
				m.urlInput.Focus()
			}
			return m, nil

		case "tab":
			// Toggle focus between urlInput and methodInput on Tab key press
			if m.focusedComponent == methodListFocus {
				m.focusedComponent = textInputFocus
				m.methodInput.Blur()
				m.urlInput.Focus()
			} else {
				m.focusedComponent = methodListFocus
				m.urlInput.Blur()
				m.methodInput.Focus()
			}
			return m, nil
		}
	}

	var cmd tea.Cmd
	// Update the input component based on which one is focused
	if m.focusedComponent == textInputFocus {
		m.urlInput, cmd = m.urlInput.Update(msg)
	} else {
		m.methodInput, cmd = m.methodInput.Update(msg)
	}

	return m, cmd
}

func (m model) View() string {
	if m.quitting {
		return "Thanks for using go-forth!\n"
	}

	content := respStyle.Render(m.text)
	return fmt.Sprintf(
		"\nPlease enter a URL for a GET request:\n\n%s\n\n%s\n\n%s\n%s\n",
		m.urlInput.View(),
		m.methodInput.View(),
		content,
		"Press ctrl+c to exit",
	)
}

func FetchData(url, method string) (string, error) {
	// Create a new request with the selected method
	req, err := http.NewRequest(method, url, nil)
	if err != nil {
		return formatJSONError("failed to create the request", err.Error()), nil
	}

	// Execute the request
	resp, err := client.Do(req)
	if err != nil {
		return formatJSONError("failed to make the request", err.Error()), nil
	}
	defer resp.Body.Close()

	// Check if the response status code is not 200 OK
	if resp.StatusCode != http.StatusOK {
		return formatJSONError("received non-200 response code", fmt.Sprintf("%d", resp.StatusCode)), nil
	}

	// Read the response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return formatJSONError("failed to read the response body", err.Error()), nil
	}

	// Determine response format based on content type
	contentType := resp.Header.Get("Content-Type")
	if strings.Contains(contentType, "application/json") && isJSON(body) {
		// Attempt to pretty-print JSON
		prettyJSON, err := prettyPrintJSON(string(body))
		if err != nil {
			return formatJSONError("error formatting JSON", err.Error()), nil
		}
		return prettyJSON, nil
	}

	// If not JSON, return as plain text with styling
	return prettyPrintText(string(body)), nil
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

func formatHTMLText(data string) (string, error) {
	// Parse the HTML
	node, err := html.Parse(strings.NewReader(data))
	if err != nil {
		return "", err
	}

	// Use a buffer to capture formatted output
	var buf bytes.Buffer
	formatNode(&buf, node, 0)
	return buf.String(), nil
}

func formatNode(buf *bytes.Buffer, n *html.Node, level int) {
	// Skip the root node and <head> element for formatting purposes
	if n.Type == html.DocumentNode || (n.Type == html.ElementNode && n.Data == "head") {
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			formatNode(buf, c, level)
		}
		return
	}

	// Check if the node has only one child and that child is a text node
	if n.Type == html.ElementNode && n.FirstChild != nil && n.FirstChild == n.LastChild && n.FirstChild.Type == html.TextNode {
		// Inline text content within tags
		indent(buf, level)
		buf.WriteString("<" + n.Data + ">")
		buf.WriteString(strings.TrimSpace(n.FirstChild.Data)) // Inline text
		buf.WriteString("</" + n.Data + ">\n")
		return
	}

	// Add opening tag with indentation
	if n.Type == html.ElementNode {
		indent(buf, level)
		buf.WriteString("<" + n.Data + ">\n")
	}

	// Process child nodes
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		formatNode(buf, c, level+1)
	}

	// Add closing tag for element nodes
	if n.Type == html.ElementNode {
		indent(buf, level)
		buf.WriteString("</" + n.Data + ">\n")
	} else if n.Type == html.TextNode {
		// Add text content with indentation for multi-line text
		text := strings.TrimSpace(n.Data)
		if text != "" {
			indent(buf, level)
			buf.WriteString(text + "\n")
		}
	}
}

func indent(buf *bytes.Buffer, level int) {
	buf.WriteString(strings.Repeat("  ", level))
}

func prettyPrintText(data string) string {
	// Apply indentation using formatHTMLText, then style with lipgloss
	formattedText, err := formatHTMLText(data)
	if err != nil {
		return formatJSONError("error formatting text", err.Error())
	}

	return lipgloss.NewStyle().
		Foreground(lipgloss.Color("250")).
		Background(lipgloss.Color("235")).
		Padding(1).
		Margin(1).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("63")).
		Render(truncateString(formattedText, 2000))
}

func prettyPrintJSON(data string) (string, error) {
	if os.Getenv("TEST_MODE") == "true" {
		// Skip pretty-printing during tests
		return data, nil
	}

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
