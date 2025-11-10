package users

import (
	"net/http"
	"orderstreamrest/internal/config"
	"orderstreamrest/internal/models/dto"
	"orderstreamrest/internal/models/entities"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

// CreateUser cria um novo usuário
// @Summary      Criar Usuário
// @Description  Cria um novo usuário no sistema
// @Tags         users
// @Accept       json
// @Produce      json
// @Security 	 BearerAuth
// @Param        user body dto.CreateUserRequest true "Dados do usuário"
// @Success      201 {object} dto.SuccessResponse{data=dto.UserCreatedResponse}
// @Failure 	 400 {object} dto.ErrorResponse "Bad Request"
// @Failure 	 401 {object} dto.AuthErrorResponse "Unauthorized"
// @Failure 	 403 {object} dto.ErrorResponse "Forbidden"
// @Failure 	 409 {object} dto.ErrorResponse "Conflict - Email já existe"
// @Failure 	 500 {object} dto.ErrorResponse "Internal Server Error"
// @Router       /users [post]
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

		// Validar que pelo menos senha ou MicrosoftId foi fornecido
		if req.Password == nil && req.MicrosoftId == nil {
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

		// Pegar ID do usuário autenticado (assumindo que está no contexto)
		currentUserId, _ := c.Get("user_id")
		var createdBy *int
		if id, ok := currentUserId.(int); ok {
			createdBy = &id
		}

		user := &entities.User{
			Name:         req.Name,
			Email:        req.Email,
			PasswordHash: passwordHash,
			UserType:     req.UserType,
			MicrosoftId:  req.MicrosoftId,
			IsActive:     true,
			CreatedBy:    createdBy,
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
// @Description  Permite que o usuário altere sua própria senha
// @Tags         users
// @Accept       json
// @Produce      json
// @Security 	 BearerAuth
// @Param        request body dto.ChangePasswordRequest true "Senha atual e nova senha"
// @Success      200 {object} dto.SuccessResponse
// @Failure 	 400 {object} dto.ErrorResponse "Bad Request"
// @Failure 	 401 {object} dto.AuthErrorResponse "Unauthorized"
// @Failure 	 403 {object} dto.ErrorResponse "Current password incorrect"
// @Failure 	 500 {object} dto.ErrorResponse "Internal Server Error"
// @Router       /users/change-password [post]
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

		// Pegar ID do usuário autenticado
		currentUserId, exists := c.Get("user_id")
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

		userId := currentUserId.(int)

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

// DeleteUser deleta (desativa) um usuário
// @Summary      Deletar Usuário
// @Description  Desativa um usuário do sistema (soft delete)
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

		// Pegar ID do usuário autenticado
		currentUserId, _ := c.Get("user_id")
		var deletedBy int
		if uid, ok := currentUserId.(int); ok {
			deletedBy = uid
		}

		// Não permitir que usuário delete a si mesmo
		if deletedBy == id {
			c.JSON(http.StatusBadRequest, dto.ErrorResponse{
				BaseResponse: dto.BaseResponse{
					Success:   false,
					Timestamp: time.Now(),
				},
				Error:   "Bad Request",
				Code:    http.StatusBadRequest,
				Message: "User cannot delete themselves",
			})
			return
		}

		if err := cfg.SqlServer.DeleteUser(c.Request.Context(), id, deletedBy); err != nil {
			c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
				BaseResponse: dto.BaseResponse{
					Success:   false,
					Timestamp: time.Now(),
				},
				Error:   "Internal Server Error",
				Code:    http.StatusInternalServerError,
				Message: "Failed to delete user",
				Details: err.Error(),
			})
			return
		}

		c.JSON(http.StatusOK, dto.SuccessResponse{
			BaseResponse: dto.BaseResponse{
				Success:   true,
				Timestamp: time.Now(),
			},
			Message: "User deleted successfully",
		})
	}
}
