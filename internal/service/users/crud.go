package users

import (
	"fmt"
	"net/http"
	"orderstreamrest/internal/config"
	"orderstreamrest/internal/models/dto"
	"orderstreamrest/internal/models/entities"
	"orderstreamrest/internal/utils"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

// CreateUser cria um novo usuário
// @Summary      Registrar Novo Usuário
// @Description  Cria um novo usuário no sistema (endpoint público para registro)
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        user body dto.CreateUserRequest true "Dados do usuário"
// @Success      201 {object} dto.SuccessResponse{data=dto.UserCreatedResponse}
// @Failure 	 400 {object} dto.ErrorResponse "Bad Request"
// @Failure 	 409 {object} dto.ErrorResponse "Conflict - Email já existe"
// @Failure 	 500 {object} dto.ErrorResponse "Internal Server Error"
// @Router       /auth/register [post]
func CreateUser(cfg *config.App) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req dto.CreateUserRequest
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

		if _, ok := utils.UserTypMapStrToInt[req.UserType]; !ok {
			c.JSON(http.StatusBadRequest, dto.ErrorResponse{
				BaseResponse: dto.BaseResponse{
					Success:   false,
					Timestamp: time.Now(),
				},
				Error:   "Bad Request",
				Code:    http.StatusBadRequest,
				Message: "Invalid parameter",
				Details: fmt.Errorf("the parameter 'userType' must be %v", utils.UserTypMapIntToStr),
			})
			return
		}

		// Validar que pelo menos senha foi fornecido
		if req.Password == nil {
			c.JSON(http.StatusBadRequest, dto.ErrorResponse{
				BaseResponse: dto.BaseResponse{
					Success:   false,
					Timestamp: time.Now(),
				},
				Error:   "Bad Request",
				Code:    http.StatusBadRequest,
				Message: "Either password or microsoftId must be provided",
			})
			return
		}

		// Validar consentimento dos termos
		if req.TermConsent.TermId == 0 || len(req.TermConsent.ItemConsents) == 0 {
			c.JSON(http.StatusBadRequest, dto.ErrorResponse{
				BaseResponse: dto.BaseResponse{
					Success:   false,
					Timestamp: time.Now(),
				},
				Error:   "Bad Request",
				Code:    http.StatusBadRequest,
				Message: "Term consent is required for registration",
			})
			return
		}

		// Verificar se o termo existe e está ativo
		term, err := cfg.SqlServer.GetTermByID(c.Request.Context(), req.TermConsent.TermId)
		if err != nil || !term.IsActive {
			c.JSON(http.StatusBadRequest, dto.ErrorResponse{
				BaseResponse: dto.BaseResponse{
					Success:   false,
					Timestamp: time.Now(),
				},
				Error:   "Bad Request",
				Code:    http.StatusBadRequest,
				Message: "Invalid or inactive term",
			})
			return
		}

		// Validar que o item obrigatório foi aceito
		hasAcceptedMandatory := false
		for _, itemConsent := range req.TermConsent.ItemConsents {
			// Buscar o item no termo
			for _, termItem := range term.Items {
				if termItem.Id == itemConsent.ItemId && termItem.IsMandatory && itemConsent.Accepted {
					hasAcceptedMandatory = true
					break
				}
			}
		}

		if !hasAcceptedMandatory {
			c.JSON(http.StatusBadRequest, dto.ErrorResponse{
				BaseResponse: dto.BaseResponse{
					Success:   false,
					Timestamp: time.Now(),
				},
				Error:   "Bad Request",
				Code:    http.StatusBadRequest,
				Message: "You must accept the mandatory term item to register",
			})
			return
		}

		// Verificar se email já existe
		existingUser, _ := cfg.SqlServer.GetUserByEmail(c.Request.Context(), req.Email)
		if existingUser != nil {
			c.JSON(http.StatusConflict, dto.ErrorResponse{
				BaseResponse: dto.BaseResponse{
					Success:   false,
					Timestamp: time.Now(),
				},
				Error:   "Conflict",
				Code:    http.StatusConflict,
				Message: "Email already exists",
			})
			return
		}

		// Hash da senha se fornecida
		var passwordHash *string
		if req.Password != nil {
			hash, err := bcrypt.GenerateFromPassword([]byte(*req.Password), bcrypt.DefaultCost)
			if err != nil {
				c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
					BaseResponse: dto.BaseResponse{
						Success:   false,
						Timestamp: time.Now(),
					},
					Error:   "Internal Server Error",
					Code:    http.StatusInternalServerError,
					Message: "Failed to hash password",
					Details: err.Error(),
				})
				return
			}
			hashStr := string(hash)
			passwordHash = &hashStr
		}

		// // Pegar ID do usuário autenticado (assumindo que está no contexto)
		// currentUserId, _ := c.Get("user_id")
		// var createdBy *int
		// if id, ok := currentUserId.(int); ok {
		// 	createdBy = &id
		// }

		temp := "pegadinha do malandro" + uuid.New().String()
		user := &entities.User{
			Name:         req.Name,
			Email:        req.Email,
			PasswordHash: passwordHash,
			UserType:     req.UserType,
			MicrosoftId:  &temp,
			IsActive:     true,
			// CreatedBy:    createdBy,
		}

		id, err := cfg.SqlServer.CreateUser(c.Request.Context(), user)
		if err != nil {
			c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
				BaseResponse: dto.BaseResponse{
					Success:   false,
					Timestamp: time.Now(),
				},
				Error:   "Internal Server Error",
				Code:    http.StatusInternalServerError,
				Message: "Failed to create user",
				Details: err.Error(),
			})
			return
		}

		// Registrar consentimento dos termos
		ipAddress := c.ClientIP()
		userAgent := c.GetHeader("User-Agent")

		consent := &entities.UserTermConsent{
			UserId:       id,
			TermId:       req.TermConsent.TermId,
			ConsentDate:  time.Now(),
			IsActive:     true,
			IPAddress:    &ipAddress,
			UserAgent:    &userAgent,
			ItemConsents: []entities.UserItemConsent{},
		}

		for _, itemConsent := range req.TermConsent.ItemConsents {
			consent.ItemConsents = append(consent.ItemConsents, entities.UserItemConsent{
				ItemId:      itemConsent.ItemId,
				Accepted:    itemConsent.Accepted,
				ConsentDate: time.Now(),
			})
		}

		err = cfg.SqlServer.RegisterUserConsent(c.Request.Context(), consent)
		if err != nil {
			// Se falhar ao registrar consentimento, reverter criação do usuário
			// Aqui você pode implementar uma lógica de rollback se necessário
			c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
				BaseResponse: dto.BaseResponse{
					Success:   false,
					Timestamp: time.Now(),
				},
				Error:   "Internal Server Error",
				Code:    http.StatusInternalServerError,
				Message: "Failed to register term consent",
				Details: err.Error(),
			})
			return
		}

		c.JSON(http.StatusCreated, dto.SuccessResponse{
			BaseResponse: dto.BaseResponse{
				Success:   true,
				Timestamp: time.Now(),
			},
			Data: dto.UserCreatedResponse{
				Id:      id,
				Message: "User created successfully",
			},
			Message: "User created successfully",
		})
	}
}

// GetUser busca um usuário por ID
// @Summary      Buscar Usuário
// @Description  Retorna um usuário específico por ID
// @Tags         users
// @Accept       json
// @Produce      json
// @Security 	 BearerAuth
// @Param        id path int true "ID do usuário"
// @Success      200 {object} dto.SuccessResponse{data=dto.UserResponse}
// @Failure 	 400 {object} dto.ErrorResponse "Bad Request"
// @Failure 	 401 {object} dto.AuthErrorResponse "Unauthorized"
// @Failure 	 404 {object} dto.ErrorResponse "Not Found"
// @Failure 	 500 {object} dto.ErrorResponse "Internal Server Error"
// @Router       /users/{id} [get]
func GetUser(cfg *config.App) gin.HandlerFunc {
	return func(c *gin.Context) {
		idParam := c.Param("id")
		id, err := strconv.Atoi(idParam)
		if err != nil {
			c.JSON(http.StatusBadRequest, dto.ErrorResponse{
				BaseResponse: dto.BaseResponse{
					Success:   false,
					Timestamp: time.Now(),
				},
				Error:   "Bad Request",
				Code:    http.StatusBadRequest,
				Message: "Invalid user ID",
			})
			return
		}

		user, err := cfg.SqlServer.GetUserByID(c.Request.Context(), id)
		if err != nil {
			c.JSON(http.StatusNotFound, dto.ErrorResponse{
				BaseResponse: dto.BaseResponse{
					Success:   false,
					Timestamp: time.Now(),
				},
				Error:   "Not Found",
				Code:    http.StatusNotFound,
				Message: "User not found",
				Details: err.Error(),
			})
			return
		}

		response := dto.UserResponse{
			Id:          user.Id,
			Name:        user.Name,
			Email:       user.Email,
			UserType:    user.UserType,
			MicrosoftId: user.MicrosoftId,
			IsActive:    user.IsActive,
			CreatedAt:   user.CreatedAt,
			UpdatedAt:   user.UpdatedAt,
			LastLoginAt: user.LastLoginAt,
		}

		c.JSON(http.StatusOK, dto.SuccessResponse{
			BaseResponse: dto.BaseResponse{
				Success:   true,
				Timestamp: time.Now(),
			},
			Data:    response,
			Message: "User retrieved successfully",
		})
	}
}

// GetAllUsers lista todos os usuários com paginação
// @Summary      Listar Usuários
// @Description  Retorna lista de usuários com paginação
// @Tags         users
// @Accept       json
// @Produce      json
// @Security 	 BearerAuth
// @Param        page query int false "Número da página" default(1)
// @Param        pageSize query int false "Tamanho da página" default(10)
// @Param        onlyActive query bool false "Apenas usuários ativos" default(false)
// @Success      200 {object} dto.SuccessResponse{data=dto.UsersListResponse}
// @Failure 	 400 {object} dto.ErrorResponse "Bad Request"
// @Failure 	 401 {object} dto.AuthErrorResponse "Unauthorized"
// @Failure 	 500 {object} dto.ErrorResponse "Internal Server Error"
// @Router       /users [get]
func GetAllUsers(cfg *config.App) gin.HandlerFunc {
	return func(c *gin.Context) {
		page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
		pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "10"))
		onlyActive, _ := strconv.ParseBool(c.DefaultQuery("onlyActive", "false"))

		if page < 1 {
			page = 1
		}
		if pageSize < 1 || pageSize > 100 {
			pageSize = 10
		}

		users, totalCount, err := cfg.SqlServer.GetAllUsers(c.Request.Context(), page, pageSize, onlyActive)
		if err != nil {
			c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
				BaseResponse: dto.BaseResponse{
					Success:   false,
					Timestamp: time.Now(),
				},
				Error:   "Internal Server Error",
				Code:    http.StatusInternalServerError,
				Message: "Failed to retrieve users",
				Details: err.Error(),
			})
			return
		}

		var userResponses []dto.UserResponse
		for _, user := range users {
			userResponses = append(userResponses, dto.UserResponse{
				Id:          user.Id,
				Name:        user.Name,
				Email:       user.Email,
				UserType:    user.UserType,
				MicrosoftId: user.MicrosoftId,
				IsActive:    user.IsActive,
				CreatedAt:   user.CreatedAt,
				UpdatedAt:   user.UpdatedAt,
				LastLoginAt: user.LastLoginAt,
			})
		}

		response := dto.UsersListResponse{
			Users:      userResponses,
			TotalCount: int(totalCount),
			Page:       page,
			PageSize:   pageSize,
		}

		c.JSON(http.StatusOK, dto.SuccessResponse{
			BaseResponse: dto.BaseResponse{
				Success:   true,
				Timestamp: time.Now(),
			},
			Data:    response,
			Message: "Users retrieved successfully",
		})
	}
}

// UpdateUser atualiza um usuário
// @Summary      Atualizar Usuário
// @Description  Atualiza os dados de um usuário
// @Tags         users
// @Accept       json
// @Produce      json
// @Security 	 BearerAuth
// @Param        id path int true "ID do usuário"
// @Param        user body dto.UpdateUserRequest true "Dados para atualização"
// @Success      200 {object} dto.SuccessResponse
// @Failure 	 400 {object} dto.ErrorResponse "Bad Request"
// @Failure 	 401 {object} dto.AuthErrorResponse "Unauthorized"
// @Failure 	 404 {object} dto.ErrorResponse "Not Found"
// @Failure 	 409 {object} dto.ErrorResponse "Conflict"
// @Failure 	 500 {object} dto.ErrorResponse "Internal Server Error"
// @Router       /users/{id} [put]
func UpdateUser(cfg *config.App) gin.HandlerFunc {
	return func(c *gin.Context) {
		idParam := c.Param("id")
		id, err := strconv.Atoi(idParam)
		if err != nil {
			c.JSON(http.StatusBadRequest, dto.ErrorResponse{
				BaseResponse: dto.BaseResponse{
					Success:   false,
					Timestamp: time.Now(),
				},
				Error:   "Bad Request",
				Code:    http.StatusBadRequest,
				Message: "Invalid user ID",
			})
			return
		}

		var req dto.UpdateUserRequest
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

		if _, ok := utils.UserTypMapStrToInt[*req.UserType]; !ok {
			c.JSON(http.StatusBadRequest, dto.ErrorResponse{
				BaseResponse: dto.BaseResponse{
					Success:   false,
					Timestamp: time.Now(),
				},
				Error:   "Bad Request",
				Code:    http.StatusBadRequest,
				Message: "Invalid parameter",
				Details: fmt.Errorf("the parameter 'userType' must be %v", utils.UserTypMapIntToStr),
			})
			return
		}

		// Buscar usuário existente
		user, err := cfg.SqlServer.GetUserByID(c.Request.Context(), id)
		if err != nil {
			c.JSON(http.StatusNotFound, dto.ErrorResponse{
				BaseResponse: dto.BaseResponse{
					Success:   false,
					Timestamp: time.Now(),
				},
				Error:   "Not Found",
				Code:    http.StatusNotFound,
				Message: "User not found",
			})
			return
		}

		// Verificar se email já está em uso por outro usuário
		if req.Email != nil && *req.Email != user.Email {
			existingUser, _ := cfg.SqlServer.GetUserByEmail(c.Request.Context(), *req.Email)
			if existingUser != nil && existingUser.Id != id {
				c.JSON(http.StatusConflict, dto.ErrorResponse{
					BaseResponse: dto.BaseResponse{
						Success:   false,
						Timestamp: time.Now(),
					},
					Error:   "Conflict",
					Code:    http.StatusConflict,
					Message: "Email already in use",
				})
				return
			}
		}

		// Atualizar campos se fornecidos
		if req.Name != nil {
			user.Name = *req.Name
		}
		if req.Email != nil {
			user.Email = *req.Email
		}
		if req.UserType != nil {
			user.UserType = *req.UserType
		}
		if req.IsActive != nil {
			user.IsActive = *req.IsActive
		}

		// Atualizar senha se fornecida
		if req.Password != nil {
			hash, err := bcrypt.GenerateFromPassword([]byte(*req.Password), bcrypt.DefaultCost)
			if err != nil {
				c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
					BaseResponse: dto.BaseResponse{
						Success:   false,
						Timestamp: time.Now(),
					},
					Error:   "Internal Server Error",
					Code:    http.StatusInternalServerError,
					Message: "Failed to hash password",
					Details: err.Error(),
				})
				return
			}

			if err := cfg.SqlServer.UpdatePassword(c.Request.Context(), id, string(hash), 0); err != nil {
				c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
					BaseResponse: dto.BaseResponse{
						Success:   false,
						Timestamp: time.Now(),
					},
					Error:   "Internal Server Error",
					Code:    http.StatusInternalServerError,
					Message: "Failed to update password",
					Details: err.Error(),
				})
				return

			}
		}

		// Atualizar usuário
		if err := cfg.SqlServer.UpdateUser(c.Request.Context(), id, user); err != nil {
			c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
				BaseResponse: dto.BaseResponse{
					Success:   false,
					Timestamp: time.Now(),
				},
				Error:   "Internal Server Error",
				Code:    http.StatusInternalServerError,
				Message: "Failed to update user",
				Details: err.Error(),
			})
			return
		}

		c.JSON(http.StatusOK, dto.SuccessResponse{
			BaseResponse: dto.BaseResponse{
				Success:   true,
				Timestamp: time.Now(),
			},
			Message: "User updated successfully",
		})
	}
}

// ChangePassword altera a senha do usuário autenticado
// @Summary      Alterar Senha
// @Description  Permite que o usuário autenticado altere sua própria senha
// @Tags         auth
// @Accept       json
// @Produce      json
// @Security 	 BearerAuth
// @Param        request body dto.ChangePasswordRequest true "Senha atual e nova senha"
// @Success      200 {object} dto.SuccessResponse
// @Failure 	 400 {object} dto.ErrorResponse "Bad Request"
// @Failure 	 401 {object} dto.AuthErrorResponse "Unauthorized"
// @Failure 	 403 {object} dto.ErrorResponse "Current password incorrect"
// @Failure 	 500 {object} dto.ErrorResponse "Internal Server Error"
// @Router       /auth/change-password [post]
// ChangePassword altera a senha do usuário autenticado
func ChangePassword(cfg *config.App) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req dto.ChangePasswordRequest
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

		// Pegar claims do JWT
		currentUser, exists := c.Get("currentUser")
		if !exists {
			c.JSON(http.StatusUnauthorized, dto.ErrorResponse{
				BaseResponse: dto.BaseResponse{
					Success:   false,
					Timestamp: time.Now(),
				},
				Error:   "Unauthorized",
				Code:    http.StatusUnauthorized,
				Message: "User not authenticated",
			})
			return
		}

		claims, ok := currentUser.(jwt.MapClaims)
		if !ok {
			c.JSON(http.StatusUnauthorized, dto.ErrorResponse{
				BaseResponse: dto.BaseResponse{
					Success:   false,
					Timestamp: time.Now(),
				},
				Error:   "Unauthorized",
				Code:    http.StatusUnauthorized,
				Message: "Invalid token claims",
			})
			return
		}

		userIdFloat, ok := claims["user_id"].(float64)
		if !ok {
			c.JSON(http.StatusUnauthorized, dto.ErrorResponse{
				BaseResponse: dto.BaseResponse{
					Success:   false,
					Timestamp: time.Now(),
				},
				Error:   "Unauthorized",
				Code:    http.StatusUnauthorized,
				Message: "Invalid user ID in token",
			})
			return
		}

		userId := int(userIdFloat)

		// Buscar usuário
		user, err := cfg.SqlServer.GetUserByID(c.Request.Context(), userId)
		if err != nil {
			c.JSON(http.StatusNotFound, dto.ErrorResponse{
				BaseResponse: dto.BaseResponse{
					Success:   false,
					Timestamp: time.Now(),
				},
				Error:   "Not Found",
				Code:    http.StatusNotFound,
				Message: "User not found",
			})
			return
		}

		// ... (restante do código permanece igual)

		// Verificar senha atual
		if user.PasswordHash == nil {
			c.JSON(http.StatusBadRequest, dto.ErrorResponse{
				BaseResponse: dto.BaseResponse{
					Success:   false,
					Timestamp: time.Now(),
				},
				Error:   "Bad Request",
				Code:    http.StatusBadRequest,
				Message: "User does not have a password (uses Microsoft authentication)",
			})
			return
		}

		err = bcrypt.CompareHashAndPassword([]byte(*user.PasswordHash), []byte(req.CurrentPassword))
		if err != nil {
			c.JSON(http.StatusForbidden, dto.ErrorResponse{
				BaseResponse: dto.BaseResponse{
					Success:   false,
					Timestamp: time.Now(),
				},
				Error:   "Forbidden",
				Code:    http.StatusForbidden,
				Message: "Current password is incorrect",
			})
			return
		}

		// Gerar hash da nova senha
		hash, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
		if err != nil {
			c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
				BaseResponse: dto.BaseResponse{
					Success:   false,
					Timestamp: time.Now(),
				},
				Error:   "Internal Server Error",
				Code:    http.StatusInternalServerError,
				Message: "Failed to hash password",
				Details: err.Error(),
			})
			return
		}

		// Atualizar senha
		if err := cfg.SqlServer.UpdatePassword(c.Request.Context(), userId, string(hash), userId); err != nil {
			c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
				BaseResponse: dto.BaseResponse{
					Success:   false,
					Timestamp: time.Now(),
				},
				Error:   "Internal Server Error",
				Code:    http.StatusInternalServerError,
				Message: "Failed to update password",
				Details: err.Error(),
			})
			return
		}

		c.JSON(http.StatusOK, dto.SuccessResponse{
			BaseResponse: dto.BaseResponse{
				Success:   true,
				Timestamp: time.Now(),
			},
			Message: "Password changed successfully",
		})
	}
}

// DeleteUser deleta (desativa) um usuário (apenas MANAGER/ADMIN)
// @Summary      Deletar Usuário
// @Description  Desativa um usuário do sistema (soft delete) - apenas MANAGER ou ADMIN
// @Tags         users
// @Accept       json
// @Produce      json
// @Security 	 BearerAuth
// @Param        id path int true "ID do usuário"
// @Success      200 {object} dto.SuccessResponse
// @Failure 	 400 {object} dto.ErrorResponse "Bad Request"
// @Failure 	 401 {object} dto.AuthErrorResponse "Unauthorized"
// @Failure 	 404 {object} dto.ErrorResponse "Not Found"
// @Failure 	 500 {object} dto.ErrorResponse "Internal Server Error"
// @Router       /users/{id} [delete]
func DeleteUser(cfg *config.App) gin.HandlerFunc {
	return func(c *gin.Context) {
		idParam := c.Param("id")
		targetId, err := strconv.Atoi(idParam)
		if err != nil {
			c.JSON(http.StatusBadRequest, dto.ErrorResponse{
				BaseResponse: dto.BaseResponse{
					Success:   false,
					Timestamp: time.Now(),
				},
				Error:   "Bad Request",
				Code:    http.StatusBadRequest,
				Message: "Invalid user ID",
			})
			return
		}

		// Pegar ID do usuário autenticado
		currentUserId, _ := c.Get("user_id")
		var deletedBy int
		if uid, ok := currentUserId.(int); ok {
			deletedBy = uid
		}

		// Não permitir que usuário delete a si mesmo via este endpoint
		if deletedBy == targetId {
			c.JSON(http.StatusBadRequest, dto.ErrorResponse{
				BaseResponse: dto.BaseResponse{
					Success:   false,
					Timestamp: time.Now(),
				},
				Error:   "Bad Request",
				Code:    http.StatusBadRequest,
				Message: "User cannot delete themselves via this endpoint. Use /auth/delete-account instead",
			})
			return
		}

		deleteUserAccount(c, cfg, targetId, deletedBy)
	}
}

// DeleteOwnAccount permite que qualquer usuário autenticado delete sua própria conta
// @Summary      Deletar Própria Conta
// @Description  Permite que qualquer usuário autenticado desative sua própria conta
// @Tags         auth
// @Accept       json
// @Produce      json
// @Security 	 BearerAuth
// @Success      200 {object} dto.SuccessResponse
// @Failure 	 401 {object} dto.AuthErrorResponse "Unauthorized"
// @Failure 	 500 {object} dto.ErrorResponse "Internal Server Error"
// @Router       /auth/delete-account [delete]
func DeleteOwnAccount(cfg *config.App) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Pegar claims do JWT
		currentUser, exists := c.Get("currentUser")
		if !exists {
			c.JSON(http.StatusUnauthorized, dto.ErrorResponse{
				BaseResponse: dto.BaseResponse{
					Success:   false,
					Timestamp: time.Now(),
				},
				Error:   "Unauthorized",
				Code:    http.StatusUnauthorized,
				Message: "User not authenticated",
			})
			return
		}

		claims, ok := currentUser.(jwt.MapClaims)
		if !ok {
			c.JSON(http.StatusUnauthorized, dto.ErrorResponse{
				BaseResponse: dto.BaseResponse{
					Success:   false,
					Timestamp: time.Now(),
				},
				Error:   "Unauthorized",
				Code:    http.StatusUnauthorized,
				Message: "Invalid token claims",
			})
			return
		}

		// Extrair user_id do claims
		userIdFloat, ok := claims["user_id"].(float64)
		if !ok {
			c.JSON(http.StatusUnauthorized, dto.ErrorResponse{
				BaseResponse: dto.BaseResponse{
					Success:   false,
					Timestamp: time.Now(),
				},
				Error:   "Unauthorized",
				Code:    http.StatusUnauthorized,
				Message: "Invalid user ID in token",
			})
			return
		}

		userId := int(userIdFloat)

		// Deletar a própria conta (userId como target e deletedBy)
		deleteUserAccount(c, cfg, userId, userId)
	}
}

// deleteUserAccount é a função auxiliar compartilhada para deletar usuário
func deleteUserAccount(c *gin.Context, cfg *config.App, targetUserId int, deletedBy int) {
	if err := cfg.SqlServer.DeleteUser(c.Request.Context(), targetUserId, deletedBy); err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			BaseResponse: dto.BaseResponse{
				Success:   false,
				Timestamp: time.Now(),
			},
			Error:   "Internal Server Error",
			Code:    http.StatusInternalServerError,
			Message: "Failed to delete account",
			Details: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, dto.SuccessResponse{
		BaseResponse: dto.BaseResponse{
			Success:   true,
			Timestamp: time.Now(),
		},
		Message: "Account deleted successfully",
	})
}
