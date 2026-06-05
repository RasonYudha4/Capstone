package middleware

import (
	"net/http"
	"sync"
	"time"

	"auth-service/models"

	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"
)

type client struct {
	limiter  *rate.Limiter
	lastSeen time.Time
}

var (
	clients = make(map[string]*client)
	mu      sync.Mutex
)

func init() {
	// Start a background goroutine to clean up old rate limiter instances every 10 minutes
	go func() {
		for {
			time.Sleep(10 * time.Minute)
			mu.Lock()
			for ip, c := range clients {
				if time.Since(c.lastSeen) > 20*time.Minute {
					delete(clients, ip)
				}
			}
			mu.Unlock()
		}
	}()
}

// RateLimit creates a middleware that limits requests by IP.
// r: rate limit (requests per second)
// b: burst size
func RateLimit(r rate.Limit, b int) gin.HandlerFunc {
	return func(c *gin.Context) {
		ip := c.ClientIP()

		mu.Lock()
		v, exists := clients[ip]
		if !exists {
			v = &client{
				limiter: rate.NewLimiter(r, b),
			}
			clients[ip] = v
		}
		v.lastSeen = time.Now()
		limiter := v.limiter
		mu.Unlock()

		if !limiter.Allow() {
			c.AbortWithStatusJSON(http.StatusTooManyRequests, models.APIResponse{
				Success: false,
				Message: "Too many requests. Please wait and try again.",
			})
			return
		}

		c.Next()
	}
}
