# AGENTS.md - Development Guide for Agentic Coding

This document provides comprehensive guidance for agentic coding agents working on the Hysteria2 VPN Microservices project.

## Project Overview

This is a distributed VPN management system built with microservices architecture:
- **Orchestrator Service**: Master server for VPS node management (Go + Gin)
- **API Service**: REST API and WebSocket server (Go + Fiber)  
- **Agent Service**: gRPC agents for VPS nodes (Go + gRPC)
- **Web Service**: React TypeScript frontend with Ant Design

## Build, Test & Development Commands

### Go Services (orchestrator, agent, api)

```bash
# Build commands
make orchestrator-build    # Build orchestrator service
make agent-build          # Build agent service  
make api-build           # Build API service

# Run in development
make orchestrator-run     # Run orchestrator (port 8081)
make agent-run           # Run agent service
make api-run             # Run API service (port 8080)

# Testing
make orchestrator-test    # Run all orchestrator tests
make agent-test          # Run all agent tests
make api-test            # Run all API tests
make test-all           # Run all tests across services

# Single test file
cd orchestrator-service && go test ./internal/handlers -v
cd api-service && go test ./internal/services -run TestSpecificFunction -v

# Protobuf generation
make proto               # Generate gRPC code from proto files
```

### Web Service (React TypeScript)

```bash
cd web-service
npm install             # Install dependencies
npm run dev            # Development server (port 3000)
npm run build          # Production build
npm run lint           # ESLint with TypeScript
npm run test           # Vitest unit tests
```

### Docker & Infrastructure

```bash
make docker-build       # Build all Docker images
make docker-run         # Start all services
make docker-stop        # Stop all services
make health             # Check service health
make db-up             # Start only database services
```

## Code Style Guidelines

### Go Code Style

**Imports Organization:**
```go
import (
    // Standard library
    "context"
    "fmt"
    "time"
    
    // External dependencies
    "github.com/gin-gonic/gin"
    "github.com/google/uuid"
    "github.com/sirupsen/logrus"
    
    // Internal modules
    "hysteria2-microservices/api-service/internal/models"
    "hysteria2-microservices/api-service/pkg/logger"
)
```

**Package Structure:**
- Use `internal/` for private packages
- Follow domain-driven design: `models/`, `handlers/`, `services/`, `repositories/`
- Interfaces in `interfaces/` subdirectories
- Shared packages in `pkg/`

**Naming Conventions:**
- Package names: lowercase, single words when possible (`handlers`, `services`)
- Struct names: PascalCase (`AuthService`, `UserRepository`)
- Interface names: often end with `er` suffix or simple names (`AuthService`, `UserRepository`)
- Methods: PascalCase for public, camelCase for private
- Variables: camelCase, descriptive names
- Constants: UPPER_SNAKE_CASE or camelCase for exported

**Error Handling:**
```go
// Always handle errors
result, err := someFunction()
if err != nil {
    // Log with context
    h.logger.Error("Operation failed", "error", err, "context", additionalInfo)
    return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
        "error": "Operation failed",
    })
}

// Use fmt.Errorf with %w for wrapping
return nil, fmt.Errorf("failed to process user: %w", err)
```

**Configuration:**
- Use Viper for configuration management
- Environment variables in UPPER_SNAKE_CASE
- Support both .env files and environment variables

**Logging:**
- Use structured logging with logrus
- Log levels: Debug, Info, Warn, Error
- Include relevant context in log entries

### React TypeScript Style

**Imports Organization:**
```tsx
import React, { useState, useEffect } from 'react'
import { BrowserRouter as Router, Routes, Route } from 'react-router-dom'
import { Layout, Button, Table } from 'antd'
import { UserOutlined } from '@ant-design/icons'

// Internal imports
import { useAuth } from './hooks/useAuth'
import { api } from './services/api'
import { User } from './types'
```

**Component Structure:**
```tsx
// Functional components with TypeScript interfaces
interface UserListProps {
  users: User[]
  onUserSelect: (user: User) => void
}

const UserList: React.FC<UserListProps> = ({ users, onUserSelect }) => {
  const [loading, setLoading] = useState(false)
  
  return (
    <div>
      {/* JSX content */}
    </div>
  )
}

export default UserList
```

**TypeScript Rules:**
- Strict mode enabled
- Use interfaces for object shapes
- Prefer union types over enums for simple cases
- Use React.FC for functional components
- Explicitly type useState, useEffect parameters

**State Management:**
- Use React hooks for local state
- Context API for global state (auth, theme)
- Custom hooks for complex logic
- Avoid prop drilling where possible

**Styling:**
- Use Ant Design components when available
- Inline styles for dynamic styles only
- CSS modules or styled-components for component-specific styles
- Follow Ant Design theming guidelines

## Testing Guidelines

### Go Testing

```go
// Test file naming: *_test.go in same package
func TestAuthService_Register(t *testing.T) {
    // Arrange
    service := setupTestService()
    
    // Act
    user, err := service.Register(ctx, "test", "test@example.com", "password")
    
    // Assert
    assert.NoError(t, err)
    assert.Equal(t, "test", user.Username)
}

// Use testify for assertions
// Use table-driven tests for multiple cases
// Mock external dependencies using interfaces
```

### React Testing

```tsx
// Use Vitest + React Testing Library
import { render, screen, fireEvent } from '@testing-library/react'
import { describe, it, expect } from 'vitest'

describe('UserList', () => {
  it('renders user list correctly', () => {
    const users = [{ id: '1', name: 'John' }]
    render(<UserList users={users} />)
    
    expect(screen.getByText('John')).toBeInTheDocument()
  })
})
```

## Database & ORM

**GORM Usage:**
- Use GORM for database operations
- Define models in `models/` packages
- Use migrations in SQL files in `migrations/`
- Follow Go naming conventions for JSON tags

**Repository Pattern:**
- Abstract database operations behind interfaces
- Use dependency injection for repositories
- Handle transactions at service layer

## API Guidelines

**REST API (Fiber):**
- Use Fiber framework for HTTP server
- Follow RESTful conventions
- Use middleware for auth, logging, rate limiting
- Return consistent JSON responses
- Use appropriate HTTP status codes

**gRPC Services:**
- Define services in `.proto` files
- Generate code using `make proto`
- Implement streaming for real-time data
- Use proper error handling with gRPC status codes

## WebSocket Integration

**Real-time Features:**
- Use WebSocket for live updates
- Socket.IO for client-server communication
- Emit events for traffic data, node status
- Handle connection/disconnection gracefully

## Graceful Shutdown

**Implementation Guidelines:**
- Handle SIGINT and SIGTERM signals for clean shutdown
- Use context cancellation for coordinated shutdown across goroutines
- Set reasonable timeouts (30-60 seconds) for shutdown operations
- Close resources in reverse order of initialization (connections, databases, caches)
- Log shutdown progress and errors for debugging

**Resource Cleanup:**
- Database connections: Use defer with error logging
- Redis/cache connections: Use defer with error logging
- gRPC connections: Use GracefulStop() for servers, Close() for clients
- HTTP servers: Use ShutdownWithContext() with timeout
- WebSocket connections: Handle via Fiber/Gin shutdown

**Testing:**
- Send SIGTERM signal to running services
- Verify logs show clean shutdown sequence
- Check no resource leaks (connections remain open)
- Test with `kill -TERM <pid>` or Ctrl+C in development

## Security Considerations

**Authentication:**
- JWT tokens with refresh mechanism
- Password hashing with Argon2
- Secure token storage in Redis
- Rate limiting on authentication endpoints

**Input Validation:**
- Use struct tags for validation
- Validate all user inputs
- Sanitize data before processing
- Use parameterized queries for database

## Environment Configuration

**Development:**
- Use `.env` files for local development
- Default ports: API (8080), Orchestrator (8081), Web (3000)
- Docker Compose for full stack
- Hot reload for React development

**Production:**
- Use environment variables
- Run behind reverse proxy (nginx)
- Use SSL/TLS certificates
- Configure proper logging levels

## Common Patterns

**Dependency Injection:**
```go
func NewAuthHandler(
    authService interfaces.AuthService,
    logger *logger.Logger,
) *AuthHandler {
    return &AuthHandler{
        authService: authService,
        logger:      logger,
    }
}
```

**Error Response Format:**
```go
return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
    "error": "Detailed error message",
    "code":  "ERROR_CODE",
})
```

**Success Response Format:**
```go
return c.JSON(fiber.Map{
    "data": result,
    "message": "Operation successful",
})
```

## Performance Considerations

- Use connection pooling for database
- Implement caching with Redis
- Use goroutines for concurrent operations
- Profile with Go's pprof tools
- Optimize React bundle size with code splitting

## Git Workflow

- Feature branches for new functionality
- Descriptive commit messages
- Code reviews for all changes
- Automated tests in CI/CD
- Semantic versioning for releases