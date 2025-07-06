# Class Diagram - Post2Post Library

```mermaid
classDiagram
    class Server {
        -network string
        -iface string
        -port int
        -listener net.Listener
        -server http.Server
        -mu sync.RWMutex
        -running bool
        -postURL string
        -client http.Client
        -roundTripChans map
        -defaultTimeout time.Duration
        -processor PayloadProcessor
        
        +NewServer() Server
        +WithNetwork(network string) Server
        +WithInterface(iface string) Server
        +WithPostURL(url string) Server
        +WithTimeout(timeout time.Duration) Server
        +WithProcessor(processor PayloadProcessor) Server
        +Start() error
        +Stop() error
        +GetPort() int
        +GetInterface() string
        +IsRunning() bool
        +GetURL() string
        +PostJSON(payload interface) error
        +PostJSONWithTailnet(payload interface, tailnetKey string) error
        +RoundTripPost(payload interface, tailnetKey string) RoundTripResponse
        +RoundTripPostWithTimeout(payload interface, tailnetKey string, timeout time.Duration) RoundTripResponse
        +GenerateTailnetKeyFromOAuth(reusable bool, ephemeral bool, preauth bool, tags string) string
        -createTailscaleClient(tailnetKey string) http.Client
        -postWithOptionalTailscale(url string, data byte, tailnetKey string) http.Response
        -roundTripHandler(w http.ResponseWriter, r http.Request)
        -webhookHandler(w http.ResponseWriter, r http.Request)
        -postProcessedResponse(callbackURL string, requestID string, payload interface, tailnetKey string)
        -defaultHandler(w http.ResponseWriter, r http.Request)
    }

    class PayloadProcessor {
        <<interface>>
        +Process(payload interface, requestID string) interface
    }

    class AdvancedPayloadProcessor {
        <<interface>>
        +ProcessWithContext(payload interface, context ProcessorContext) interface
    }

    class ProcessorContext {
        +RequestID string
        +URL string
        +TailnetKey string
        +ReceivedAt time.Time
    }

    class PostData {
        +URL string
        +Payload interface
        +RequestID string
        +TailnetKey string
    }

    class RoundTripResponse {
        +Payload interface
        +Success bool
        +Error string
        +Timeout bool
        +RequestID string
    }

    class HelloWorldProcessor {
        +Process(payload interface, requestID string) interface
    }

    class EchoProcessor {
        +Process(payload interface, requestID string) interface
    }

    class TimestampProcessor {
        +Process(payload interface, requestID string) interface
    }

    class CounterProcessor {
        -count int64
        -mu sync.Mutex
        +Process(payload interface, requestID string) interface
        +GetCount() int64
        +Reset()
    }

    class TransformProcessor {
        -transformFunc function
        +NewTransformProcessor(function) TransformProcessor
        +Process(payload interface, requestID string) interface
    }

    class ValidatorProcessor {
        -validateFunc function
        +NewValidatorProcessor(function) ValidatorProcessor
        +Process(payload interface, requestID string) interface
    }

    class AdvancedContextProcessor {
        -processFunc function
        +NewAdvancedContextProcessor(function) AdvancedContextProcessor
        +ProcessWithContext(payload interface, context ProcessorContext) interface
    }

    class ChainProcessor {
        -processors PayloadProcessor[]
        +NewChainProcessor(processors PayloadProcessor) ChainProcessor
        +Process(payload interface, requestID string) interface
    }

    %% Relationships
    Server --> PayloadProcessor : uses
    Server --> PostData : creates
    Server --> RoundTripResponse : returns
    Server --> ProcessorContext : creates
    
    PayloadProcessor <|.. HelloWorldProcessor : implements
    PayloadProcessor <|.. EchoProcessor : implements
    PayloadProcessor <|.. TimestampProcessor : implements
    PayloadProcessor <|.. CounterProcessor : implements
    PayloadProcessor <|.. TransformProcessor : implements
    PayloadProcessor <|.. ValidatorProcessor : implements
    PayloadProcessor <|.. ChainProcessor : implements
    
    AdvancedPayloadProcessor <|.. AdvancedContextProcessor : implements
    
    AdvancedContextProcessor --> ProcessorContext : uses
    
    ChainProcessor --> PayloadProcessor : contains multiple
```

## Key Relationships

### **Composition**
- `Server` contains configuration, HTTP components, and channels
- `ChainProcessor` contains multiple `PayloadProcessor` instances

### **Interface Implementation**
- Multiple built-in processors implement `PayloadProcessor`
- `AdvancedContextProcessor` implements `AdvancedPayloadProcessor`

### **Dependency**
- `Server` depends on `PayloadProcessor` for request processing
- Processors use `ProcessorContext` for advanced processing

### **Data Flow**
- `PostData` flows into the server
- Processors transform data
- `RoundTripResponse` flows back to clients

## Design Patterns Used

### **Strategy Pattern**
- `PayloadProcessor` interface allows different processing strategies
- Processors can be swapped at runtime

### **Chain of Responsibility**
- `ChainProcessor` implements sequential processing
- Each processor can modify or validate data

### **Builder Pattern**
- `Server` configuration uses fluent interface
- Method chaining for easy setup: `NewServer().WithPort(8080).WithProcessor(processor)`

### **Factory Pattern**
- Constructor functions for processors (e.g., `NewTransformProcessor`)
- `NewServer()` factory method