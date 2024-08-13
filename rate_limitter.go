package middleware

import (
	"fmt"
	"golang.org/x/time/rate"
	"net"
	"net/http"
	"runtime/debug"
	"sync"
	"time"
)

type LimiterConfig struct {
	Enabled bool
	Rps     int
	Burst   int
}

func AddRateLimit(config LimiterConfig, next http.Handler) http.Handler {
	type client struct {
		limiter  *rate.Limiter
		lastSeen time.Time
	}

	var (
		mu      sync.Mutex
		clients = make(map[string]*client)
	)

	go func() {
		for {
			time.Sleep(time.Minute)
			mu.Lock()
			for ip, client := range clients {
				if time.Since(client.lastSeen) > 3*time.Minute {
					delete(clients, ip)
				}
			}
			mu.Unlock()
		}
	}()

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if config.Enabled {
			ip, _, err := net.SplitHostPort(r.RemoteAddr)
			if err != nil {
				// Handle cases where the address does not contain a port
				ip = r.RemoteAddr
				if net.ParseIP(ip) == nil {
					fmt.Println("Invalid IP address:", ip)
					http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
					return
				}
			}
			mu.Lock()
			if _, found := clients[ip]; !found {
				clients[ip] = &client{limiter: rate.NewLimiter(
					rate.Limit(config.Rps), config.Burst,
				)}
			}
			if !clients[ip].limiter.Allow() {
				mu.Unlock()
				rateLimitExceededResponse(w, r, fmt.Errorf("rate limit exceeded"))
				return
			}
			mu.Unlock()
		}
		next.ServeHTTP(w, r)
	})
}

func ServerError(w http.ResponseWriter, r *http.Request, err error) {
	trace := fmt.Sprintf("%s\n%s", err.Error(), debug.Stack())
	fmt.Println(trace)
	r.Header.Set("Content-Type", "text/html")
	http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
}

func rateLimitExceededResponse(w http.ResponseWriter, r *http.Request, err error) {
	trace := fmt.Sprintf("%s\n%s", err.Error(), debug.Stack())
	fmt.Println(trace)
	r.Header.Set("Content-Type", "text/html")
	http.Error(w, http.StatusText(http.StatusTooManyRequests), http.StatusTooManyRequests)
}
