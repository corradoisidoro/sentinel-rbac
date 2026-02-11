package middleware

import (
	"net"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"
)

type RateLimiterConfig struct {
	GlobalRPS   int
	GlobalBurst int

	IPRPS   int
	IPBurst int

	RouteRPS   int
	RouteBurst int
}

type ipLimiter struct {
	limiter  *rate.Limiter
	lastSeen time.Time
}

func NewRateLimiter(cfg RateLimiterConfig) gin.HandlerFunc {
	// Basic config validation
	if cfg.GlobalRPS <= 0 || cfg.IPRPS <= 0 || cfg.RouteRPS <= 0 {
		panic("rate limiter config values must be > 0")
	}

	globalLimiter := rate.NewLimiter(rate.Limit(cfg.GlobalRPS), cfg.GlobalBurst)

	ipLimiters := make(map[string]*ipLimiter)
	routeLimiters := make(map[string]*rate.Limiter)

	var mu sync.Mutex

	// Cleanup goroutine
	go func() {
		ticker := time.NewTicker(5 * time.Minute)
		defer ticker.Stop()

		for range ticker.C {
			now := time.Now()
			mu.Lock()
			for ip, entry := range ipLimiters {
				if now.Sub(entry.lastSeen) > 10*time.Minute {
					delete(ipLimiters, ip)
				}
			}
			mu.Unlock()
		}
	}()

	return func(c *gin.Context) {
		if !globalLimiter.Allow() {
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"error": "global rate limit exceeded",
			})
			return
		}

		now := time.Now()
		ip := clientIP(c)

		mu.Lock()
		ipLim, ok := ipLimiters[ip]
		if !ok {
			ipLim = &ipLimiter{
				limiter:  rate.NewLimiter(rate.Limit(cfg.IPRPS), cfg.IPBurst),
				lastSeen: now,
			}
			ipLimiters[ip] = ipLim
		}
		ipLim.lastSeen = now
		mu.Unlock()

		if !ipLim.limiter.Allow() {
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"error": "too many requests from this IP",
			})
			return
		}

		// Per-route limiter
		routeKey := c.FullPath()
		if routeKey == "" {
			routeKey = c.Request.URL.Path
		}

		mu.Lock()
		routeLim, ok := routeLimiters[routeKey]
		if !ok {
			routeLim = rate.NewLimiter(rate.Limit(cfg.RouteRPS), cfg.RouteBurst)
			routeLimiters[routeKey] = routeLim
		}
		mu.Unlock()

		if !routeLim.Allow() {
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"error": "route rate limit exceeded",
			})
			return
		}

		c.Next()
	}
}

func clientIP(c *gin.Context) string {
	ip := c.ClientIP()
	if parsed := net.ParseIP(ip); parsed != nil {
		return parsed.String()
	}
	return "unknown"
}
