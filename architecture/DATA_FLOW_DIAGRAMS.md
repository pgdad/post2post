# Data Flow Diagrams - Post2Post Library

## 1. Overall System Data Flow

```mermaid
flowchart TD
    subgraph "Client Side"
        ClientApp[Client Application]
        ClientServer[Client Post2Post Server]
        ClientEnv[Environment Config]
    end
    
    subgraph "Network Layer"
        Internet[Internet/HTTP]
        TailscaleNet[Tailscale Mesh Network]
    end
    
    subgraph "Receiver Side"
        ReceiverServer[Receiver Post2Post Server]
        ProcessorEngine[Processor Engine]
        ReceiverApp[Receiver Application]
    end
    
    subgraph "External Services"
        OAuth[OAuth Provider]
        TailscaleAPI[Tailscale API]
    end
    
    %% Configuration Flow
    ClientEnv --> ClientServer
    
    %% Auth Key Generation Flow
    ClientServer --> OAuth
    OAuth --> TailscaleAPI
    TailscaleAPI --> ClientServer
    
    %% Request Flow
    ClientApp --> ClientServer
    ClientServer --> Internet
    ClientServer --> TailscaleNet
    Internet --> ReceiverServer
    TailscaleNet --> ReceiverServer
    
    %% Processing Flow
    ReceiverServer --> ProcessorEngine
    ProcessorEngine --> ReceiverApp
    ReceiverApp --> ProcessorEngine
    ProcessorEngine --> ReceiverServer
    
    %% Response Flow
    ReceiverServer --> Internet
    ReceiverServer --> TailscaleNet
    Internet --> ClientServer
    TailscaleNet --> ClientServer
    ClientServer --> ClientApp
    
    %% Styling
    classDef client fill:#e3f2fd
    classDef network fill:#e8f5e8
    classDef receiver fill:#fff3e0
    classDef external fill:#ffebee
    
    class ClientApp,ClientServer,ClientEnv client
    class Internet,TailscaleNet network
    class ReceiverServer,ProcessorEngine,ReceiverApp receiver
    class OAuth,TailscaleAPI external
```

## 2. Round-Trip Data Flow

```mermaid
flowchart TD
    Start([Client Initiates Round-Trip]) --> PreparePayload[Prepare Payload Data]
    
    PreparePayload --> GenerateRequestID[Generate Unique Request ID]
    GenerateRequestID --> CreateChannel[Create Response Channel]
    
    CreateChannel --> CheckTailscale{Has Tailscale Key?}
    
    CheckTailscale -->|Yes| CreateTSClient[Create Tailscale Client]
    CheckTailscale -->|No| UseRegularHTTP[Use Regular HTTP Client]
    
    CreateTSClient --> SendRequest[Send POST Request]
    UseRegularHTTP --> SendRequest
    
    SendRequest --> ReceiverReceives[Receiver Gets Request]
    
    ReceiverReceives --> ExtractData[Extract URL, Payload, Request ID]
    ExtractData --> ProcessPayload[Process Payload via Configured Processor]
    
    ProcessPayload --> GenerateResponse[Generate Response Payload]
    GenerateResponse --> SendCallback[Send Response to Callback URL]
    
    SendCallback --> ClientReceivesResponse[Client Receives Response on /roundtrip]
    ClientReceivesResponse --> MatchRequestID[Match Request ID to Waiting Channel]
    
    MatchRequestID --> DeliverToChannel[Deliver Response to Channel]
    DeliverToChannel --> ClientGetsResult[Client Gets RoundTripResponse]
    
    ClientGetsResult --> CleanupChannel[Cleanup Response Channel]
    CleanupChannel --> End([Round-Trip Complete])
    
    %% Timeout Path
    CreateChannel --> StartTimeout[Start Timeout Timer]
    StartTimeout --> TimeoutCheck{Timeout Reached?}
    TimeoutCheck -->|Yes| TimeoutResponse[Return Timeout Response]
    TimeoutCheck -->|No| DeliverToChannel
    TimeoutResponse --> CleanupChannel
```

## 3. Payload Processing Data Flow

```mermaid
flowchart TD
    IncomingRequest[Incoming HTTP Request] --> ParseJSON[Parse JSON Body]
    ParseJSON --> ValidateStructure[Validate Request Structure]
    
    ValidateStructure --> ExtractComponents[Extract Components]
    ExtractComponents --> RequestID[Request ID]
    ExtractComponents --> PayloadData[Payload Data]
    ExtractComponents --> CallbackURL[Callback URL]
    ExtractComponents --> TailnetKey[Tailscale Key]
    
    PayloadData --> ProcessorSelection{Processor Type?}
    
    ProcessorSelection -->|Basic| BasicProcessor[Basic Processor.Process]
    ProcessorSelection -->|Advanced| AdvancedProcessor[Advanced Processor.ProcessWithContext]
    ProcessorSelection -->|Chain| ChainProcessor[Chain Processor]
    ProcessorSelection -->|None| DefaultEcho[Default Echo Processing]
    
    %% Basic Processing Path
    BasicProcessor --> ProcessorLogic1[Apply Transformation Logic]
    ProcessorLogic1 --> BasicResult[Return Processed Payload]
    
    %% Advanced Processing Path
    AdvancedProcessor --> CreateContext[Create ProcessorContext]
    CreateContext --> ContextData[RequestID, URL, TailnetKey, ReceivedAt]
    ContextData --> ProcessorLogic2[Apply Context-Aware Logic]
    ProcessorLogic2 --> AdvancedResult[Return Processed Payload]
    
    %% Chain Processing Path
    ChainProcessor --> FirstProcessor[First Processor in Chain]
    FirstProcessor --> ChainValidation{Processing Success?}
    ChainValidation -->|Yes| NextProcessor[Next Processor]
    ChainValidation -->|No| ChainError[Return Error]
    NextProcessor --> LastProcessor[Last Processor]
    LastProcessor --> ChainResult[Return Final Processed Payload]
    
    %% Default Processing Path
    DefaultEcho --> EchoResult[Return Original Payload]
    
    %% Convergence
    BasicResult --> PrepareResponse[Prepare Response Data]
    AdvancedResult --> PrepareResponse
    ChainResult --> PrepareResponse
    EchoResult --> PrepareResponse
    ChainError --> ErrorResponse[Prepare Error Response]
    
    PrepareResponse --> ResponseData[Response JSON]
    ErrorResponse --> ResponseData
    
    ResponseData --> DeliveryMethod{Delivery Method?}
    
    DeliveryMethod -->|Callback URL| SendCallback[Send to Callback URL]
    DeliveryMethod -->|Round-trip| SendToChannel[Send to Response Channel]
    
    SendCallback --> NetworkDelivery[Network Delivery]
    SendToChannel --> ChannelDelivery[Channel Delivery]
    
    NetworkDelivery --> Complete[Processing Complete]
    ChannelDelivery --> Complete
```

## 4. OAuth Auth Key Generation Data Flow

```mermaid
flowchart TD
    TriggerOAuth[Trigger OAuth Key Generation] --> ReadEnvVars[Read Environment Variables]
    
    ReadEnvVars --> ClientID[TS_API_CLIENT_ID]
    ReadEnvVars --> ClientSecret[TS_API_CLIENT_SECRET]
    ReadEnvVars --> Tags[TAILSCALE_TAGS]
    
    ClientID --> ValidateCredentials{Credentials Valid?}
    ClientSecret --> ValidateCredentials
    
    ValidateCredentials -->|No| CredentialError[Return Credential Error]
    ValidateCredentials -->|Yes| SetAPIFlag[Set I_Acknowledge_This_API_Is_Unstable = true]
    
    SetAPIFlag --> CreateOAuthConfig[Create OAuth2 Config]
    CreateOAuthConfig --> OAuthCredentials[Client Credentials Config]
    
    OAuthCredentials --> RequestToken[Request OAuth Token]
    RequestToken --> OAuthProvider[POST to /oauth/token]
    
    OAuthProvider --> TokenResponse{Token Success?}
    TokenResponse -->|No| OAuthError[Return OAuth Error]
    TokenResponse -->|Yes| AccessToken[Extract Access Token]
    
    AccessToken --> CreateTSClient[Create Tailscale Client]
    CreateTSClient --> SetHTTPClient[Set OAuth HTTP Client]
    SetHTTPClient --> SetBaseURL[Set Tailscale API Base URL]
    
    SetBaseURL --> PrepareKeyRequest[Prepare Key Creation Request]
    PrepareKeyRequest --> KeyCapabilities[Define Key Capabilities]
    
    KeyCapabilities --> Reusable[Reusable Flag]
    KeyCapabilities --> Ephemeral[Ephemeral Flag]
    KeyCapabilities --> Preauth[Preauthorized Flag]
    KeyCapabilities --> TagList[Tag List]
    
    TagList --> SendKeyRequest[Send Key Creation Request]
    SendKeyRequest --> TailscaleAPI[POST to /tailnet/-/keys]
    
    TailscaleAPI --> KeyResponse{Key Creation Success?}
    KeyResponse -->|No| TailscaleError[Return Tailscale API Error]
    KeyResponse -->|Yes| AuthKey[Extract Auth Key]
    
    AuthKey --> LogKeyGeneration[Log Key Generation]
    LogKeyGeneration --> ReturnKey[Return tskey-auth-... Key]
    
    %% Error Convergence
    CredentialError --> ErrorEnd[Error Response]
    OAuthError --> ErrorEnd
    TailscaleError --> ErrorEnd
    
    %% Success Convergence
    ReturnKey --> SuccessEnd[Success Response]
```

## 5. Network Selection and Fallback Data Flow

```mermaid
flowchart TD
    NetworkRequest[Initiate Network Request] --> CheckTailnetKey{Has Tailscale Key?}
    
    CheckTailnetKey -->|No| UseRegularHTTP[Configure Regular HTTP Client]
    CheckTailnetKey -->|Yes| AttemptTailscale[Attempt Tailscale Connection]
    
    AttemptTailscale --> CreateTSNetServer[Create tsnet.Server]
    CreateTSNetServer --> ConfigureTSServer[Configure Tailscale Server]
    
    ConfigureTSServer --> Hostname[Set Hostname]
    ConfigureTSServer --> AuthKey[Set Auth Key]
    ConfigureTSServer --> EphemeralFlag[Set Ephemeral Flag]
    ConfigureTSServer --> LogFunction[Set Log Function]
    
    LogFunction --> StartTSServer[Start Tailscale Server]
    StartTSServer --> TSStartResult{Server Start Success?}
    
    TSStartResult -->|No| TSStartError[Tailscale Start Error]
    TSStartResult -->|Yes| CreateHTTPClient[Create HTTP Client from tsnet]
    
    CreateHTTPClient --> TSClientReady[Tailscale Client Ready]
    
    %% Fallback Logic
    TSStartError --> LogFallback[Log Fallback Message]
    LogFallback --> UseRegularHTTP
    
    %% HTTP Client Configuration
    UseRegularHTTP --> SetTimeout[Set HTTP Timeout]
    SetTimeout --> RegularClientReady[Regular HTTP Client Ready]
    
    %% Request Execution
    TSClientReady --> ExecuteRequest[Execute HTTP Request]
    RegularClientReady --> ExecuteRequest
    
    ExecuteRequest --> PrepareRequest[Prepare HTTP Request]
    PrepareRequest --> SetHeaders[Set Request Headers]
    SetHeaders --> SetContentType[Set Content-Type: application/json]
    SetContentType --> SetUserAgent[Set User-Agent]
    
    SetUserAgent --> SendRequest[Send HTTP Request]
    SendRequest --> RequestResult{Request Success?}
    
    RequestResult -->|No| NetworkError[Network Error]
    RequestResult -->|Yes| ProcessResponse[Process HTTP Response]
    
    ProcessResponse --> CheckStatusCode{Status Code OK?}
    CheckStatusCode -->|No| HTTPError[HTTP Error Response]
    CheckStatusCode -->|Yes| ReadResponseBody[Read Response Body]
    
    ReadResponseBody --> CloseResponse[Close Response Body]
    CloseResponse --> RequestSuccess[Request Success]
    
    %% Error Handling
    NetworkError --> ErrorLogging[Log Network Error]
    HTTPError --> ErrorLogging
    ErrorLogging --> ReturnError[Return Error to Caller]
    
    %% Success Path
    RequestSuccess --> ResponseLogging[Log Success]
    ResponseLogging --> ReturnSuccess[Return Success to Caller]
```

## 6. Configuration Data Flow

```mermaid
flowchart TD
    AppStart[Application Startup] --> LoadConfig[Load Configuration]
    
    LoadConfig --> EnvVars[Environment Variables]
    LoadConfig --> FluentAPI[Fluent API Calls]
    LoadConfig --> Defaults[Default Values]
    
    %% Environment Variables
    EnvVars --> TSClientID[TS_API_CLIENT_ID]
    EnvVars --> TSClientSecret[TS_API_CLIENT_SECRET]
    EnvVars --> TSAuthKey[TAILSCALE_AUTH_KEY]
    EnvVars --> TSTags[TAILSCALE_TAGS]
    EnvVars --> ReceiverURL[RECEIVER_URL]
    EnvVars --> ListenInterface[LISTEN_INTERFACE]
    
    %% Fluent API Configuration
    FluentAPI --> NetworkConfig[WithNetwork(network)]
    FluentAPI --> InterfaceConfig[WithInterface(iface)]
    FluentAPI --> URLConfig[WithPostURL(url)]
    FluentAPI --> TimeoutConfig[WithTimeout(duration)]
    FluentAPI --> ProcessorConfig[WithProcessor(processor)]
    
    %% Configuration Merging
    TSClientID --> ServerConfig[Server Configuration]
    TSClientSecret --> ServerConfig
    TSAuthKey --> ServerConfig
    TSTags --> ServerConfig
    ReceiverURL --> ServerConfig
    ListenInterface --> ServerConfig
    
    NetworkConfig --> ServerConfig
    InterfaceConfig --> ServerConfig
    URLConfig --> ServerConfig
    TimeoutConfig --> ServerConfig
    ProcessorConfig --> ServerConfig
    
    Defaults --> ServerConfig
    
    %% Configuration Validation
    ServerConfig --> ValidateConfig[Validate Configuration]
    ValidateConfig --> ConfigValid{Configuration Valid?}
    
    ConfigValid -->|No| ConfigError[Configuration Error]
    ConfigValid -->|Yes| ApplyConfig[Apply Configuration to Server]
    
    ApplyConfig --> NetworkSetup[Network Setup]
    ApplyConfig --> ProcessorSetup[Processor Setup]
    ApplyConfig --> ClientSetup[HTTP Client Setup]
    ApplyConfig --> OAuthSetup[OAuth Setup]
    
    %% Setup Results
    NetworkSetup --> ServerReady[Server Ready]
    ProcessorSetup --> ServerReady
    ClientSetup --> ServerReady
    OAuthSetup --> ServerReady
    
    ConfigError --> StartupFailed[Startup Failed]
    ServerReady --> StartupSuccess[Startup Success]
```

These data flow diagrams illustrate how information moves through the post2post system, from initial configuration through request processing and response delivery, including error handling and fallback mechanisms.