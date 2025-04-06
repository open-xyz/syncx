# SyncX

SyncX is an open source tool that simplifies the self-hosting of open source projects. It provides features like automated vulnerability scanning, built-in load balancing, and an intuitive web interface.

## Features

- **Project Hosting**: Host and serve open source projects
- **Automated Vulnerability Scanning**: Scan projects for security issues using Trivy
- **Load Balancing**: Distribute traffic across multiple instances
- **Intuitive Interface**: Web-based dashboard to manage projects

## Requirements

- Go 1.16 or higher
- Trivy (for vulnerability scanning)

## Installation

### From Source

1. Clone this repository:

   ```bash
   git clone https://github.com/open-xyz/syncx.git
   cd syncx
   ```

2. Install dependencies:

   ```bash
   go mod download
   ```

3. Build the application:

   ```bash
   go build -o syncx cmd/main.go
   ```

4. Install Trivy (optional, required for vulnerability scanning):

   ```bash
   # For Arch Linux/Manjaro
   sudo pacman -S trivy

   # For Ubuntu/Debian
   sudo apt-get install trivy

   # For macOS
   brew install trivy
   ```

### Using Binary Releases

Coming soon.

## Configuration

SyncX uses a YAML configuration file located at `config/syncx.yaml`. You can specify a different location using the `-config` flag when starting the application.

Example configuration:

```yaml
server:
  port: 8080
  host: "0.0.0.0"

projects:
  directory: "./projects"

database:
  path: "./syncx.db"

balancer:
  endpoints:
    - "http://localhost:8081"
    - "http://localhost:8082"

scanning:
  auto_scan_on_add: true
```

## Usage

1. Start the SyncX server:

   ```bash
   ./syncx
   ```

2. Access the web dashboard at `http://localhost:8080`

3. Add a project by providing a name and Git repository URL

4. View, scan, and manage your projects through the dashboard

## Development

To run the application in development mode:

```bash
go run cmd/main.go
```

## License

MIT License

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.
