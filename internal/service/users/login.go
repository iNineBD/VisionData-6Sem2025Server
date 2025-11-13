package users

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"math/big"
	"net/http"
	"net/url"
	"orderstreamrest/internal/config"
	"orderstreamrest/internal/middleware"
	"orderstreamrest/internal/models/dto"
	"orderstreamrest/internal/models/entities"
	"orderstreamrest/internal/utils"
	"os"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
	"golang.org/x/oauth2"
)

var microsoftOauthConfig = &oauth2.Config{
	ClientID:     os.Getenv("MICROSOFT_CLIENT_ID"),
	ClientSecret: os.Getenv("MICROSOFT_CLIENT_SECRET"),
	RedirectURL:  "http://localhost:8080/auth/microsoft/callback",
	Scopes: []string{
		"openid",
		"profile",
		"email",
	},
	Endpoint: oauth2.Endpoint{
		AuthURL:  "https://login.microsoftonline.com/common/oauth2/v2.0/authorize",
		TokenURL: "https://login.microsoftonline.com/common/oauth2/v2.0/token",
	},
}

// LoginHandler autentica um usuário e retorna um JWT token
// @Summary      Login
// @Description  Endpoint unificado para autenticação. Aceita login tradicional (email/senha) ou login via Microsoft (id_token).
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        credentials body dto.LoginRequest true "Credenciais de login"
// @Success      200 {object} dto.SuccessResponse{data=dto.LoginResponse}
// @Failure      400 {object} dto.ErrorResponse "Bad Request - Dados inválidos"
// @Failure      401 {object} dto.ErrorResponse "Unauthorized - Credenciais inválidas"
// @Failure      403 {object} dto.ErrorResponse "Forbidden - Usuário inativo"
// @Failure      500 {object} dto.ErrorResponse "Internal Server Error"
// @Router       /auth/login [post]
func LoginHandler(a *config.App) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req dto.LoginRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, dto.ErrorResponse{
				BaseResponse: dto.BaseResponse{Success: false, Timestamp: time.Now()},
				Error:        "Bad Request",
				Code:         http.StatusBadRequest,
				Message:      "Invalid request body",
				Details:      err.Error(),
			})
			return
		}

		ctx := c.Request.Context()

		var user *entities.User
		var err error

		switch req.LoginType {
		case "password":
			// Existing password flow
			user, err = a.SqlServer.GetUserByEmail(ctx, req.Email)
			if err != nil {
				c.JSON(http.StatusUnauthorized, dto.ErrorResponse{
					BaseResponse: dto.BaseResponse{Success: false, Timestamp: time.Now()},
					Error:        "Unauthorized",
					Code:         http.StatusUnauthorized,
					Message:      "Invalid credentials",
				})
				return
			}
			if !user.IsActive {
				c.JSON(http.StatusForbidden, dto.ErrorResponse{
					BaseResponse: dto.BaseResponse{Success: false, Timestamp: time.Now()},
					Error:        "Forbidden",
					Code:         http.StatusForbidden,
					Message:      "User account is inactive",
				})
				return
			}
			if user.PasswordHash == nil {
				c.JSON(http.StatusBadRequest, dto.ErrorResponse{
					BaseResponse: dto.BaseResponse{Success: false, Timestamp: time.Now()},
					Error:        "Bad Request",
					Code:         http.StatusBadRequest,
					Message:      "User uses Microsoft authentication. Please use Microsoft login",
				})
				return
			}
			if err := bcrypt.CompareHashAndPassword([]byte(*user.PasswordHash), []byte(req.Password)); err != nil {
				c.JSON(http.StatusUnauthorized, dto.ErrorResponse{
					BaseResponse: dto.BaseResponse{Success: false, Timestamp: time.Now()},
					Error:        "Unauthorized",
					Code:         http.StatusUnauthorized,
					Message:      "Invalid credentials",
				})
				return
			}

		case "microsoft":
			// Two possibilities:
			// 1) Your backend did the OAuth flow and you already have the id_token -> the front POSTs it here.
			// 2) Or front didn't touch MS and this endpoint will be used only when you want front to POST id_token.
			// In our preferred backend-only flow, the backend handles entire OAuth and you won't use this branch.
			if req.MicrosoftIDToken == "" {
				c.JSON(http.StatusBadRequest, dto.ErrorResponse{
					BaseResponse: dto.BaseResponse{Success: false, Timestamp: time.Now()},
					Error:        "Bad Request",
					Code:         http.StatusBadRequest,
					Message:      "missing microsoft_id_token",
				})
				return
			}

			claims, err := validateMicrosoftIDToken(ctx, req.MicrosoftIDToken)
			if err != nil {
				c.JSON(http.StatusUnauthorized, dto.ErrorResponse{
					BaseResponse: dto.BaseResponse{Success: false, Timestamp: time.Now()},
					Error:        "Unauthorized",
					Code:         http.StatusUnauthorized,
					Message:      fmt.Sprintf("Invalid Microsoft token: %v", err),
				})
				return
			}

			// Try find by Microsoft subject first, then by email
			user, err = a.SqlServer.GetUserByMicrosoftID(ctx, claims.Subject)
			if err != nil {
				user, err = a.SqlServer.GetUserByEmail(ctx, claims.Email)
				if err != nil {
					// user not found -> create new user linked to this Microsoft subject
					newUser := &entities.User{
						Name:        claims.Name,
						Email:       claims.Email,
						UserType:    utils.UserTypMapIntToStr[3], // default type
						IsActive:    true,
						CreatedAt:   time.Now(),
						MicrosoftId: &claims.Subject,
					}
					if _, err := a.SqlServer.CreateUser(ctx, newUser); err != nil {
						c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
							BaseResponse: dto.BaseResponse{Success: false, Timestamp: time.Now()},
							Error:        "Internal Server Error",
							Code:         http.StatusInternalServerError,
							Message:      "failed to create user",
							Details:      err.Error(),
						})
						return
					}
					user = newUser
				} else {
					// user exists by email but not linked to Microsoft -> link it
					user.MicrosoftId = &claims.Subject
					now := time.Now()
					user.UpdatedAt = &now
					if err := a.SqlServer.UpdateUser(ctx, user.Id, user); err != nil {
						log.Printf("warning: failed to link microsoft id for user %d: %v", user.Id, err)
					}
				}
			}

			if !user.IsActive {
				c.JSON(http.StatusForbidden, dto.ErrorResponse{
					BaseResponse: dto.BaseResponse{Success: false, Timestamp: time.Now()},
					Error:        "Forbidden",
					Code:         http.StatusForbidden,
					Message:      "User account is inactive",
				})
				return
			}

		default:
			c.JSON(http.StatusBadRequest, dto.ErrorResponse{
				BaseResponse: dto.BaseResponse{Success: false, Timestamp: time.Now()},
				Error:        "Bad Request",
				Code:         http.StatusBadRequest,
				Message:      "Invalid login_type",
			})
			return
		}

		// Generate internal JWT
		token, err := middleware.GenerateJWT(int64(user.Id), user.Email, int64(utils.UserTypMapStrToInt[user.UserType]))
		if err != nil {
			c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
				BaseResponse: dto.BaseResponse{Success: false, Timestamp: time.Now()},
				Error:        "Internal Server Error",
				Code:         http.StatusInternalServerError,
				Message:      "Failed to generate authentication token",
				Details:      err.Error(),
			})
			return
		}

		// Update LastLoginAt (non-blocking)
		now := time.Now()
		user.LastLoginAt = &now
		if err := a.SqlServer.UpdateUser(ctx, user.Id, user); err != nil {
			log.Printf("failed to update LastLoginAt for user %d: %v", user.Id, err)
		}

		expiresAt := time.Now().Add(1 * time.Hour)

		c.JSON(http.StatusOK, dto.SuccessResponse{
			BaseResponse: dto.BaseResponse{Success: true, Timestamp: time.Now()},
			Data: dto.LoginResponse{
				Token:     token,
				TokenType: "Bearer",
				ExpiresIn: int(time.Until(expiresAt).Seconds()),
				ExpiresAt: expiresAt,
				User: dto.UserResponse{
					Id:          user.Id,
					Name:        user.Name,
					Email:       user.Email,
					UserType:    user.UserType,
					MicrosoftId: user.MicrosoftId,
					IsActive:    user.IsActive,
					CreatedAt:   user.CreatedAt,
					UpdatedAt:   user.UpdatedAt,
					LastLoginAt: user.LastLoginAt,
				},
			},
			Message: "Login successful",
		})
	}
}

const microsoftJWKSURL = "https://login.microsoftonline.com/common/discovery/v2.0/keys"

type MicrosoftClaims struct {
	Email   string `json:"email"`
	Name    string `json:"name"`
	Subject string `json:"sub"`
	jwt.RegisteredClaims
}

type jwksResponse struct {
	Keys []struct {
		Kty string `json:"kty"`
		N   string `json:"n"`
		E   string `json:"e"`
		Alg string `json:"alg"`
		Kid string `json:"kid"`
		Use string `json:"use"`
	} `json:"keys"`
}

// contains helper for audience
func containsAudience(aud jwt.ClaimStrings, want string) bool {
	for _, a := range aud {
		if a == want {
			return true
		}
	}
	return false
}

func validateMicrosoftIDToken(ctx context.Context, idToken string) (*MicrosoftClaims, error) {
	// split token header to find kid
	parts := strings.Split(idToken, ".")
	if len(parts) < 2 {
		return nil, errors.New("invalid jwt format")
	}
	headerBytes, err := base64.RawURLEncoding.DecodeString(parts[0])
	if err != nil {
		return nil, fmt.Errorf("decode header: %w", err)
	}
	var header map[string]interface{}
	if err := json.Unmarshal(headerBytes, &header); err != nil {
		return nil, fmt.Errorf("parse header: %w", err)
	}
	kid, ok := header["kid"].(string)
	if !ok || kid == "" {
		return nil, errors.New("kid not present in token header")
	}

	// fetch JWKS
	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, microsoftJWKSURL, nil)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("fetch jwks: %w", err)
	}
	defer resp.Body.Close()

	var jwks jwksResponse
	if err := json.NewDecoder(resp.Body).Decode(&jwks); err != nil {
		return nil, fmt.Errorf("decode jwks: %w", err)
	}

	// find matching key
	var chosen struct {
		Kty, N, E, Alg, Kid, Use string
	}
	found := false
	for _, k := range jwks.Keys {
		if k.Kid == kid {
			chosen.Kty = k.Kty
			chosen.N = k.N
			chosen.E = k.E
			chosen.Alg = k.Alg
			chosen.Kid = k.Kid
			chosen.Use = k.Use
			found = true
			break
		}
	}
	if !found {
		return nil, fmt.Errorf("no jwk found for kid=%s", kid)
	}

	// build RSA public key
	nb, err := base64.RawURLEncoding.DecodeString(chosen.N)
	if err != nil {
		return nil, fmt.Errorf("decode N: %w", err)
	}
	eb, err := base64.RawURLEncoding.DecodeString(chosen.E)
	if err != nil {
		return nil, fmt.Errorf("decode E: %w", err)
	}
	e := 0
	for _, b := range eb {
		e = e<<8 + int(b)
	}
	pub := &rsa.PublicKey{
		N: new(big.Int).SetBytes(nb),
		E: e,
	}

	// parse and validate signature
	token, err := jwt.ParseWithClaims(idToken, &MicrosoftClaims{}, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return pub, nil
	})
	if err != nil {
		return nil, fmt.Errorf("parse/verify token: %w", err)
	}

	claims, ok := token.Claims.(*MicrosoftClaims)
	if !ok || !token.Valid {
		return nil, errors.New("invalid token or claims")
	}

	// ---------- explicit validations ----------
	now := time.Now()

	// expiration
	if claims.ExpiresAt == nil {
		return nil, errors.New("token missing exp")
	}
	if claims.ExpiresAt.Time.Before(now) {
		return nil, errors.New("token expired")
	}

	// audience (must match your client id)
	clientID := os.Getenv("MICROSOFT_CLIENT_ID")
	if clientID != "" {
		if !containsAudience(claims.Audience, clientID) {
			return nil, fmt.Errorf("token aud mismatch (want %s)", clientID)
		}
	}

	// issuer (basic check)
	if !strings.HasPrefix(claims.Issuer, "https://login.microsoftonline.com/") &&
		!strings.HasPrefix(claims.Issuer, "https://sts.windows.net/") {
		return nil, fmt.Errorf("unexpected issuer: %s", claims.Issuer)
	}

	// nbf / iat (optional strict checks)
	if claims.NotBefore != nil && claims.NotBefore.Time.After(now.Add(1*time.Minute)) {
		return nil, errors.New("token not valid yet (nbf)")
	}
	// ------------------------------------------------

	return claims, nil
}

// ---------------------------
// OAuth2 config & helpers
// ---------------------------

var oauthConfig *oauth2.Config

func InitOAuthConfig() {
	clientID := os.Getenv("MICROSOFT_CLIENT_ID")
	clientSecret := os.Getenv("MICROSOFT_CLIENT_SECRET")
	redirectURL := os.Getenv("REDIRECT_URL") // e.g. http://localhost:8080/auth/microsoft/callback

	if clientID == "" || clientSecret == "" || redirectURL == "" {
		log.Fatal("MICROSOFT_CLIENT_ID, MICROSOFT_CLIENT_SECRET and REDIRECT_URL must be set")
	}

	oauthConfig = &oauth2.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		RedirectURL:  redirectURL,
		Scopes:       []string{"openid", "profile", "email"},
		Endpoint: oauth2.Endpoint{
			AuthURL:  "https://login.microsoftonline.com/common/oauth2/v2.0/authorize",
			TokenURL: "https://login.microsoftonline.com/common/oauth2/v2.0/token",
		},
	}
}

// generateState param (for CSRF protection) - in production save the state in a session or DB
func generateState() (string, error) {
	b := make([]byte, 32)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}

// MicrosoftLoginHandler inicia o fluxo OAuth2 da Microsoft
// @Summary      Iniciar login via Microsoft
// @Description  Redireciona o usuário para o portal de autenticação da Microsoft (OAuth2). Esse endpoint não requer body e deve ser acessado via navegador.
// @Tags         auth
// @Produce      json
// @Success      302 {string} string "Redirect para a página de login da Microsoft"
// @Failure      500 {object} dto.ErrorResponse "Internal Server Error - Falha ao gerar estado ou construir URL"
// @Router       /auth/microsoft/login [get]
func MicrosoftLoginHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		state, err := generateState()
		if err != nil {
			c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
				BaseResponse: dto.BaseResponse{Success: false, Timestamp: time.Now()},
				Error:        "Internal Server Error",
				Code:         http.StatusInternalServerError,
				Message:      "failed to generate state",
			})
			return
		}
		url := microsoftOauthConfig.AuthCodeURL(state, oauth2.AccessTypeOffline)
		c.Redirect(http.StatusFound, url)
	}
}

// MicrosoftCallbackHandler recebe o código OAuth2 da Microsoft e gera um JWT interno
// @Summary      Callback de autenticação Microsoft
// @Description  Endpoint que recebe o `code` da Microsoft após autenticação, valida o `id_token`, cria ou atualiza o usuário no banco e retorna um JWT interno.
// @Tags         auth
// @Produce      json
// @Param        code query string true "Código de autorização retornado pela Microsoft"
// @Success      302 {string} string "Redirect para o frontend com o JWT na query string"
// @Failure      400 {object} dto.ErrorResponse "Bad Request - Código ausente"
// @Failure      401 {object} dto.ErrorResponse "Unauthorized - Token Microsoft inválido"
// @Failure      500 {object} dto.ErrorResponse "Internal Server Error - Falha ao trocar o código ou gerar JWT"
// @Router       /auth/microsoft/callback [get]
func MicrosoftCallbackHandler(cfg *config.App) gin.HandlerFunc {
	return func(c *gin.Context) {
		code := c.Query("code")
		if code == "" {
			c.JSON(http.StatusBadRequest, dto.ErrorResponse{
				BaseResponse: dto.BaseResponse{Success: false, Timestamp: time.Now()},
				Error:        "Bad Request",
				Code:         http.StatusBadRequest,
				Message:      "missing code",
			})
			return
		}

		// Troca o code por token Microsoft
		token, err := microsoftOauthConfig.Exchange(context.Background(), code)
		if err != nil {
			c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
				BaseResponse: dto.BaseResponse{Success: false, Timestamp: time.Now()},
				Error:        "Internal Server Error",
				Code:         http.StatusInternalServerError,
				Message:      "failed to exchange code for token",
				Details:      err.Error(),
			})
			return
		}

		rawIDToken := token.Extra("id_token")
		if rawIDToken == nil {
			c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
				BaseResponse: dto.BaseResponse{Success: false, Timestamp: time.Now()},
				Error:        "Internal Server Error",
				Code:         http.StatusInternalServerError,
				Message:      "id_token not found in token response",
			})
			return
		}

		idToken := rawIDToken.(string)
		claims, err := validateMicrosoftIDToken(context.Background(), idToken)
		if err != nil {
			c.JSON(http.StatusUnauthorized, dto.ErrorResponse{
				BaseResponse: dto.BaseResponse{Success: false, Timestamp: time.Now()},
				Error:        "Unauthorized",
				Code:         http.StatusUnauthorized,
				Message:      fmt.Sprintf("invalid microsoft id_token: %v", err),
			})
			return
		}

		// Busca ou cria usuário no banco
		user, err := cfg.SqlServer.GetUserByMicrosoftID(c.Request.Context(), claims.Subject)
		if err != nil {
			user, err = cfg.SqlServer.GetUserByEmail(c.Request.Context(), claims.Email)
			if err != nil {
				newUser := &entities.User{
					Name:        claims.Name,
					Email:       claims.Email,
					IsActive:    true,
					MicrosoftId: &claims.Subject,
					CreatedAt:   time.Now(),
					UserType:    utils.UserTypMapIntToStr[3],
				}
				if _, err := cfg.SqlServer.CreateUser(c.Request.Context(), newUser); err != nil {
					c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
						BaseResponse: dto.BaseResponse{Success: false, Timestamp: time.Now()},
						Error:        "Internal Server Error",
						Code:         http.StatusInternalServerError,
						Message:      "failed to create user",
						Details:      err.Error(),
					})
					return
				}
				user = newUser
			}
		}

		// Gera JWT interno
		tokenStr, err := middleware.GenerateJWT(int64(user.Id), user.Email, int64(utils.UserTypMapStrToInt[user.UserType]))
		if err != nil {
			c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
				BaseResponse: dto.BaseResponse{Success: false, Timestamp: time.Now()},
				Error:        "Internal Server Error",
				Code:         http.StatusInternalServerError,
				Message:      "failed to generate jwt",
				Details:      err.Error(),
			})
			return
		}

		// Redireciona para o front-end com o JWT
		redirectURL := fmt.Sprintf("%s?token=%s&id=%d&email=%s&name=%s&role=%s",
			os.Getenv("URL_REDIRECT_FRONT"),
			tokenStr,
			user.Id,
			url.QueryEscape(user.Email),
			url.QueryEscape(user.Name),
			url.QueryEscape(user.UserType),
		)

		c.Redirect(http.StatusFound, redirectURL)
	}
}
