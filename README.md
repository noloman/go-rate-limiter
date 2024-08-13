# go-rate-limiter

## Usage

To use the `go-rate-limiter` module with the [chi v5 library](https://github.com/go-chi/chi), you need to configure it and add it to your router's middleware. Below is an example of how to set up and use the rate limiter:

```go
package main

import (
    limiter "github.com/noloman/go-rate-limiter"
    "github.com/go-chi/chi/v5"
    "net/http"
)

func main() {
    // Configure the rate limiter
    config := limiter.LimiterConfig{
        Enabled: true,
        Rps:     2,    // Requests per second
        Burst:   4,    // Maximum burst size
    }

    // Create a new rate limiter handler
    rateLimiterHandler := limiter.RateLimiterHandler{
        Config: config,
    }

    // Initialize your Chi router and add the rate limiter middleware
    r := chi.NewRouter()
    r.Use(rateLimiterHandler.AddRateLimit)

    // Define your routes
    r.Get("/", func(w http.ResponseWriter, r *http.Request) {
        w.Write([]byte("Hello, world!"))
    })

    // Start the server
    http.ListenAndServe(":8080", r)
}
