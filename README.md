# Go Browser Inventory

`go-browser-inventory` is a command-line tool written in Go that scans and lists browser extensions for Chrome, Edge, and Firefox. It provides output in either a human-readable console format or JSON, making it suitable for both interactive use and scripting.

## Features
- Supports Chrome, Edge, and Firefox browsers
- Lists extension details: name, version, ID, enabled status, and browser
- Outputs in console-friendly format by default or JSON with the `-json` flag
- Debug mode for troubleshooting with the `-debug` flag
- Cross-platform: works on Windows, macOS, and Linux

## Prerequisites
- [Go](https://golang.org/dl/) 1.24 or later installed
- One or more supported browsers (Chrome, Edge, Firefox) installed with extensions

## Installation
1. **Clone or Download the Repository**:
   If hosted on a Git repository (e.g., GitHub), clone it:
    
    git clone https://github.com/yourusername/go-browser-inventory.git
    cd go-browser-inventory
    
   Otherwise, download and extract the source to a directory named `go-browser-inventory`.

2. **Build the Binary**:
   From the project root:
    
    go build -o go-browser-inventory
    
   This creates an executable named `go-browser-inventory` (or `go-browser-inventory.exe` on Windows).

3. **(Optional) Move to PATH**:
   To run it from anywhere, move the binary to a directory in your PATH (e.g., `/usr/local/bin` on Unix-like systems):
    
    mv go-browser-inventory /usr/local/bin/

## Usage
Run the tool from the `go-browser-inventory` directory or anywhere if added to your PATH.

### Basic Commands
- **List all extensions (console format)**:
    
    ./go-browser-inventory
    
   Example output:
    
    Browser Extensions:
    ==================
    1. Google Wallet
       Browser: Chrome
       Version: 1.0.0.6
       ID: nmmhkkegccagdldgiimedpiccmgmieda
       Enabled: true
    ------------------
    2. uBlock Origin
       Browser: Firefox
       Version: 1.44.4
       ID: uBlock0@raymondhill.net
       Enabled: true
    ------------------
    Total extensions: 2

- **List extensions for a specific browser**:
    
    ./go-browser-inventory -browser chrome
    
   Valid browsers: `chrome`, `edge`, `firefox`.

- **Output in JSON format**:
    
    ./go-browser-inventory -json
    
   Example output:
    
    {
      "extensions": [
        {
          "name": "Google Wallet",
          "version": "1.0.0.6",
          "id": "nmmhkkegccagdldgiimedpiccmgmieda",
          "enabled": true,
          "browser": "Chrome"
        },
        {
          "name": "uBlock Origin",
          "version": "1.44.4",
          "id": "uBlock0@raymondhill.net",
          "enabled": true,
          "browser": "Firefox"
        }
      ],
      "total": 2
    }

- **Enable debug output**:
    
    ./go-browser-inventory -debug
    
   Adds detailed logs for troubleshooting (e.g., file paths scanned, message resolution steps).

- **Combine flags**:
    
    ./go-browser-inventory -browser chrome -json -debug

- **Show help**:
    
    ./go-browser-inventory -help
    
   Displays usage and examples.

### Flags
- `-browser <name>`: Filter by browser (chrome, edge, firefox). Default: all browsers.
- `-json`: Output in JSON instead of console format. Default: false.
- `-debug`: Enable debug logging. Default: false.
- `-help`: Show help information.

## Project Structure
    
    go-browser-inventory/
    ├── main.go              # Entry point and CLI logic
    ├── internal/
    │   ├── browsers/
    │   │   ├── structs.go   # Type definitions (Extension, BrowserConfig, etc.)
    │   │   ├── browsers.go  # Core inventory logic and browser configs
    │   │   ├── chromium.go  # Chrome and Edge extension handling
    │   │   └── firefox.go   # Firefox extension handling
    ├── go.mod               # Go module definition
    ├── README.md            # This file

- **`main.go`**: Handles command-line flags and output formatting.
- **`internal/browsers/`**: Contains all browser-specific logic and types, kept internal to prevent external imports.

## How It Works
- Scans default profile directories for Chrome, Edge, and Firefox.
- For Chromium-based browsers (Chrome, Edge), reads `manifest.json` files in the `Extensions` directory and resolves `__MSG_` placeholders using locale files.
- For Firefox, parses `extensions.json` in the profile directory.
- Outputs results based on the specified flags.

## Limitations
- Only supports Chrome, Edge, and Firefox.
- Assumes default profile locations; custom profiles may not be detected.
- Requires read access to browser profile directories.

## Contributing
1. Fork or clone the repository.
2. Make changes in a new branch:
    
    git checkout -b feature/your-feature

3. Test your changes:
    
    go test ./...
    go run . -browser chrome -json -debug

4. Submit a pull request or push your changes.

Feel free to open issues for bugs or feature requests!

## License
This project is unlicensed (public domain) unless specified otherwise. Use it as you see fit!