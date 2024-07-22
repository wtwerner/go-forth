# go-forth is a HTTP Client for the CLI

This is a proof of concept for making HTTP GET requests in Go and pretty-printing the JSON response using `lipgloss`.

## Setup

1. Install Go: https://golang.org/doc/install
2. Clone the repository:
    ```sh
    git clone https://github.com/yourusername/go-http-client-poc.git
    cd go-http-client-poc
    ```
3. Install dependencies:
    ```sh
    go get -u github.com/charmbracelet/lipgloss
    ```

## Usage

Run the application:

  ```sh
  go run main.go
  ```

The application uses the environment variable `FETCH_URL` to determine where to send the HTTP request.