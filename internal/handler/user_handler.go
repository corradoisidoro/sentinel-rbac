package handler

import (
	"net/http"
	"time"

	appErr "github.com/corradoisidoro/sentinel-rbac/internal/errors"
	"github.com/corradoisidoro/sentinel-rbac/internal/service"
	"github.com/gin-gonic/gin"
)

type UserHandler struct {
	service service.UserService
}

type registrationRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=8"`
}

type loginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

func NewUserHandler(service service.UserService) *UserHandler {
	return &UserHandler{service: service}
}

func (h *UserHandler) Register(c *gin.Context) {
	var req registrationRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": appErr.ErrFailedToParseRequestBody})
		return
	}

	user, err := h.service.Register(c.Request.Context(), req.Email, req.Password)
	if err != nil {
		switch err {
		case appErr.ErrUserAlreadyExists:
			c.JSON(http.StatusConflict, gin.H{"error": appErr.ErrUserAlreadyExists})
		case appErr.ErrInvalidInput:
			c.JSON(http.StatusBadRequest, gin.H{"error": appErr.ErrInvalidInput})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": appErr.ErrInternal})
		}

		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"id":    user.ID,
		"email": user.Email,
		"role":  user.Role,
	})
}

func (h *UserHandler) Login(c *gin.Context) {
	var req loginRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": appErr.ErrFailedToParseRequestBody})
		return
	}

	tokenString, err := h.service.Login(c.Request.Context(), req.Email, req.Password)
	if err != nil {
		switch err {
		case appErr.ErrUserNotFound, appErr.ErrInvalidPassword:
			c.JSON(http.StatusUnauthorized, gin.H{"error": appErr.ErrInvalidPassword})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": appErr.ErrInternal})
		}
		return
	}

	const thirtyDays = 30 * 24 * time.Hour
	c.SetSameSite(http.SameSiteLaxMode)
	c.SetCookie("Authorization", tokenString, int(thirtyDays.Seconds()), "/", "", false, true)

	c.JSON(http.StatusOK, gin.H{
		"message": "logged in successfully",
		"token":   tokenString,
	})
}

func (h *UserHandler) Validate(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"message": "User is authenticated",
		"user":    c.MustGet("user")})
}

func (h *UserHandler) Logout(c *gin.Context) {
	c.SetCookie("Authorization", "", -1, "/", "", false, true)
	c.JSON(http.StatusOK, gin.H{
		"message": "logged out successfully",
	})
}
