package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/corradoisidoro/sentinel-rbac/internal/middleware"
	"github.com/corradoisidoro/sentinel-rbac/internal/models"
	"github.com/corradoisidoro/sentinel-rbac/internal/repository/mocks"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func TestRequireAuth_NoCookie(t *testing.T) {
	gin.SetMode(gin.TestMode)

	repo := new(mocks.UserRepositoryMock)
	m := middleware.NewAuthMiddleware([]byte("secret"), repo)

	r := gin.New()
	r.GET("/protected", m.RequireAuth, func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	req := httptest.NewRequest("GET", "/protected", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestRequireAuth_InvalidToken(t *testing.T) {
	gin.SetMode(gin.TestMode)

	repo := new(mocks.UserRepositoryMock)
	m := middleware.NewAuthMiddleware([]byte("secret"), repo)

	r := gin.New()
	r.GET("/protected", m.RequireAuth, func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	req := httptest.NewRequest("GET", "/protected", nil)
	req.AddCookie(&http.Cookie{Name: "Authorization", Value: "invalid"})

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestRequireAuth_ValidToken_UserLoaded(t *testing.T) {
	gin.SetMode(gin.TestMode)

	user := &models.User{
		Model: gorm.Model{ID: 1},
		Role:  "admin",
	}

	repo := new(mocks.UserRepositoryMock)
	repo.On("FindById", mock.Anything, uint(1)).Return(user, nil)

	m := middleware.NewAuthMiddleware([]byte("secret"), repo)

	// Create valid token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub": float64(1),
		"exp": time.Now().Add(time.Hour).Unix(),
	})
	tokenString, _ := token.SignedString([]byte("secret"))

	r := gin.New()
	r.GET("/protected", m.RequireAuth, func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	req := httptest.NewRequest("GET", "/protected", nil)
	req.AddCookie(&http.Cookie{Name: "Authorization", Value: tokenString})

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	repo.AssertExpectations(t)
}

func TestAuthorizeRole_NoUserInContext(t *testing.T) {
	gin.SetMode(gin.TestMode)

	repo := new(mocks.UserRepositoryMock)
	m := middleware.NewAuthMiddleware([]byte("secret"), repo)

	r := gin.New()
	r.GET("/admin", m.AuthorizeRole("admin"), func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	req := httptest.NewRequest("GET", "/admin", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestAuthorizeRole_WrongRole(t *testing.T) {
	gin.SetMode(gin.TestMode)

	repo := new(mocks.UserRepositoryMock)
	m := middleware.NewAuthMiddleware([]byte("secret"), repo)

	r := gin.New()
	r.GET("/admin",
		func(c *gin.Context) {
			c.Set("user", &models.User{Role: "user"})
		},
		m.AuthorizeRole("admin"),
		func(c *gin.Context) {
			c.Status(http.StatusOK)
		},
	)

	req := httptest.NewRequest("GET", "/admin", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusForbidden, w.Code)
}

func TestAuthorizeRole_Allowed(t *testing.T) {
	gin.SetMode(gin.TestMode)

	repo := new(mocks.UserRepositoryMock)
	m := middleware.NewAuthMiddleware([]byte("secret"), repo)

	r := gin.New()
	r.GET("/admin",
		func(c *gin.Context) {
			c.Set("user", &models.User{Role: "admin"})
		},
		m.AuthorizeRole("admin"),
		func(c *gin.Context) {
			c.Status(http.StatusOK)
		},
	)

	req := httptest.NewRequest("GET", "/admin", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
}
