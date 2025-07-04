# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

post2post is a Go library for string transformations and utilities. It provides simple functions for common string operations with no external dependencies.

## Development Commands

### Testing
```bash
go test                    # Run all tests
go test -v                 # Run tests with verbose output
go test -cover            # Run tests with coverage
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
- `post2post.go` - Main library code with exported functions
- `post2post_test.go` - Test suite
- `go.mod` - Go module definition (github.com/pgdad/post2post)
- `README.md` - Documentation

### Key Functions
- `Transform()` - Higher-order function for string transformations
- `Reverse()` - Unicode-safe string reversal
- `ToTitle()` - Title case conversion
- `Greet()` - Simple greeting generator

## Development Guidelines

- All functions should be exported (capitalized) for library usage
- Include comprehensive tests for all functions
- Maintain Unicode safety for string operations
- Keep the library dependency-free
- Follow Go naming conventions and documentation standards