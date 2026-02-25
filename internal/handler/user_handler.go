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
	Role     string `json:"role" binding:"required,oneof=user admin"`
}

type loginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

func NewUserHandler(service service.UserService) *UserHandler {
	return &UserHandler{service: service}
}

// Register godoc
// @Summary Register a new user
// @Description Creates a new user account with email, password, and role.
// @Tags Auth
// @Accept json
// @Produce json
// @Param request body registrationRequest true "User registration payload"
// @Success 201 {object} map[string]interface{} "User created"
// @Failure 400 {object} map[string]string "Invalid input"
// @Failure 409 {object} map[string]string "User already exists"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /auth/register [post]
func (h *UserHandler) Register(c *gin.Context) {
	var req registrationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": appErr.ErrFailedToParseRequestBody.Error()})
		return
	}

	user, err := h.service.Register(c.Request.Context(), req.Email, req.Password, req.Role)
	if err != nil {
		switch err {
		case appErr.ErrUserAlreadyExists:
			c.JSON(http.StatusConflict, gin.H{"error": appErr.ErrUserAlreadyExists.Error()})
		case appErr.ErrInvalidInput:
			c.JSON(http.StatusBadRequest, gin.H{"error": appErr.ErrInvalidInput.Error()})
		case appErr.ErrInvalidRole:
			c.JSON(http.StatusBadRequest, gin.H{"error": appErr.ErrInvalidRole.Error()})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": appErr.ErrInternal.Error()})
		}

		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"id":    user.ID,
		"email": user.Email,
		"role":  user.Role,
	})
}

// Login godoc
// @Summary Login user
// @Description Authenticates a user and returns a JWT token.
// @Tags Auth
// @Accept json
// @Produce json
// @Param request body loginRequest true "Login payload"
// @Success 200 {object} map[string]interface{} "Login successful"
// @Failure 400 {object} map[string]string "Invalid request"
// @Failure 401 {object} map[string]string "Invalid credentials"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /auth/login [post]
func (h *UserHandler) Login(c *gin.Context) {
	var req loginRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": appErr.ErrFailedToParseRequestBody.Error()})
		return
	}

	tokenString, err := h.service.Login(c.Request.Context(), req.Email, req.Password)
	if err != nil {
		switch err {
		case appErr.ErrUserNotFound, appErr.ErrInvalidPassword:
			c.JSON(http.StatusUnauthorized, gin.H{"error": appErr.ErrInvalidPassword.Error()})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": appErr.ErrInternal.Error()})
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

// Profile godoc
// @Summary Get authenticated user profile
// @Description Returns the authenticated user's profile.
// @Tags Users
// @Produce json
// @Success 200 {object} map[string]interface{} "User profile"
// @Router /users/profile [get]
func (h *UserHandler) Profile(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"message": "User is authenticated",
		"user":    c.MustGet("user")})
}

// Logout godoc
// @Summary Logout user
// @Description Clears the authentication cookie.
// @Tags Auth
// @Produce json
// @Success 200 {object} map[string]string "Logged out"
// @Router /auth/logout [post]
func (h *UserHandler) Logout(c *gin.Context) {
	c.SetCookie("Authorization", "", -1, "/", "", false, true)
	c.JSON(http.StatusOK, gin.H{
		"message": "logged out successfully",
	})
}

// Admin godoc
// @Summary Admin-only endpoint
// @Description Accessible only to users with the admin role.
// @Tags Users
// @Produce json
// @Success 200 {object} map[string]string "Admin dashboard"
// @Failure 403 {object} map[string]string "Forbidden"
// @Router /users/admin [get]
func (h *UserHandler) Admin(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"message": "Welcome to the admin dashboard",
	})
}
