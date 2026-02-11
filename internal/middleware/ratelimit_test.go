package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/corradoisidoro/sentinel-rbac/internal/middleware"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func newTestRouter(cfg middleware.RateLimiterConfig) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(middleware.NewRateLimiter(cfg))

	r.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})

	return r
}

func performRequest(r http.Handler, ip, path string) *httptest.ResponseRecorder {
	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", path, nil)
	req.RemoteAddr = ip + ":1234"
	r.ServeHTTP(w, req)
	return w
}

func TestGlobalRateLimit(t *testing.T) {
	r := newTestRouter(middleware.RateLimiterConfig{
		GlobalRPS:   1,
		GlobalBurst: 1,
		IPRPS:       100,
		IPBurst:     100,
		RouteRPS:    100,
		RouteBurst:  100,
	})

	// First request allowed
	res1 := performRequest(r, "1.1.1.1", "/test")
	assert.Equal(t, 200, res1.Code)

	// Second request immediately should be rate-limited
	res2 := performRequest(r, "1.1.1.1", "/test")
	assert.Equal(t, 429, res2.Code)
}

func TestIPRateLimit(t *testing.T) {
	r := newTestRouter(middleware.RateLimiterConfig{
		GlobalRPS:   100,
		GlobalBurst: 100,
		IPRPS:       1,
		IPBurst:     1,
		RouteRPS:    100,
		RouteBurst:  100,
	})

	// First request allowed
	res1 := performRequest(r, "2.2.2.2", "/test")
	assert.Equal(t, 200, res1.Code)

	// Second request from same IP blocked
	res2 := performRequest(r, "2.2.2.2", "/test")
	assert.Equal(t, 429, res2.Code)

	// Different IP should still be allowed
	res3 := performRequest(r, "3.3.3.3", "/test")
	assert.Equal(t, 200, res3.Code)
}

func TestRouteRateLimit(t *testing.T) {
	r := newTestRouter(middleware.RateLimiterConfig{
		GlobalRPS:   100,
		GlobalBurst: 100,
		IPRPS:       100,
		IPBurst:     100,
		RouteRPS:    1,
		RouteBurst:  1,
	})

	// First request allowed
	res1 := performRequest(r, "4.4.4.4", "/test")
	assert.Equal(t, 200, res1.Code)

	// Second request blocked
	res2 := performRequest(r, "4.4.4.4", "/test")
	assert.Equal(t, 429, res2.Code)
}

func TestBurstBehavior(t *testing.T) {
	r := newTestRouter(middleware.RateLimiterConfig{
		GlobalRPS:   1,
		GlobalBurst: 3, // allow 3 quick bursts
		IPRPS:       100,
		IPBurst:     100,
		RouteRPS:    100,
		RouteBurst:  100,
	})

	// Burst should allow 3 requests
	res1 := performRequest(r, "5.5.5.5", "/test")
	res2 := performRequest(r, "5.5.5.5", "/test")
	res3 := performRequest(r, "5.5.5.5", "/test")

	assert.Equal(t, 200, res1.Code)
	assert.Equal(t, 200, res2.Code)
	assert.Equal(t, 200, res3.Code)

	// Fourth should be blocked
	res4 := performRequest(r, "5.5.5.5", "/test")
	assert.Equal(t, 429, res4.Code)
}

func TestLimiterReset(t *testing.T) {
	r := newTestRouter(middleware.RateLimiterConfig{
		GlobalRPS:   1,
		GlobalBurst: 1,
		IPRPS:       1,
		IPBurst:     1,
		RouteRPS:    1,
		RouteBurst:  1,
	})

	// First request allowed
	res1 := performRequest(r, "6.6.6.6", "/test")
	assert.Equal(t, 200, res1.Code)

	// Second blocked
	res2 := performRequest(r, "6.6.6.6", "/test")
	assert.Equal(t, 429, res2.Code)

	// Wait for refill
	time.Sleep(1100 * time.Millisecond)

	// Should be allowed again
	res3 := performRequest(r, "6.6.6.6", "/test")
	assert.Equal(t, 200, res3.Code)
}
