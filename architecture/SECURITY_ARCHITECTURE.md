# Security Architecture - Post2Post Library

## 1. Security Overview

```mermaid
graph TB
    subgraph "Security Layers"
        subgraph "Network Security"
            TailscaleEncryption[Tailscale End-to-End Encryption]
            TLSTermination[TLS/HTTPS Termination]
            NetworkSegmentation[Network Segmentation]
        end
        
        subgraph "Authentication & Authorization"
            OAuthAuthentication[OAuth 2.0 Authentication]
            EphemeralKeys[Ephemeral Key Management]
            TagBasedAccess[Tag-Based Access Control]
            IAMIntegration[IAM Role Integration]
        end
        
        subgraph "Data Security"
            PayloadEncryption[Payload Encryption in Transit]
            SecretManagement[Secret Management]
            DataValidation[Input Data Validation]
            SanitizedLogging[Sanitized Logging]
        end
        
        subgraph "Runtime Security"
            ThreadSafety[Thread-Safe Operations]
            TimeoutProtection[Timeout Protection]
            ErrorHandling[Secure Error Handling]
            ResourceLimiting[Resource Limiting]
        end
    end
    
    subgraph "Security Controls"
        ZeroTrustPrinciples[Zero Trust Principles]
        PrincipleOfLeastPrivilege[Principle of Least Privilege]
        DefenseInDepth[Defense in Depth]
        ContinuousMonitoring[Continuous Monitoring]
    end
    
    %% Layer Relationships
    TailscaleEncryption --> PayloadEncryption
    OAuthAuthentication --> EphemeralKeys
    EphemeralKeys --> TagBasedAccess
    TagBasedAccess --> NetworkSegmentation
    
    %% Control Relationships
    ZeroTrustPrinciples --> TailscaleEncryption
    ZeroTrustPrinciples --> OAuthAuthentication
    PrincipleOfLeastPrivilege --> TagBasedAccess
    PrincipleOfLeastPrivilege --> EphemeralKeys
    DefenseInDepth --> TLSTermination
    DefenseInDepth --> DataValidation
    ContinuousMonitoring --> SanitizedLogging
    ContinuousMonitoring --> ErrorHandling
```

## 2. Tailscale Mesh Network Security

```mermaid
graph TB
    subgraph "Tailscale Security Model"
        subgraph "Identity & Access"
            MachineIdentity[Machine Identity]
            UserIdentity[User Identity]
            DeviceCertificates[Device Certificates]
            ACLPolicies[ACL Policies]
        end
        
        subgraph "Network Encryption"
            WireGuardProtocol[WireGuard Protocol]
            End2EndEncryption[End-to-End Encryption]
            KeyRotation[Automatic Key Rotation]
            PerfectForwardSecrecy[Perfect Forward Secrecy]
        end
        
        subgraph "Network Isolation"
            ZeroTrustNetworking[Zero Trust Networking]
            NetworkSegmentation[Network Segmentation]
            TrafficInspection[Traffic Inspection]
            MalwareProtection[Malware Protection]
        end
        
        subgraph "Device Management"
            EphemeralDevices[Ephemeral Devices]
            DeviceApproval[Device Approval Process]
            DeviceRevocation[Device Revocation]
            DeviceMonitoring[Device Monitoring]
        end
    end
    
    subgraph "Post2Post Integration"
        AuthKeyGeneration[OAuth-Based Auth Key Generation]
        TaggedDevices[Tagged Device Creation]
        AutomaticCleanup[Automatic Device Cleanup]
        SecureChannels[Secure Communication Channels]
    end
    
    %% Identity Flow
    MachineIdentity --> DeviceCertificates
    UserIdentity --> ACLPolicies
    DeviceCertificates --> WireGuardProtocol
    ACLPolicies --> NetworkSegmentation
    
    %% Encryption Flow
    WireGuardProtocol --> End2EndEncryption
    End2EndEncryption --> KeyRotation
    KeyRotation --> PerfectForwardSecrecy
    
    %% Device Lifecycle
    EphemeralDevices --> DeviceApproval
    DeviceApproval --> DeviceMonitoring
    DeviceMonitoring --> DeviceRevocation
    
    %% Post2Post Integration
    AuthKeyGeneration --> TaggedDevices
    TaggedDevices --> EphemeralDevices
    EphemeralDevices --> AutomaticCleanup
    AutomaticCleanup --> SecureChannels
```

### **Tailscale Security Features**

1. **WireGuard Protocol**
   - State-of-the-art VPN protocol
   - Cryptographically sound design
   - Minimal attack surface
   - High performance encryption

2. **Zero Trust Architecture**
   - No implicit trust based on network location
   - Every connection authenticated and authorized
   - Continuous verification of trust
   - Encrypted end-to-end communication

3. **Ephemeral Device Security**
   ```go
   // Automatic ephemeral device creation
   authKey, err := server.GenerateTailnetKeyFromOAuth(
       true,  // reusable for multiple connections
       true,  // ephemeral - automatically removed when offline
       false, // not preauthorized - requires explicit approval
       "tag:ephemeral-device,tag:post2post",
   )
   ```

## 3. OAuth 2.0 Security Implementation

```mermaid
sequenceDiagram
    participant App as "Post2Post Application"
    participant OAuth as "OAuth Provider"
    participant TailscaleAPI as "Tailscale API"
    participant SecretStore as "Secret Store"
    
    Note over App, SecretStore: OAuth 2.0 Client Credentials Flow with Security Controls
    
    App->>SecretStore: Retrieve OAuth Credentials
    SecretStore-->>App: Client ID & Secret (encrypted)
    
    App->>App: Validate credential format
    App->>App: Set API acknowledgment flag
    
    App->>OAuth: POST /oauth/token<br/>grant_type=client_credentials<br/>scope=devices
    Note over App, OAuth: HTTPS with client authentication
    
    OAuth->>OAuth: Validate client credentials
    OAuth->>OAuth: Check scope permissions
    OAuth-->>App: Access Token (short-lived)
    
    App->>App: Validate token format
    App->>TailscaleAPI: POST /tailnet/-/keys<br/>Authorization: Bearer <token>
    Note over App, TailscaleAPI: HTTPS with bearer token
    
    TailscaleAPI->>TailscaleAPI: Validate token & permissions
    TailscaleAPI->>TailscaleAPI: Generate ephemeral auth key
    TailscaleAPI-->>App: Auth Key (tskey-auth-...)
    
    App->>App: Log key generation (without key value)
    App->>App: Use key for secure communication
    
    Note over App: Key automatically expires with device
```

### **OAuth Security Controls**

1. **Credential Management**
   ```go
   // Secure credential handling
   clientID := os.Getenv("TS_API_CLIENT_ID")
   clientSecret := os.Getenv("TS_API_CLIENT_SECRET")
   
   if clientID == "" || clientSecret == "" {
       return "", fmt.Errorf("OAuth credentials not configured")
   }
   
   // Never log sensitive credentials
   log.Printf("Using OAuth client: %s...", clientID[:min(8, len(clientID))])
   ```

2. **Token Security**
   - Short-lived access tokens (1 hour expiry)
   - Automatic token refresh
   - No persistent token storage
   - Secure token transmission (HTTPS only)

3. **Scope Limitation**
   ```go
   credentials := clientcredentials.Config{
       ClientID:     clientID,
       ClientSecret: clientSecret,
       TokenURL:     "https://api.tailscale.com/api/v2/oauth/token",
       Scopes:       []string{"devices"}, // Minimal required scope
   }
   ```

## 4. Data Protection and Privacy

```mermaid
graph TB
    subgraph "Data Classification"
        PublicData[Public Data]
        InternalData[Internal Data]
        ConfidentialData[Confidential Data]
        RestrictedData[Restricted Data]
    end
    
    subgraph "Data Protection Measures"
        subgraph "Encryption"
            TransitEncryption[Encryption in Transit]
            RestEncryption[Encryption at Rest]
            KeyManagement[Key Management]
        end
        
        subgraph "Access Control"
            DataClassification[Data Classification]
            AccessPolicies[Access Policies]
            AuditLogging[Audit Logging]
        end
        
        subgraph "Data Handling"
            DataMinimization[Data Minimization]
            DataRetention[Data Retention Policies]
            SecureDisposal[Secure Data Disposal]
        end
    end
    
    subgraph "Privacy Controls"
        PIIIdentification[PII Identification]
        DataAnonymization[Data Anonymization]
        ConsentManagement[Consent Management]
        ComplianceReporting[Compliance Reporting]
    end
    
    %% Data Flow
    PublicData --> DataMinimization
    InternalData --> TransitEncryption
    ConfidentialData --> RestEncryption
    RestrictedData --> KeyManagement
    
    %% Protection Flow
    TransitEncryption --> AccessPolicies
    RestEncryption --> AuditLogging
    KeyManagement --> DataRetention
    
    %% Privacy Flow
    DataClassification --> PIIIdentification
    PIIIdentification --> DataAnonymization
    DataAnonymization --> ConsentManagement
    ConsentManagement --> ComplianceReporting
```

### **Data Protection Implementation**

1. **Payload Security**
   ```go
   // Secure payload handling
   func (s *Server) webhookHandler(w http.ResponseWriter, r *http.Request) {
       // Read body with size limits
       body, err := io.ReadAll(io.LimitReader(r.Body, MaxPayloadSize))
       if err != nil {
           log.Printf("Error reading request body: %v", err)
           http.Error(w, "Request too large", http.StatusRequestEntityTooLarge)
           return
       }
       
       // Validate JSON structure
       var requestData PostData
       if err := json.Unmarshal(body, &requestData); err != nil {
           log.Printf("Invalid JSON payload")
           http.Error(w, "Invalid JSON", http.StatusBadRequest)
           return
       }
       
       // Sanitize for logging (remove sensitive data)
       sanitizedPayload := sanitizeForLogging(requestData.Payload)
       log.Printf("Processing payload: %+v", sanitizedPayload)
   }
   ```

2. **Secret Management**
   ```go
   // Secure secret handling
   func sanitizeForLogging(data interface{}) interface{} {
       // Remove sensitive fields from logging
       sensitiveFields := []string{
           "password", "secret", "key", "token", 
           "auth", "credential", "private",
       }
       
       // Implementation to redact sensitive data
       return redactSensitiveData(data, sensitiveFields)
   }
   ```

## 5. Access Control and Authorization

```mermaid
graph TB
    subgraph "Access Control Matrix"
        subgraph "Subjects"
            Users[Users]
            Applications[Applications]
            Services[Services]
            Devices[Devices]
        end
        
        subgraph "Objects"
            APIs[API Endpoints]
            Data[Data Resources]
            Networks[Network Resources]
            Functions[Function Calls]
        end
        
        subgraph "Actions"
            Read[Read Operations]
            Write[Write Operations]
            Execute[Execute Operations]
            Admin[Administrative Operations]
        end
        
        subgraph "Conditions"
            TimeConstraints[Time Constraints]
            LocationConstraints[Location Constraints]
            DeviceConstraints[Device Constraints]
            NetworkConstraints[Network Constraints]
        end
    end
    
    subgraph "Authorization Mechanisms"
        RBAC[Role-Based Access Control]
        ABAC[Attribute-Based Access Control]
        TagBasedAccess[Tag-Based Access Control]
        PolicyEngine[Policy Engine]
    end
    
    %% Access Control Flow
    Users --> RBAC
    Applications --> ABAC
    Services --> TagBasedAccess
    Devices --> PolicyEngine
    
    %% Object Protection
    RBAC --> APIs
    ABAC --> Data
    TagBasedAccess --> Networks
    PolicyEngine --> Functions
    
    %% Condition Enforcement
    TimeConstraints --> RBAC
    LocationConstraints --> ABAC
    DeviceConstraints --> TagBasedAccess
    NetworkConstraints --> PolicyEngine
```

### **Tailscale ACL Configuration**

```json
{
  "tagOwners": {
    "tag:post2post-client": [],
    "tag:post2post-receiver": [],
    "tag:ephemeral-device": [],
    "tag:production": ["admin@company.com"],
    "tag:development": ["dev-team@company.com"]
  },
  
  "groups": {
    "group:post2post-admins": ["admin@company.com"],
    "group:post2post-developers": ["dev1@company.com", "dev2@company.com"],
    "group:post2post-services": []
  },
  
  "acls": [
    {
      "action": "accept",
      "src": ["group:post2post-admins"],
      "dst": ["*:*"]
    },
    {
      "action": "accept", 
      "src": ["tag:post2post-client"],
      "dst": ["tag:post2post-receiver:8080", "tag:post2post-receiver:8082"]
    },
    {
      "action": "accept",
      "src": ["tag:ephemeral-device"],
      "dst": ["tag:post2post-receiver:8080", "tag:post2post-receiver:8082"]
    },
    {
      "action": "accept",
      "src": ["tag:development"],
      "dst": ["tag:development:*"]
    },
    {
      "action": "deny",
      "src": ["tag:development"],
      "dst": ["tag:production:*"]
    }
  ]
}
```

## 6. Security Monitoring and Incident Response

```mermaid
graph TB
    subgraph "Security Monitoring"
        subgraph "Detection"
            LogAnalysis[Log Analysis]
            AnomalyDetection[Anomaly Detection]
            ThreatIntelligence[Threat Intelligence]
            BehaviorAnalysis[Behavior Analysis]
        end
        
        subgraph "Alerting"
            SecurityAlerts[Security Alerts]
            EscalationPolicies[Escalation Policies]
            NotificationChannels[Notification Channels]
            AlertCorrelation[Alert Correlation]
        end
        
        subgraph "Response"
            IncidentResponse[Incident Response]
            AutomaticMitigation[Automatic Mitigation]
            ForensicAnalysis[Forensic Analysis]
            RecoveryProcedures[Recovery Procedures]
        end
    end
    
    subgraph "Security Metrics"
        AuthenticationMetrics[Authentication Metrics]
        NetworkMetrics[Network Security Metrics]
        AccessMetrics[Access Control Metrics]
        VulnerabilityMetrics[Vulnerability Metrics]
    end
    
    %% Monitoring Flow
    LogAnalysis --> SecurityAlerts
    AnomalyDetection --> AlertCorrelation
    ThreatIntelligence --> IncidentResponse
    BehaviorAnalysis --> AutomaticMitigation
    
    %% Metrics Flow
    AuthenticationMetrics --> LogAnalysis
    NetworkMetrics --> AnomalyDetection
    AccessMetrics --> BehaviorAnalysis
    VulnerabilityMetrics --> ThreatIntelligence
```

### **Security Logging Implementation**

```go
// Security-focused logging
type SecurityLogger struct {
    logger *log.Logger
}

func (sl *SecurityLogger) LogAuthenticationAttempt(clientID string, success bool, reason string) {
    event := SecurityEvent{
        Type:      "authentication",
        Timestamp: time.Now(),
        ClientID:  redactSensitive(clientID),
        Success:   success,
        Reason:    reason,
        Source:    getClientIP(),
    }
    
    sl.logger.Printf("SECURITY_EVENT: %+v", event)
    
    if !success {
        sl.alertSecurityTeam("Authentication failure", event)
    }
}

func (sl *SecurityLogger) LogNetworkConnection(sourceIP, destIP string, protocol string, success bool) {
    event := SecurityEvent{
        Type:      "network_connection",
        Timestamp: time.Now(),
        SourceIP:  sourceIP,
        DestIP:    destIP,
        Protocol:  protocol,
        Success:   success,
    }
    
    sl.logger.Printf("SECURITY_EVENT: %+v", event)
    
    // Check for suspicious patterns
    if sl.detectAnomalousTraffic(sourceIP, destIP) {
        sl.alertSecurityTeam("Anomalous network traffic", event)
    }
}
```

## 7. Compliance and Regulatory Considerations

```mermaid
graph TB
    subgraph "Compliance Frameworks"
        GDPR[GDPR Compliance]
        SOC2[SOC 2 Compliance]
        HIPAA[HIPAA Compliance]
        PCI_DSS[PCI DSS Compliance]
        ISO27001[ISO 27001 Compliance]
    end
    
    subgraph "Security Controls"
        DataProtection[Data Protection Controls]
        AccessControl[Access Control Requirements]
        AuditTrails[Audit Trail Requirements]
        IncidentResponse[Incident Response Procedures]
        RiskManagement[Risk Management Processes]
    end
    
    subgraph "Documentation"
        SecurityPolicies[Security Policies]
        ProcedureDocuments[Procedure Documents]
        ComplianceReports[Compliance Reports]
        AuditDocumentation[Audit Documentation]
        TrainingMaterials[Training Materials]
    end
    
    %% Compliance Mapping
    GDPR --> DataProtection
    SOC2 --> AccessControl
    HIPAA --> AuditTrails
    PCI_DSS --> IncidentResponse
    ISO27001 --> RiskManagement
    
    %% Documentation Requirements
    DataProtection --> SecurityPolicies
    AccessControl --> ProcedureDocuments
    AuditTrails --> ComplianceReports
    IncidentResponse --> AuditDocumentation
    RiskManagement --> TrainingMaterials
```

This security architecture ensures that the post2post library provides robust protection at multiple layers while maintaining usability and performance. The integration with Tailscale's zero-trust networking model adds an additional layer of security for sensitive communications.