# Deployment Architecture - Post2Post Library

## 1. Deployment Patterns Overview

```mermaid
graph TB
    subgraph "Deployment Patterns"
        Standalone[Standalone Application]
        Microservices[Microservices Architecture]
        Serverless[Serverless Functions]
        Container[Containerized Deployment]
        Hybrid[Hybrid Cloud/Edge]
    end
    
    subgraph "Network Topologies"
        DirectHTTP[Direct HTTP Communication]
        TailscaleMesh[Tailscale Mesh Network]
        VPNOverlay[VPN Overlay Networks]
        ServiceMesh[Service Mesh Integration]
    end
    
    subgraph "Infrastructure"
        OnPremise[On-Premise]
        PublicCloud[Public Cloud]
        EdgeComputing[Edge Computing]
        MultiCloud[Multi-Cloud]
    end
    
    Standalone --> DirectHTTP
    Standalone --> TailscaleMesh
    
    Microservices --> TailscaleMesh
    Microservices --> ServiceMesh
    
    Serverless --> DirectHTTP
    Serverless --> VPNOverlay
    
    Container --> TailscaleMesh
    Container --> ServiceMesh
    
    Hybrid --> TailscaleMesh
    Hybrid --> VPNOverlay
```

## 2. Standalone Application Deployment

```mermaid
graph TB
    subgraph "Server Infrastructure"
        subgraph "Application Server"
            App[Go Application]
            Post2Post[Post2Post Library]
            ConfigFiles[Configuration Files]
            
            App --> Post2Post
            ConfigFiles --> App
        end
        
        subgraph "Network Interfaces"
            PublicInterface[Public HTTP Interface]
            TailscaleInterface[Tailscale Interface]
            LoopbackInterface[Loopback Interface]
        end
        
        subgraph "External Dependencies"
            TailscaleService[Tailscale Service]
            OAuthProvider[OAuth Provider]
            ExternalAPIs[External APIs]
        end
    end
    
    Post2Post --> PublicInterface
    Post2Post --> TailscaleInterface
    Post2Post --> LoopbackInterface
    
    TailscaleInterface --> TailscaleService
    Post2Post --> OAuthProvider
    App --> ExternalAPIs
    
    %% Environment Configuration
    subgraph "Environment"
        EnvVars[Environment Variables]
        SystemConfig[System Configuration]
        TLSCerts[TLS Certificates]
    end
    
    EnvVars --> Post2Post
    SystemConfig --> TailscaleService
    TLSCerts --> PublicInterface
```

**Configuration Example:**
```bash
# Environment Setup
export TS_API_CLIENT_ID="tskey-client-abc123"
export TS_API_CLIENT_SECRET="secret-xyz789"
export LISTEN_INTERFACE="0.0.0.0"
export PORT="8080"

# Application Deployment
./my-app-with-post2post
```

## 3. Microservices Architecture

```mermaid
graph TB
    subgraph "Service Mesh Layer"
        ServiceMesh[Service Mesh Control Plane]
        TailscaleMesh[Tailscale Overlay Network]
    end
    
    subgraph "Application Services"
        subgraph "Service A"
            AppA[Application A]
            Post2PostA[Post2Post Instance]
            AppA --> Post2PostA
        end
        
        subgraph "Service B"
            AppB[Application B]
            Post2PostB[Post2Post Instance]
            AppB --> Post2PostB
        end
        
        subgraph "Service C"
            AppC[Application C]
            Post2PostC[Post2Post Instance]
            AppC --> Post2PostC
        end
        
        subgraph "API Gateway"
            Gateway[API Gateway]
            LoadBalancer[Load Balancer]
            Gateway --> LoadBalancer
        end
    end
    
    subgraph "Infrastructure Services"
        ConfigService[Configuration Service]
        MonitoringService[Monitoring Service]
        LoggingService[Logging Service]
        SecretManagement[Secret Management]
    end
    
    %% Service Communication
    Post2PostA -.->|Secure Channel| Post2PostB
    Post2PostB -.->|Secure Channel| Post2PostC
    Post2PostC -.->|Secure Channel| Post2PostA
    
    %% Service Mesh Integration
    ServiceMesh --> Post2PostA
    ServiceMesh --> Post2PostB
    ServiceMesh --> Post2PostC
    
    %% Tailscale Integration
    TailscaleMesh -.-> Post2PostA
    TailscaleMesh -.-> Post2PostB
    TailscaleMesh -.-> Post2PostC
    
    %% Infrastructure Integration
    ConfigService --> Post2PostA
    ConfigService --> Post2PostB
    ConfigService --> Post2PostC
    
    MonitoringService --> Post2PostA
    MonitoringService --> Post2PostB
    MonitoringService --> Post2PostC
    
    %% External Access
    Gateway --> Post2PostA
    Gateway --> Post2PostB
    Gateway --> Post2PostC
```

**Kubernetes Deployment Example:**
```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: post2post-service
spec:
  replicas: 3
  selector:
    matchLabels:
      app: post2post-service
  template:
    metadata:
      labels:
        app: post2post-service
    spec:
      containers:
      - name: app
        image: my-app:latest
        env:
        - name: TS_API_CLIENT_ID
          valueFrom:
            secretKeyRef:
              name: tailscale-secrets
              key: client-id
        - name: TS_API_CLIENT_SECRET
          valueFrom:
            secretKeyRef:
              name: tailscale-secrets
              key: client-secret
        ports:
        - containerPort: 8080
```

## 4. Serverless Deployment (AWS Lambda)

```mermaid
graph TB
    subgraph "AWS Infrastructure"
        subgraph "Lambda Functions"
            LambdaReceiver[Lambda Receiver Function]
            LambdaProcessor[Lambda Processor Function]
            LambdaClient[Lambda Client Function]
        end
        
        subgraph "API Gateway"
            APIGW[API Gateway]
            LambdaURL[Lambda Function URLs]
        end
        
        subgraph "AWS Services"
            IAMRoles[IAM Roles]
            SecretsManager[Secrets Manager]
            CloudWatch[CloudWatch Logs]
            VPC[VPC Configuration]
        end
        
        subgraph "External Services"
            TailscaleAPI[Tailscale API]
            TailscaleRelay[Tailscale Relay]
        end
    end
    
    %% API Gateway Integration
    APIGW --> LambdaReceiver
    LambdaURL --> LambdaProcessor
    
    %% Lambda Communication
    LambdaClient --> LambdaURL
    LambdaReceiver --> LambdaProcessor
    
    %% AWS Service Integration
    LambdaReceiver --> IAMRoles
    LambdaProcessor --> IAMRoles
    LambdaClient --> IAMRoles
    
    SecretsManager --> LambdaReceiver
    SecretsManager --> LambdaProcessor
    SecretsManager --> LambdaClient
    
    CloudWatch --> LambdaReceiver
    CloudWatch --> LambdaProcessor
    CloudWatch --> LambdaClient
    
    %% Tailscale Integration
    LambdaReceiver --> TailscaleAPI
    LambdaProcessor --> TailscaleAPI
    LambdaReceiver -.-> TailscaleRelay
    LambdaProcessor -.-> TailscaleRelay
    
    %% VPC Configuration
    VPC --> LambdaReceiver
    VPC --> LambdaProcessor
```

**Lambda Function Configuration:**
```go
// lambda/main.go
package main

import (
    "context"
    "github.com/aws/aws-lambda-go/events"
    "github.com/aws/aws-lambda-go/lambda"
    "github.com/pgdad/post2post"
)

func handler(ctx context.Context, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
    server := post2post.NewServer()
    
    // Generate ephemeral auth key for this invocation
    authKey, err := server.GenerateTailnetKeyFromOAuth(
        true, true, false, "tag:lambda,tag:ephemeral",
    )
    if err != nil {
        return events.APIGatewayProxyResponse{
            StatusCode: 500,
            Body:       err.Error(),
        }, nil
    }
    
    // Process request with Tailscale networking
    // Implementation details...
    
    return events.APIGatewayProxyResponse{
        StatusCode: 200,
        Body:       "Success",
    }, nil
}

func main() {
    lambda.Start(handler)
}
```

## 5. Container Deployment with Docker

```mermaid
graph TB
    subgraph "Container Infrastructure"
        subgraph "Application Containers"
            ClientContainer[Client Container]
            ReceiverContainer[Receiver Container]
            ProcessorContainer[Processor Container]
        end
        
        subgraph "Supporting Containers"
            TailscaleContainer[Tailscale Sidecar]
            MonitoringContainer[Monitoring Sidecar]
            ConfigContainer[Config Sidecar]
        end
        
        subgraph "Container Orchestration"
            DockerCompose[Docker Compose]
            Kubernetes[Kubernetes]
            DockerSwarm[Docker Swarm]
        end
        
        subgraph "Container Registry"
            DockerHub[Docker Hub]
            PrivateRegistry[Private Registry]
            ArtifactRegistry[Artifact Registry]
        end
    end
    
    subgraph "Network Infrastructure"
        BridgeNetwork[Bridge Network]
        OverlayNetwork[Overlay Network]
        TailscaleNetwork[Tailscale Network]
    end
    
    %% Container Relationships
    ClientContainer --> TailscaleContainer
    ReceiverContainer --> TailscaleContainer
    ProcessorContainer --> TailscaleContainer
    
    ClientContainer --> MonitoringContainer
    ReceiverContainer --> MonitoringContainer
    ProcessorContainer --> MonitoringContainer
    
    %% Orchestration
    DockerCompose --> ClientContainer
    DockerCompose --> ReceiverContainer
    DockerCompose --> ProcessorContainer
    
    %% Registry
    DockerHub --> ClientContainer
    PrivateRegistry --> ReceiverContainer
    ArtifactRegistry --> ProcessorContainer
    
    %% Networking
    BridgeNetwork --> ClientContainer
    OverlayNetwork --> ReceiverContainer
    TailscaleNetwork -.-> TailscaleContainer
```

**Docker Compose Example:**
```yaml
version: '3.8'
services:
  client:
    build: ./client
    environment:
      - TS_API_CLIENT_ID=${TS_API_CLIENT_ID}
      - TS_API_CLIENT_SECRET=${TS_API_CLIENT_SECRET}
      - RECEIVER_URL=http://receiver:8082/webhook
    depends_on:
      - receiver
    networks:
      - post2post-network
      
  receiver:
    build: ./receiver
    environment:
      - TS_API_CLIENT_ID=${TS_API_CLIENT_ID}
      - TS_API_CLIENT_SECRET=${TS_API_CLIENT_SECRET}
      - LISTEN_INTERFACE=0.0.0.0
    ports:
      - "8082:8082"
    networks:
      - post2post-network
      
  tailscale-sidecar:
    image: tailscale/tailscale:latest
    environment:
      - TS_AUTHKEY=${TS_AUTHKEY}
      - TS_STATE_DIR=/var/lib/tailscale
    volumes:
      - tailscale-state:/var/lib/tailscale
    cap_add:
      - NET_ADMIN
    networks:
      - post2post-network

networks:
  post2post-network:
    driver: bridge

volumes:
  tailscale-state:
```

## 6. CI/CD Pipeline Integration

```mermaid
graph LR
    subgraph "Source Control"
        GitRepo[Git Repository]
        GitBranches[Feature Branches]
        GitTags[Release Tags]
    end
    
    subgraph "CI Pipeline"
        BuildStage[Build Stage]
        TestStage[Test Stage]
        SecurityScan[Security Scan]
        IntegrationTest[Integration Test]
    end
    
    subgraph "CD Pipeline"
        StagingDeploy[Staging Deployment]
        ProductionDeploy[Production Deployment]
        RollbackMechanism[Rollback Mechanism]
    end
    
    subgraph "Environment Management"
        DevEnv[Development Environment]
        StagingEnv[Staging Environment]
        ProdEnv[Production Environment]
    end
    
    subgraph "Tailscale Integration"
        EphemeralKeys[Ephemeral Key Generation]
        NetworkTesting[Network Testing]
        SecurityValidation[Security Validation]
    end
    
    %% Source Control Flow
    GitRepo --> BuildStage
    GitBranches --> BuildStage
    GitTags --> ProductionDeploy
    
    %% CI Flow
    BuildStage --> TestStage
    TestStage --> SecurityScan
    SecurityScan --> IntegrationTest
    
    %% CD Flow
    IntegrationTest --> StagingDeploy
    StagingDeploy --> ProductionDeploy
    ProductionDeploy --> RollbackMechanism
    
    %% Environment Flow
    StagingDeploy --> StagingEnv
    ProductionDeploy --> ProdEnv
    BuildStage --> DevEnv
    
    %% Tailscale Integration
    IntegrationTest --> EphemeralKeys
    StagingDeploy --> NetworkTesting
    ProductionDeploy --> SecurityValidation
```

**GitHub Actions Example:**
```yaml
name: Deploy Post2Post Application
on:
  push:
    branches: [main]
    
jobs:
  deploy:
    runs-on: ubuntu-latest
    steps:
    - uses: actions/checkout@v3
    
    - name: Setup Go
      uses: actions/setup-go@v3
      with:
        go-version: 1.21
        
    - name: Build Application
      run: |
        go build ./...
        
    - name: Test with Tailscale Integration
      env:
        TS_API_CLIENT_ID: ${{ secrets.TS_API_CLIENT_ID }}
        TS_API_CLIENT_SECRET: ${{ secrets.TS_API_CLIENT_SECRET }}
        TAILSCALE_TAGS: "tag:ci,tag:github-actions"
      run: |
        go test ./...
        
    - name: Deploy to Production
      env:
        TS_API_CLIENT_ID: ${{ secrets.PROD_TS_API_CLIENT_ID }}
        TS_API_CLIENT_SECRET: ${{ secrets.PROD_TS_API_CLIENT_SECRET }}
      run: |
        ./deploy-script.sh
```

## 7. Multi-Environment Architecture

```mermaid
graph TB
    subgraph "Development Environment"
        DevApp[Dev Application]
        DevPost2Post[Dev Post2Post]
        DevTailscale[Dev Tailscale Network]
        
        DevApp --> DevPost2Post
        DevPost2Post -.-> DevTailscale
    end
    
    subgraph "Staging Environment"
        StagingApp[Staging Application]
        StagingPost2Post[Staging Post2Post]
        StagingTailscale[Staging Tailscale Network]
        
        StagingApp --> StagingPost2Post
        StagingPost2Post -.-> StagingTailscale
    end
    
    subgraph "Production Environment"
        ProdApp[Production Application]
        ProdPost2Post[Production Post2Post]
        ProdTailscale[Production Tailscale Network]
        
        ProdApp --> ProdPost2Post
        ProdPost2Post -.-> ProdTailscale
    end
    
    subgraph "Cross-Environment Services"
        SecretManagement[Secret Management]
        MonitoringService[Monitoring]
        LoggingService[Centralized Logging]
        ConfigManagement[Configuration Management]
    end
    
    %% Environment Isolation
    DevTailscale -.-> StagingTailscale
    StagingTailscale -.-> ProdTailscale
    
    %% Shared Services
    SecretManagement --> DevPost2Post
    SecretManagement --> StagingPost2Post
    SecretManagement --> ProdPost2Post
    
    MonitoringService --> DevPost2Post
    MonitoringService --> StagingPost2Post
    MonitoringService --> ProdPost2Post
    
    LoggingService --> DevPost2Post
    LoggingService --> StagingPost2Post
    LoggingService --> ProdPost2Post
    
    ConfigManagement --> DevPost2Post
    ConfigManagement --> StagingPost2Post
    ConfigManagement --> ProdPost2Post
```

This deployment architecture provides flexibility for various infrastructure patterns while maintaining security and scalability through Tailscale integration and proper environment management.