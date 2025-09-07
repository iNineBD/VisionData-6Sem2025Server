package middleware

import (
	"errors"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt"
)

// Gera token JWT
func GenerateJWT(userID int64, email string, role int64) (string, error) {
	jwtKey, err := LoadJWTKey()
	if err != nil {
		return "erro ao carregar JWT", err
	}
	claims := jwt.MapClaims{

		"user_id": userID,
		"email":   email,
		"role":    role,
		"exp":     time.Now().Add(1 * time.Hour).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(jwtKey))
}

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

func DecoteTokenJWT(token string) (jwt.MapClaims, error) {

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
func Auth() gin.HandlerFunc {
	return func(c *gin.Context) {
		token := c.GetHeader("Authorization")
		if token == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "JWT token not provided"})
			return
		}

		parts := strings.Split(token, " ")
		if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid Authorization header format"})
			return
		}

		token = parts[1]

		claims, err := DecoteTokenJWT(token)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
			return
		}

		c.Set("currentUser", claims)
		c.Next()
	}
}

func LoadJWTKey() (string, error) {
	return os.Getenv("JWT_SECRET"), nil
}
