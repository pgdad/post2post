# Component Architecture - Post2Post Library

## High-Level Component View

```mermaid
graph TB
    subgraph "Post2Post Library"
        subgraph "Core Components"
            Server[Server Engine]
            ProcessorMgr[Processor Manager]
            OAuth[OAuth Integration]
            HttpClient[HTTP Client Manager]
        end
        
        subgraph "Processor Ecosystem"
            BasicProc[Basic Processors]
            AdvancedProc[Advanced Processors]
            ChainProc[Chain Processor]
            CustomProc[Custom Processors]
        end
        
        subgraph "Network Layer"
            HttpHandler[HTTP Handlers]
            TailscaleClient[Tailscale Client]
            RegularHttp[Regular HTTP]
        end
        
        subgraph "Configuration"
            EnvConfig[Environment Config]
            FluentAPI[Fluent Builder API]
        end
    end
    
    subgraph "External Dependencies"
        TailscaleAPI[Tailscale API]
        OAuthProvider[OAuth Provider]
        TSNet[Tailscale tsnet]
    end
    
    subgraph "Client Applications"
        Examples[Example Applications]
        UserCode[User Applications]
    end
    
    %% Core relationships
    Server --> ProcessorMgr
    Server --> HttpClient
    Server --> OAuth
    ProcessorMgr --> BasicProc
    ProcessorMgr --> AdvancedProc
    ProcessorMgr --> ChainProc
    HttpClient --> TailscaleClient
    HttpClient --> RegularHttp
    TailscaleClient --> TSNet
    
    %% Configuration relationships
    FluentAPI --> Server
    EnvConfig --> OAuth
    EnvConfig --> TailscaleClient
    
    %% External connections
    OAuth --> OAuthProvider
    OAuth --> TailscaleAPI
    TailscaleClient --> TSNet
    
    %% Client connections
    Examples --> Server
    UserCode --> Server
    UserCode --> CustomProc
    
    %% Styling
    classDef coreComponent fill:#e1f5fe
    classDef processor fill:#f3e5f5
    classDef network fill:#e8f5e8
    classDef config fill:#fff3e0
    classDef external fill:#ffebee
    classDef client fill:#f1f8e9
    
    class Server,ProcessorMgr,OAuth,HttpClient coreComponent
    class BasicProc,AdvancedProc,ChainProc,CustomProc processor
    class HttpHandler,TailscaleClient,RegularHttp network
    class EnvConfig,FluentAPI config
    class TailscaleAPI,OAuthProvider,TSNet external
    class Examples,UserCode client
```

## Detailed Component Breakdown

### 1. **Server Engine**
**Purpose**: Central orchestrator for HTTP server operations and request lifecycle management

**Responsibilities**:
- HTTP server lifecycle (start/stop)
- Request routing and handling
- Round-trip communication coordination
- Response channel management
- Thread-safe operations

**Key Interfaces**:
```go
type Server struct {
    // Network configuration
    network, iface string
    port int
    
    // HTTP components
    listener net.Listener
    server *http.Server
    client *http.Client
    
    // Round-trip management
    roundTripChans map[string]chan *RoundTripResponse
    
    // Processing
    processor PayloadProcessor
}
```

### 2. **Processor Manager**
**Purpose**: Manages payload processing through configurable processor interfaces

**Components**:
- **Basic Processors**: Simple transformation (Echo, HelloWorld, Timestamp)
- **Advanced Processors**: Context-aware processing with metadata
- **Chain Processor**: Sequential processing pipeline
- **Custom Processors**: User-defined processing logic

**Processing Pipeline**:
```
Request → Processor Selection → Processing → Response Generation → Callback
```

### 3. **OAuth Integration**
**Purpose**: Seamless Tailscale auth key generation using OAuth credentials

**Features**:
- Automatic OAuth token acquisition
- Tailscale API integration
- Ephemeral key generation
- Environment-based configuration

**Flow**:
```
OAuth Credentials → Token Request → Tailscale API → Auth Key → Client Usage
```

### 4. **Network Layer**
**Purpose**: Dual-mode networking with Tailscale and regular HTTP support

**Components**:
- **Tailscale Client**: Secure mesh networking via tsnet
- **Regular HTTP**: Standard HTTP client for fallback
- **HTTP Handlers**: Request routing and response management

**Selection Logic**:
```
If (tailnetKey provided) → Use Tailscale
Else → Use Regular HTTP
If (Tailscale fails) → Fallback to Regular HTTP
```

## Component Interactions

### **Configuration Flow**
```mermaid
graph LR
    EnvVars[Environment Variables] --> Config[Configuration Loading]
    FluentAPI[Fluent API Calls] --> Config
    Config --> ServerSetup[Server Setup]
    Config --> ProcessorSetup[Processor Setup]
    Config --> NetworkSetup[Network Setup]
```

### **Request Processing Flow**
```mermaid
graph TD
    Request[HTTP Request] --> Router[Request Router]
    Router --> WebhookHandler[Webhook Handler]
    Router --> RoundTripHandler[RoundTrip Handler]
    
    WebhookHandler --> ProcessorMgr[Processor Manager]
    ProcessorMgr --> Processor[Selected Processor]
    Processor --> ResponseGen[Response Generation]
    ResponseGen --> Callback[Callback Delivery]
    
    RoundTripHandler --> ChannelMgr[Channel Manager]
    ChannelMgr --> ResponseDelivery[Response Delivery]
```

### **Network Selection Flow**
```mermaid
graph TD
    NetworkRequest[Network Request] --> HasTailnetKey{Has TailnetKey?}
    HasTailnetKey -->|Yes| TailscaleAttempt[Attempt Tailscale]
    HasTailnetKey -->|No| RegularHTTP[Use Regular HTTP]
    
    TailscaleAttempt --> TailscaleSuccess{Success?}
    TailscaleSuccess -->|Yes| TailscaleDelivery[Deliver via Tailscale]
    TailscaleSuccess -->|No| FallbackHTTP[Fallback to HTTP]
    
    TailscaleDelivery --> Complete[Request Complete]
    FallbackHTTP --> Complete
    RegularHTTP --> Complete
```

## Deployment Architecture

### **Standalone Application**
```mermaid
graph TB
    subgraph "Application Process"
        App[User Application]
        Post2Post[Post2Post Library]
        App --> Post2Post
    end
    
    subgraph "Network Layer"
        HTTP[HTTP Interface]
        Tailscale[Tailscale Interface]
    end
    
    Post2Post --> HTTP
    Post2Post --> Tailscale
    
    HTTP --> Internet[Internet]
    Tailscale --> TailscaleNetwork[Tailscale Mesh]
```

### **Microservices Architecture**
```mermaid
graph TB
    subgraph "Service A"
        AppA[Application A]
        Post2PostA[Post2Post Instance A]
        AppA --> Post2PostA
    end
    
    subgraph "Service B"
        AppB[Application B]
        Post2PostB[Post2Post Instance B]
        AppB --> Post2PostB
    end
    
    subgraph "Service C"
        AppC[Application C]
        Post2PostC[Post2Post Instance C]
        AppC --> Post2PostC
    end
    
    subgraph "Tailscale Mesh Network"
        TailscaleNet[Secure Overlay Network]
    end
    
    Post2PostA -.-> TailscaleNet
    Post2PostB -.-> TailscaleNet
    Post2PostC -.-> TailscaleNet
    
    Post2PostA --> Post2PostB
    Post2PostB --> Post2PostC
    Post2PostC --> Post2PostA
```

## Extension Points

### **Custom Processors**
```go
type MyCustomProcessor struct {}

func (p *MyCustomProcessor) Process(payload interface{}, requestID string) (interface{}, error) {
    // Custom processing logic
    return processedPayload, nil
}

// Usage
server := post2post.NewServer().WithProcessor(&MyCustomProcessor{})
```

### **Network Extensions**
- Custom HTTP clients
- Additional security layers
- Protocol adapters
- Monitoring integrations

### **Configuration Extensions**
- Configuration file support
- Dynamic reconfiguration
- Service discovery integration
- Health check endpoints

This component architecture provides a flexible, extensible foundation for HTTP communication with optional security enhancements through Tailscale integration.