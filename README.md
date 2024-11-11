
# go-forth: A CLI HTTP Client with Enhanced User Interaction

`go-forth` is a CLI-based HTTP client written in Go, designed for making various HTTP requests and displaying formatted responses. With support for multiple HTTP methods, styled JSON and HTML response formatting, and intuitive navigation, `go-forth` provides a minimal yet powerful demonstration of CLI application development in Go.

## Features

- **HTTP Method Selection**: Easily switch between `GET`, `POST`, `PUT`, `DELETE`, and `PATCH` requests using a dropdown input.
- **JSON and HTML Response Formatting**: Automatically detects and pretty-prints JSON responses with syntax highlighting and indents HTML responses for readability.
- **Error Handling**: Provides clear error messages for invalid URLs, unsupported methods, non-200 HTTP status codes, and unsupported response formats.
- **Enhanced CLI Interface**: Uses `Bubble Tea` for interactive text and dropdown inputs, making the CLI experience smooth and intuitive.
- **Keyboard Navigation**: Switch between inputs using `Tab` and arrow keys for seamless navigation.

## Setup

### Prerequisites

1. **Install Go**: Ensure you have Go installed. [Go installation instructions](https://golang.org/doc/install)

### Installation

1. **Clone the Repository**:
    ```sh
    git clone https://github.com/wtwerner/go-forth.git
    cd go-forth
    ```
2. **Install Dependencies**:
    ```sh
    go get -u github.com/charmbracelet/lipgloss
    go get -u github.com/charmbracelet/bubbletea
    go get -u golang.org/x/net/html
    ```

## Usage

Run the application from the command line:

```sh
go run main.go
```

### Interactive Mode

1. **Enter a URL**: Type the URL for the desired endpoint.
2. **Select an HTTP Method**: Use `Tab` to switch to the method dropdown and choose from `GET`, `POST`, `PUT`, `DELETE`, and `PATCH`.
3. **View Formatted Response**: The application fetches the data and displays JSON responses with syntax highlighting or indented HTML responses for readability.

### Example

```plaintext
Please enter a URL for a GET request:
> https://api.github.com

{
  "current_user_url": "https://api.github.com/user",
  "authorizations_url": "https://api.github.com/authorizations",
  ...
}
```

## Testing

The project includes a comprehensive test suite covering:

- **URL and HTTP Method Validation**: Ensures URLs and methods are validated correctly.
- **Error Handling**: Simulates various error scenarios (e.g., invalid URL, unsupported method).
- **JSON and HTML Formatting**: Verifies that JSON and HTML responses are formatted properly.
  
To run the tests:

```sh
go test -v
```

Tests use `httptest` to mock HTTP responses, allowing simulation of different server responses (e.g., JSON data, non-200 status codes, plain text).

## Project Structure

- `main.go`: Contains the core CLI application logic, including HTTP requests, input handling, and response formatting.
- `main_test.go`: Includes unit tests with mocked responses to ensure accurate application behavior.

## Dependencies

- [Bubble Tea](https://github.com/charmbracelet/bubbletea) - For building the interactive CLI interface
- [Lip Gloss](https://github.com/charmbracelet/lipgloss) - For styling JSON and HTML responses in the terminal
- [HTML Parser](https://pkg.go.dev/golang.org/x/net/html) - For structured formatting of HTML responses

---

Feel free to contribute, report issues, or suggest improvements for `go-forth` by opening an issue or pull request on GitHub.
