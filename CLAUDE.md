# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

post2post is a Go library for starting and managing local web servers with configurable network options. It provides a simple API to create HTTP servers that can listen on TCP4 or TCP6 networks with customizable interfaces.

## Development Commands

### Testing
```bash
go test                    # Run all tests
go test -v                 # Run tests with verbose output
go test -cover            # Run tests with coverage
go test -race             # Run tests with race detection
```

### Building
```bash
go build                  # Build the library
go mod tidy              # Clean up dependencies
```

### Code Quality
```bash
go fmt ./...             # Format code
go vet ./...             # Run static analysis
```

## Architecture

This is a simple Go library with a flat structure:
- `post2post.go` - Main library code with Server struct and methods
- `post2post_test.go` - Comprehensive test suite
- `go.mod` - Go module definition (github.com/pgdad/post2post)
- `README.md` - Documentation

### Key Components

#### Server struct
- Thread-safe server management with mutex protection
- Configurable network type (tcp4/tcp6) and interface
- Automatic port assignment when using port 0
- Built-in HTTP handler for basic server information

#### Core Methods
- `NewServer()` - Creates server with defaults (TCP4, all interfaces, random port)
- `WithNetwork()` - Sets network type (tcp4 or tcp6)
- `WithInterface()` - Sets listening interface
- `Start()` - Starts the server and assigns port
- `Stop()` - Stops the server and cleans up resources
- `GetPort()` - Returns assigned port
- `GetInterface()` - Returns listening interface ("localhost" if unspecified)
- `IsRunning()` - Returns server status
- `GetNetwork()` - Returns network type

### Default Behavior
- Network: tcp4
- Interface: "" (all interfaces, returns "localhost" via GetInterface())
- Port: 0 (randomly assigned by OS)
- Handler: Simple HTTP handler returning server information

## Development Guidelines

- All public methods should be thread-safe using mutex protection
- Server state should be properly managed (running/stopped)
- Include comprehensive tests including edge cases and concurrent access
- Keep the library dependency-free (uses only standard library)
- Follow Go naming conventions and documentation standards
- Handle errors appropriately and provide meaningful error messages