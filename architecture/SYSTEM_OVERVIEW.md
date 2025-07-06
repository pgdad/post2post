# Post2Post System Architecture Overview

## System Purpose
Post2Post is a Go library that enables bidirectional HTTP communication with configurable payload processing and optional Tailscale networking integration. It supports round-trip messaging patterns where a client posts data to a receiver and waits for a response back to a callback URL.

## Core Components

### 1. **Server (`Server` struct)**
- **Purpose**: HTTP server that handles incoming requests and manages round-trip communications
- **Key Features**: 
  - Configurable network interfaces and ports
  - Round-trip messaging with timeout handling
  - Webhook endpoint with payload processing
  - Optional Tailscale integration

### 2. **Payload Processing System**
- **Purpose**: Extensible system for processing incoming webhook payloads
- **Components**:
  - `PayloadProcessor` interface for basic processing
  - `AdvancedPayloadProcessor` interface for context-aware processing
  - Built-in processors (Echo, Transform, Validator, etc.)
  - Chain processor for combining multiple processors

### 3. **OAuth Integration**
- **Purpose**: Automatic Tailscale auth key generation using OAuth credentials
- **Features**:
  - Seamless integration with Tailscale API
  - Ephemeral key generation
  - Environment-based configuration

### 4. **Tailscale Networking**
- **Purpose**: Secure peer-to-peer communication using Tailscale mesh networking
- **Integration**: Optional enhancement for secure communication channels

## Architecture Patterns

### **Producer-Consumer Pattern**
- Clients produce HTTP requests with payloads
- Receivers consume and process payloads
- Responses flow back through callback URLs

### **Plugin Architecture**
- Configurable payload processors
- Extensible through interface implementation
- Chain of responsibility for complex processing

### **Environment-Based Configuration**
- OAuth credentials from environment variables
- Network interface configuration
- Tag-based Tailscale device management

## Key Design Principles

### 1. **Flexibility**
- Configurable server parameters (network, interface, ports)
- Pluggable payload processing
- Optional Tailscale integration

### 2. **Security**
- Ephemeral Tailscale key generation
- OAuth-based authentication
- Secure networking through Tailscale mesh

### 3. **Reliability**
- Timeout handling for round-trip operations
- Graceful fallback (Tailscale → regular HTTP)
- Thread-safe operations with mutex protection

### 4. **Ease of Use**
- Fluent API with method chaining
- Automatic configuration detection
- Comprehensive error handling and documentation

## Integration Points

### **External Dependencies**
- Tailscale API for mesh networking
- OAuth2 for authentication
- Standard Go HTTP library

### **Environment Integration**
- Environment variables for configuration
- CI/CD pipeline compatibility
- Container-friendly design

## Scalability Considerations

### **Horizontal Scaling**
- Stateless server design
- Independent payload processors
- Ephemeral Tailscale devices

### **Performance**
- Concurrent request handling
- Configurable timeouts
- Efficient HTTP client pooling

This architecture enables secure, flexible, and scalable HTTP communication patterns with optional mesh networking capabilities.