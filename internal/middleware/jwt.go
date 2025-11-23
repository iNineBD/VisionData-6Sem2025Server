package middleware

import (
	"errors"
	"fmt"
	"net/http"
	"orderstreamrest/internal/models/dto"
	"os"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt"
)

// GenerateJWT generates a JWT token for a given user ID, email, and role
func GenerateJWT(userID int64, email string, role int64) (string, error) {
	jwtKey := os.Getenv("JWT_SECRET")
	claims := jwt.MapClaims{

		"user_id": userID,
		"email":   email,
		"role":    role,
		"exp":     time.Now().Add(1 * time.Hour).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(jwtKey))
}

// VerifyToken verifies a JWT token and returns the token if valid
func VerifyToken(token string) (*jwt.Token, error) {

	tokenVerify, err := jwt.Parse(token, func(newToken *jwt.Token) (any, error) {
		if _, isValid := newToken.Method.(*jwt.SigningMethodHMAC); !isValid {
			return nil, fmt.Errorf("unexpected signing method: %v", newToken.Header["alg"])
		}
		return []byte(os.Getenv("JWT_SECRET")), nil
	})
	if err != nil {
		err = errors.New("failed to verify token: " + err.Error())
		return nil, err
	}
	return tokenVerify, nil
}

// DecodeTokenJWT decodes a JWT token and returns the claims
func DecodeTokenJWT(token string) (jwt.MapClaims, error) {

	tokenVerify, err := VerifyToken(token)

	if err != nil {
		err = errors.New("failed to decode token " + err.Error())
		return nil, err
	}

	claims, isOk := tokenVerify.Claims.(jwt.MapClaims)

	if isOk && tokenVerify.Valid {
		return claims, nil
	}

	return nil, fmt.Errorf("invalid token")
}

// Auth is a middleware function that checks for a valid JWT token in the Authorization header
func Auth(minAccesScope int64) gin.HandlerFunc {
	return func(c *gin.Context) {
		token := c.GetHeader("Authorization")
		if token == "" {
			authError := dto.NewAuthErrorResponse(c, "Invalid token")
			c.AbortWithStatusJSON(http.StatusUnauthorized, authError)
			return
		}

		parts := strings.Split(token, " ")
		if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
			authError := dto.NewAuthErrorResponse(c, "Invalid token format. Use: Bearer <token>")
			c.AbortWithStatusJSON(http.StatusUnauthorized, authError)
			return
		}

		token = parts[1]

		claims, err := DecodeTokenJWT(token)
		if err != nil {
			authError := dto.NewAuthErrorResponse(c, "Invalid token")
			c.AbortWithStatusJSON(http.StatusUnauthorized, authError)
			return
		}

		if claims["role"] == nil {
			authError := dto.NewAuthErrorResponse(c, "Invalid token: missing role")
			c.AbortWithStatusJSON(http.StatusUnauthorized, authError)
			return
		}

		/*userRoleInt, ok := claims["role"].(int64)
		if !ok {
			userRoleFloatConv, okConv := claims["role"].(float64)
			if !okConv {
				authError := dto.NewAuthErrorResponse(c, "Invalid token: invalid role type")
				c.AbortWithStatusJSON(http.StatusUnauthorized, authError)
				return
			}
			userRoleInt = int64(userRoleFloatConv)
		}

		if userRoleInt > minAccesScope {
			authError := dto.NewAuthErrorResponse(c, "Insufficient permissions")
			c.AbortWithStatusJSON(http.StatusForbidden, authError)
			return
		}*/

		c.Set("currentUser", claims)
		c.Next()
	}
}
