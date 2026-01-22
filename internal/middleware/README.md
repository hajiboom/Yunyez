# Middleware Package

This package contains commonly used HTTP middleware for the Yunyez project.

## Available Middlewares

### 1. CORS Middleware
Handles Cross-Origin Resource Sharing (CORS) headers to allow cross-domain requests.

### 2. Logger Middleware
Logs HTTP request details including status codes, IP addresses, request methods, URIs, and latency.

### 3. Recovery Middleware
Recovers from panics in handlers and prevents server crashes by logging the error and returning a 500 response.

### 4. Rate Limit Middleware
Implements rate limiting using a token bucket algorithm to prevent abuse and control request frequency.

### 5. Security Headers Middleware
Sets common security headers to enhance application security:
- X-Frame-Options
- X-Content-Type-Options
- X-XSS-Protection
- Strict-Transport-Security
- Content-Security-Policy
- Referrer-Policy
- Permissions-Policy

### 6. Authentication Middleware
Validates JWT tokens in the Authorization header for protected routes.

## Usage

### Using Individual Middleware

```go
package main

import (
    "github.com/gin-gonic/gin"
    "go.uber.org/zap"
    "yunyez/internal/middleware"
)

func main() {
    r := gin.Default()
    
    // Initialize logger
    logger, _ := zap.NewProduction()
    defer logger.Sync()
    
    // Apply individual middleware
    r.Use(middleware.CORSMiddleware())
    r.Use(middleware.LoggerToFile(logger))
    r.Use(middleware.RecoveryMiddleware(logger))
    r.Use(middleware.SecurityHeadersMiddleware())
    r.Use(middleware.RateLimitMiddleware(middleware.RateLimitConfig{
        Mode:  middleware.LocalMode, // Use local mode, change to DistributedMode for multi-instance setups
        Limit: rate.Every(1 * time.Second), // Limit to 1 request per second
        Burst: 10, // Allow bursts of up to 10 requests
    }))
    
    // Protected route with auth middleware
    protected := r.Group("/api/protected")
    protected.Use(middleware.AuthMiddleware("your-secret-key"))
    protected.GET("/data", func(c *gin.Context) {
        c.JSON(200, gin.H{"message": "Protected data"})
    })
    
    r.Run(":8080")
}
```

### Using the SetupMiddlewares Helper

```go
package main

import (
    "github.com/gin-gonic/gin"
    "go.uber.org/zap"
    "yunyez/internal/middleware"
)

func main() {
    r := gin.Default()
    
    // Initialize logger
    logger, _ := zap.NewProduction()
    defer logger.Sync()
    
    // Apply all common middlewares at once
    middlewares := middleware.SetupMiddlewares(logger, "your-jwt-secret")
    r.Use(middlewares...)
    
    r.Run(":8080")
}
```

### Applying Middlewares to Specific Groups

```go
// Apply middlewares to specific route groups
api := r.Group("/api")
api.Use(middleware.LoggerToFile(logger), middleware.CORSMiddleware())

admin := r.Group("/admin")
admin.Use(middleware.AuthMiddleware("secret"), middleware.SecurityHeadersMiddleware())
```

## Configuration

Most middlewares are configurable:

- **Rate Limiter**: Configure the rate limit and burst size
- **Auth Middleware**: Pass your JWT secret key
- **Logger**: Pass your configured zap logger instance

## Best Practices

1. Order matters: Generally, logging and recovery should come first
2. Use authentication middleware only on protected routes
3. Adjust rate limits based on your application's needs
4. Customize CORS settings for production environments (don't use "*" wildcard)
5. Use strong JWT secrets in production