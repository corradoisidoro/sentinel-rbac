package middleware

import (
	"fmt"
	"net/http"

	"github.com/corradoisidoro/sentinel-rbac/internal/errors"
	"github.com/corradoisidoro/sentinel-rbac/internal/repository"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt"
)

type AuthMiddleware struct {
	secret []byte
	repo   repository.UserRepository
}

func NewAuthMiddleware(secret []byte, repo repository.UserRepository) *AuthMiddleware {
	return &AuthMiddleware{
		secret: secret,
		repo:   repo,
	}
}

func (m *AuthMiddleware) RequireAuth(c *gin.Context) {
	tokenString, err := c.Cookie("Authorization")
	if err != nil {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": errors.ErrNoTokenProvided.Error()})
		return
	}

	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (any, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])

		}
		return m.secret, nil
	})

	if err != nil || !token.Valid {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": errors.ErrTokenExpiredOrInvalid.Error()})
		return
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok {
		// jwt-go/v5 handles 'exp' validation automatically in Parse, but we need to extract 'sub'
		sub, ok := claims["sub"].(float64)
		if !ok {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": errors.ErrInvalidTokenSubject.Error()})
			return
		}

		user, err := m.repo.FindById(c.Request.Context(), uint(sub))
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": errors.ErrUserNotFound.Error()})
			return
		}

		c.Set("user", user)
		c.Next()
	} else {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": errors.ErrInvalidClaims.Error()})
	}
}
