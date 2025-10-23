package users

import (
	"net/http"
	"orderstreamrest/internal/config"
	"orderstreamrest/internal/middleware"
	"orderstreamrest/internal/models/dto"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

// Login autentica um usuário e retorna um JWT token
// @Summary      Login
// @Description  Autentica um usuário com email e senha e retorna um JWT token
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
func Login(cfg *config.App) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req dto.LoginRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, dto.ErrorResponse{
				BaseResponse: dto.BaseResponse{
					Success:   false,
					Timestamp: time.Now(),
				},
				Error:   "Bad Request",
				Code:    http.StatusBadRequest,
				Message: "Invalid request body",
				Details: err.Error(),
			})
			return
		}

		// Buscar usuário por email
		user, err := cfg.SqlServer.GetUserByEmail(c.Request.Context(), req.Email)
		if err != nil {
			c.JSON(http.StatusUnauthorized, dto.ErrorResponse{
				BaseResponse: dto.BaseResponse{
					Success:   false,
					Timestamp: time.Now(),
				},
				Error:   "Unauthorized",
				Code:    http.StatusUnauthorized,
				Message: "Invalid credentials",
			})
			return
		}

		// Verificar se usuário está ativo
		if !user.IsActive {
			c.JSON(http.StatusForbidden, dto.ErrorResponse{
				BaseResponse: dto.BaseResponse{
					Success:   false,
					Timestamp: time.Now(),
				},
				Error:   "Forbidden",
				Code:    http.StatusForbidden,
				Message: "User account is inactive",
			})
			return
		}

		// Verificar se usuário tem senha (não é apenas Microsoft Auth)
		if user.PasswordHash == nil {
			c.JSON(http.StatusBadRequest, dto.ErrorResponse{
				BaseResponse: dto.BaseResponse{
					Success:   false,
					Timestamp: time.Now(),
				},
				Error:   "Bad Request",
				Code:    http.StatusBadRequest,
				Message: "User uses Microsoft authentication. Please use Microsoft login",
			})
			return
		}

		// Verificar senha
		err = bcrypt.CompareHashAndPassword([]byte(*user.PasswordHash), []byte(req.Password))
		if err != nil {
			c.JSON(http.StatusUnauthorized, dto.ErrorResponse{
				BaseResponse: dto.BaseResponse{
					Success:   false,
					Timestamp: time.Now(),
				},
				Error:   "Unauthorized",
				Code:    http.StatusUnauthorized,
				Message: "Invalid credentials",
			})
			return
		}

		// Gerar JWT token
		token, err := middleware.GenerateJWT(int64(user.Id), user.Email, 1)
		if err != nil {
			c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
				BaseResponse: dto.BaseResponse{
					Success:   false,
					Timestamp: time.Now(),
				},
				Error:   "Internal Server Error",
				Code:    http.StatusInternalServerError,
				Message: "Failed to generate authentication token",
				Details: err.Error(),
			})
			return
		}

		// Atualizar LastLoginAt
		now := time.Now()
		user.LastLoginAt = &now
		if err := cfg.SqlServer.UpdateUser(c.Request.Context(), user.Id, user); err != nil {
			// Log error but don't fail the login
			// A falha em atualizar LastLoginAt não deve impedir o login
		}

		// Calcular tempo de expiração (1 hora a partir de agora)
		expiresAt := time.Now().Add(1 * time.Hour)

		c.JSON(http.StatusOK, dto.SuccessResponse{
			BaseResponse: dto.BaseResponse{
				Success:   true,
				Timestamp: time.Now(),
			},
			Data: dto.LoginResponse{
				Token:     token,
				TokenType: "Bearer",
				ExpiresIn: 3600, // segundos (1 hora)
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
