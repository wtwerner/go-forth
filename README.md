# go-forth: An HTTP Client for the CLI

`go-forth` is a CLI-based HTTP client written in Go, designed to make HTTP GET requests and pretty-print JSON responses using `lipgloss` for enhanced readability. This tool demonstrates how to build a minimal CLI HTTP client with JSON formatting and error handling.

## Features

- **Send HTTP GET Requests** to any valid URL entered via the CLI.
- **Pretty-Print JSON Responses** using the `lipgloss` library for styled output.
- **Robust Error Handling** with detailed error messages for invalid URLs, non-JSON responses, and unexpected HTTP status codes.
- **Testing** includes mocked HTTP responses to simulate various server responses.

## Setup

1. **Install Go**: Make sure you have Go installed. [Go installation instructions](https://golang.org/doc/install)
2. **Clone the Repository**:
    ```sh
    git clone https://github.com/wtwerner/go-forth.git
    cd go-forth
    ```
3. **Install Dependencies**:
    ```sh
    go get -u github.com/charmbracelet/lipgloss
    go get -u github.com/charmbracelet/bubbletea
    ```

## Usage

Run the application from the command line:

```sh
go run main.go
```

After running the application, enter a URL for a GET request when prompted in the CLI. The application will fetch the data from the specified URL and display the JSON response in a styled format.

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

## Environment Variable

You can set a default URL by using the environment variable FETCH_URL. If this variable is set, the application will use it as the default URL for GET requests.

### Example:

```sh
export FETCH_URL=https://api.github.com
go run main.go
```

## Testing

The project includes a test suite that covers URL validation, error handling, JSON formatting, and simulated HTTP responses.

To run the tests:

```sh
go test -v
```

Tests use httptest to mock HTTP responses, simulating different scenarios, such as valid JSON responses, non-200 status codes, and non-JSON content types.

## Project Structure

- `main.go`: Contains the core CLI application logic, HTTP request handling, JSON formatting, and error handling.
- `main_test.go`: Includes unit tests and mock server responses to ensure the application behaves correctly under various conditions.

## Dependencies

- [Bubble Tea](https://github.com/charmbracelet/bubbletea) - For building the CLI interface
- [Lip Gloss](https://github.com/charmbracelet/lipgloss) - For styling the JSON output in the terminal

---

Feel free to contribute, report issues, or suggest improvements for `go-forth` by opening an issue or pull request on GitHub.
