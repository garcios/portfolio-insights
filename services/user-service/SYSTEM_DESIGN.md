# User Service - System Design

## Executive Summary

The **User Service** is the central authority for user identity and profile management. It interfaces with an external **Login Consent Provider** (e.g., ORY Hydra) to handle authentication flows and manages user profile data (names, settings, preferences). It ensures that all other services have a consistent view of the user.

## Architecture Overview

The following C4 Container diagram shows the service's interactions.

```mermaid
graph TD
    %% Subgraph Context
    subgraph "Portfolio Insights Platform"
        gateway[API Gateway]
        
        subgraph "Services"
            svc_user["User Service<br/>(Go, gRPC)"]
        end
        
        subgraph "Data Store"
            db[("User DB<br/>(PostgreSQL)")]
        end
        
        ext_auth["Login Consent Provider<br/>(e.g., ORY Hydra)"]
    end
    
    user((User))
    
    %% Relationships
    user -->|Login / Update Profile| gateway
    
    gateway -->|gRPC: GetProfile| svc_user
    gateway -->|OAuth2 / OIDC| ext_auth
    
    ext_auth -->|HTTP: Verify Login/Consent| svc_user
    
    svc_user -->|Persists Data| db
    
    %% Styling
    classDef service fill:#f9f,stroke:#333,stroke-width:2px;
    classDef db fill:#ff9,stroke:#333,stroke-width:2px;
    classDef external fill:#fff,stroke:#333,stroke-dasharray: 5 5;
    
    class svc_user service;
    class db db;
    class ext_auth external;
```

## Tech Stack

*   **Language**: Golang (1.24+)
*   **Communication**: gRPC (Internal APIs), HTTP (Provider Callbacks)
*   **Database**: PostgreSQL
*   **Migration**: Golang-Migrate
*   **Observability**: Prometheus, Slog

## Data Design

The core entity is the `User`.

```mermaid
erDiagram
    USER {
        uuid id PK
        string email UK
        string username UK
        string password_hash
        string first_name
        string last_name
        string role "USER, ADMIN"
        jsonb preferences "Settings"
        timestamp created_at
        timestamp updated_at
        timestamp last_login_at
    }
```

## API Interface

The service exposes a gRPC interface.

| Method | Request | Response | Description |
|--------|---------|----------|-------------|
| `GetUser` | `id` | `User` | Retrieves full user profile. |
| `UpdateUser` | `id`, `changes` | `User` | Updates editable fields (Name, Preferences). |
| `ValidateCredentials` | `email`, `password` | `Result` | Used internally by the Login flow to verify passwords. |
| `CreateUser` | `RegisterInput` | `User` | Registration logic (Hashing password, creating DB record). |

## Key Workflows

### Verifying User (Login Flow)

This workflow describes how the User Service acts as the "Identity Provider" backend for the Login Consent Provider.

1.  **Initiate**: User clicks "Login" in Web App, redirected to Login Consent Provider.
2.  **Redirect**: Provider redirects user to the `Login UI` (part of the platform frontend/gateway).
3.  **Submit**: User enters Email/Password.
4.  **Verify**:
    *   `Login UI` calls `User Service` (`ValidateCredentials`).
    *   Service checks DB (bcrypt comparison).
5.  **Consent**: 
    *   If valid, `Login UI` tells Provider to "Accept Login".
    *   Provider issues ID Token / Access Token.
6.  **Complete**: User is redirected back to the App with a valid session.

## Scalability & Trade-offs

### Scalability strategies
*   **Caching**: User profiles are high-read, low-write. Redis can be used to cache `GetUser` responses, invalidated on `UpdateUser`.
*   **Stateless Ops**: Authentication state is managed by the OIDC Provider (Tokens), not the User Service, allowing horizontal scaling.

### Trade-offs
*   **External Dependency**: Critical reliance on the Login Consent Provider. If Hydra/Auth0 is down, no one can log in.
*   **Complexity**: Implementing a custom Login UI to interface with Hydra is more complex than a monolithic Auth implementation but offers standard OAuth2/OIDC compliance.
