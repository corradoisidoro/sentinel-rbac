package handler_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	appErr "github.com/corradoisidoro/sentinel-rbac/internal/errors"
	"github.com/corradoisidoro/sentinel-rbac/internal/handler"
	"github.com/corradoisidoro/sentinel-rbac/internal/models"
	serviceMocks "github.com/corradoisidoro/sentinel-rbac/internal/service/mocks"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"gorm.io/gorm"
)

func setupRouter(h *handler.UserHandler) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.POST("/register", h.Register)
	r.POST("/login", h.Login)

	// Fake auth middleware for tests
	r.GET("/validate", func(c *gin.Context) {
		c.Set("user", gin.H{
			"id":    1,
			"email": "test@example.com",
			"role":  "user",
		})
		h.Validate(c)
	})

	return r
}

func TestRegisterHandler_Success(t *testing.T) {
	service := new(serviceMocks.UserServiceMock)
	h := handler.NewUserHandler(service)
	router := setupRouter(h)

	service.On(
		"Register",
		mock.Anything,
		"test@example.com",
		"password123",
	).Return(&models.User{
		Model: gorm.Model{ID: 1},
		Email: "test@example.com",
		Role:  "user",
	}, nil)

	body, _ := json.Marshal(gin.H{
		"email":    "test@example.com",
		"password": "password123",
	})

	req, _ := http.NewRequest(http.MethodPost, "/register", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	resp := httptest.NewRecorder()

	router.ServeHTTP(resp, req)

	assert.Equal(t, http.StatusCreated, resp.Code)
	assert.Contains(t, resp.Body.String(), "test@example.com")
}

func TestRegisterHandler_UserAlreadyExists(t *testing.T) {
	service := new(serviceMocks.UserServiceMock)
	h := handler.NewUserHandler(service)
	router := setupRouter(h)

	service.On(
		"Register",
		mock.Anything,
		"test@example.com",
		"password123",
	).Return(nil, appErr.ErrUserAlreadyExists)

	body, _ := json.Marshal(gin.H{
		"email":    "test@example.com",
		"password": "password123",
	})

	req, _ := http.NewRequest(http.MethodPost, "/register", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	resp := httptest.NewRecorder()

	router.ServeHTTP(resp, req)

	assert.Equal(t, http.StatusConflict, resp.Code)
}

func TestRegisterHandler_InvalidJSON(t *testing.T) {
	service := new(serviceMocks.UserServiceMock)
	h := handler.NewUserHandler(service)
	router := setupRouter(h)

	req, _ := http.NewRequest(http.MethodPost, "/register", bytes.NewBuffer([]byte(`{`)))
	req.Header.Set("Content-Type", "application/json")
	resp := httptest.NewRecorder()

	router.ServeHTTP(resp, req)

	assert.Equal(t, http.StatusBadRequest, resp.Code)
}

func TestLoginHandler_Success(t *testing.T) {
	service := new(serviceMocks.UserServiceMock)
	h := handler.NewUserHandler(service)
	router := setupRouter(h)

	service.On(
		"Login",
		mock.Anything,
		"test@example.com",
		"password123",
	).Return("jwt-token", nil)

	body, _ := json.Marshal(gin.H{
		"email":    "test@example.com",
		"password": "password123",
	})

	req, _ := http.NewRequest(http.MethodPost, "/login", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	resp := httptest.NewRecorder()

	router.ServeHTTP(resp, req)

	assert.Equal(t, http.StatusOK, resp.Code)
	assert.Contains(t, resp.Body.String(), "jwt-token")

	// ✅ Cookie is set
	cookies := resp.Result().Cookies()
	assert.Len(t, cookies, 1)
	assert.Equal(t, "Authorization", cookies[0].Name)
	assert.Equal(t, "jwt-token", cookies[0].Value)
}

func TestLoginHandler_UserNotFound(t *testing.T) {
	service := new(serviceMocks.UserServiceMock)
	h := handler.NewUserHandler(service)
	router := setupRouter(h)

	service.On(
		"Login",
		mock.Anything,
		"test@example.com",
		"password123",
	).Return("", appErr.ErrUserNotFound)

	body, _ := json.Marshal(gin.H{
		"email":    "test@example.com",
		"password": "password123",
	})

	req, _ := http.NewRequest(http.MethodPost, "/login", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	resp := httptest.NewRecorder()

	router.ServeHTTP(resp, req)

	assert.Equal(t, http.StatusUnauthorized, resp.Code)
}

func TestLoginHandler_InvalidPassword(t *testing.T) {
	service := new(serviceMocks.UserServiceMock)
	h := handler.NewUserHandler(service)
	router := setupRouter(h)

	service.On(
		"Login",
		mock.Anything,
		"test@example.com",
		"wrongpassword",
	).Return("", appErr.ErrInvalidPassword)

	body, _ := json.Marshal(gin.H{
		"email":    "test@example.com",
		"password": "wrongpassword",
	})

	req, _ := http.NewRequest(http.MethodPost, "/login", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	resp := httptest.NewRecorder()

	router.ServeHTTP(resp, req)

	assert.Equal(t, http.StatusUnauthorized, resp.Code)
}

func TestLoginHandler_InvalidJSON(t *testing.T) {
	service := new(serviceMocks.UserServiceMock)
	h := handler.NewUserHandler(service)
	router := setupRouter(h)

	req, _ := http.NewRequest(http.MethodPost, "/login", bytes.NewBuffer([]byte(`{`)))
	req.Header.Set("Content-Type", "application/json")
	resp := httptest.NewRecorder()

	router.ServeHTTP(resp, req)

	assert.Equal(t, http.StatusBadRequest, resp.Code)
}

func TestValidateHandler_Success(t *testing.T) {
	service := new(serviceMocks.UserServiceMock)
	h := handler.NewUserHandler(service)
	router := setupRouter(h)

	req, _ := http.NewRequest(http.MethodGet, "/validate", nil)
	resp := httptest.NewRecorder()

	router.ServeHTTP(resp, req)

	assert.Equal(t, http.StatusOK, resp.Code)
	assert.Contains(t, resp.Body.String(), "User is authenticated")
	assert.Contains(t, resp.Body.String(), "test@example.com")
}
