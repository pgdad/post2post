# Sequence Diagrams - Post2Post Library

## 1. Round-Trip Post Sequence

```mermaid
sequenceDiagram
    participant Client
    participant Server
    participant Receiver
    participant Channel as "Response Channel"
    
    Note over Client, Channel: Round-trip HTTP communication with response callback
    
    Client->>Server: Start()
    Server->>Server: Listen on configured port
    
    Client->>Server: RoundTripPost(payload, tailnetKey)
    Server->>Server: Generate requestID
    Server->>Server: Create response channel
    
    Server->>Receiver: POST /webhook<br/>{ url, payload, requestID, tailnetKey }
    Receiver-->>Server: 200 OK (acknowledgment)
    
    Note over Receiver: Process payload asynchronously
    
    Receiver->>Server: POST /roundtrip<br/>{ requestID, processedPayload }
    Server->>Channel: Send response to waiting channel
    Server-->>Receiver: 200 OK
    
    Channel->>Server: Response received
    Server->>Client: Return RoundTripResponse
    
    Note over Client, Server: Cleanup response channel
```

## 2. Webhook Processing with Payload Processors

```mermaid
sequenceDiagram
    participant Client
    participant Server
    participant Processor as "PayloadProcessor"
    participant Receiver as "External System"
    
    Note over Client, Receiver: Webhook processing with configurable payload transformation
    
    Client->>Server: Configure processor
    Server->>Server: Store processor reference
    
    Client->>Server: POST /webhook<br/>{ payload, requestID, callbackURL }
    Server-->>Client: 200 OK (immediate response)
    
    par Async Processing
        Server->>Processor: Process(payload, requestID)
        
        alt Basic Processor
            Processor->>Processor: Transform payload
            Processor->>Server: Return processed payload
        else Advanced Processor
            Processor->>Processor: ProcessWithContext(payload, context)
            Note over Processor: Context includes requestID, URL, timestamp
            Processor->>Server: Return processed payload
        end
        
        Server->>Receiver: POST callback URL<br/>{ requestID, processedPayload }
        Receiver-->>Server: Response (optional)
    end
    
    Note over Server, Receiver: Callback happens asynchronously after acknowledgment
```

## 3. OAuth Auth Key Generation

```mermaid
sequenceDiagram
    participant Client
    participant Server
    participant OAuth as "OAuth Provider"
    participant TailscaleAPI as "Tailscale API"
    
    Note over Client, TailscaleAPI: Automatic Tailscale auth key generation using OAuth
    
    Client->>Server: GenerateTailnetKeyFromOAuth(reusable, ephemeral, preauth, tags)
    
    Server->>Server: Set I_Acknowledge_This_API_Is_Unstable = true
    Server->>Server: Read TS_API_CLIENT_ID, TS_API_CLIENT_SECRET
    
    Server->>OAuth: POST /oauth/token<br/>grant_type=client_credentials
    OAuth->>Server: Access token
    
    Server->>TailscaleAPI: POST /tailnet/-/keys<br/>Authorization: Bearer <token><br/>{ capabilities: { devices: { create: {...} } } }
    TailscaleAPI->>Server: Auth key response
    
    Server->>Client: Return auth key (tskey-auth-...)
    
    Note over Client, Server: Key can now be used for Tailscale networking
```

## 4. Tailscale-Enhanced Communication

```mermaid
sequenceDiagram
    participant Client
    participant ClientServer as "Client Server"
    participant TailscaleNet as "Tailscale Network"
    participant ReceiverServer as "Receiver Server"
    participant Receiver
    
    Note over Client, Receiver: Secure communication through Tailscale mesh network
    
    Client->>ClientServer: Start() with Tailscale interface
    Client->>ReceiverServer: Start() with Tailscale interface
    
    Client->>ClientServer: PostJSONWithTailnet(payload, authKey)
    ClientServer->>ClientServer: createTailscaleClient(authKey)
    
    ClientServer->>TailscaleNet: Establish tsnet connection
    TailscaleNet-->>ClientServer: Secure tunnel ready
    
    ClientServer->>TailscaleNet: POST via Tailscale tunnel
    TailscaleNet->>ReceiverServer: Forward encrypted request
    ReceiverServer->>Receiver: Process webhook
    
    Receiver->>ReceiverServer: Generate response
    ReceiverServer->>TailscaleNet: POST response via Tailscale
    TailscaleNet->>ClientServer: Forward encrypted response
    ClientServer->>Client: Delivery confirmation
    
    Note over ClientServer, ReceiverServer: All communication encrypted end-to-end
```

## 5. Error Handling and Fallback

```mermaid
sequenceDiagram
    participant Client
    participant Server
    participant Tailscale as "Tailscale Service"
    participant HTTP as "Regular HTTP"
    participant Receiver
    
    Note over Client, Receiver: Error handling with graceful fallback
    
    Client->>Server: PostJSONWithTailnet(payload, authKey)
    
    alt Tailscale Available
        Server->>Tailscale: createTailscaleClient(authKey)
        Tailscale-->>Server: HTTP client configured
        Server->>Receiver: POST via Tailscale
        Receiver-->>Server: Response
        Server->>Client: Success
    else Tailscale Fails
        Server->>Tailscale: createTailscaleClient(authKey)
        Tailscale-->>Server: Error (network/auth issue)
        
        Note over Server: Log error and fallback
        Server->>HTTP: Use regular HTTP client
        Server->>Receiver: POST via regular HTTP
        Receiver-->>Server: Response
        Server->>Client: Success (with fallback notice)
    end
    
    Note over Client, Receiver: Graceful degradation ensures reliability
```

## 6. Chain Processor Sequence

```mermaid
sequenceDiagram
    participant Client
    participant Server
    participant ChainProcessor
    participant Processor1 as "Validator"
    participant Processor2 as "Transform"
    participant Processor3 as "Counter"
    
    Note over Client, Processor3: Sequential processing through processor chain
    
    Client->>Server: Configure ChainProcessor([Validator, Transform, Counter])
    Server->>ChainProcessor: Store processor chain
    
    Client->>Server: POST /webhook { payload }
    Server->>ChainProcessor: Process(payload, requestID)
    
    ChainProcessor->>Processor1: Process(payload, requestID)
    Processor1->>Processor1: Validate payload structure
    
    alt Validation Success
        Processor1->>ChainProcessor: Return validated payload
        ChainProcessor->>Processor2: Process(validatedPayload, requestID)
        Processor2->>Processor2: Transform data format
        Processor2->>ChainProcessor: Return transformed payload
        
        ChainProcessor->>Processor3: Process(transformedPayload, requestID)
        Processor3->>Processor3: Increment counter, add metadata
        Processor3->>ChainProcessor: Return final payload
        
        ChainProcessor->>Server: Return processed payload
        Server->>Client: Success response
    else Validation Failure
        Processor1->>ChainProcessor: Return error
        ChainProcessor->>Server: Return error
        Server->>Client: 500 Internal Server Error
    end
    
    Note over Client, Processor3: Chain stops at first error
```

These sequence diagrams illustrate the key interaction patterns in the post2post library, showing how components collaborate to provide flexible HTTP communication with optional security enhancements.