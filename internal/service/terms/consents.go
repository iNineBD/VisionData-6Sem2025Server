package terms

import (
	"errors"
	"fmt"
	"net/http"
	"orderstreamrest/internal/config"
	"orderstreamrest/internal/models/dto"
	"orderstreamrest/internal/models/entities"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt"
)

// Helper function to extract userId from JWT claims
func getUserIdFromContext(c *gin.Context) (int, error) {
	claimsInterface, exists := c.Get("currentUser")
	if !exists {
		return 0, errors.New("user not found in context")
	}

	// O middleware salva como jwt.MapClaims
	claims, ok := claimsInterface.(jwt.MapClaims)
	if !ok {
		// Tentar como map[string]interface{} também
		claimsMap, ok := claimsInterface.(map[string]interface{})
		if !ok {
			return 0, fmt.Errorf("invalid claims format: %T", claimsInterface)
		}
		claims = jwt.MapClaims(claimsMap)
	}

	userIdFloat, ok := claims["user_id"].(float64)
	if !ok {
		return 0, errors.New("invalid user_id type")
	}

	return int(userIdFloat), nil
}

// GetMyConsentStatus retorna o status de consentimento do usuário autenticado com termo completo
// @Summary      Status do Consentimento
// @Description  Retorna o termo de uso ativo completo e o status de consentimento do usuário autenticado
// @Tags         consents
// @Accept       json
// @Produce      json
// @Security 	 BearerAuth
// @Success      200 {object} dto.SuccessResponse{data=dto.MyConsentStatusResponse}
// @Failure 	 401 {object} dto.AuthErrorResponse "Unauthorized"
// @Failure 	 403 {object} dto.AuthErrorResponse "Forbidden"
// @Failure 	 500 {object} dto.ErrorResponse "Internal Server Error"
// @Router       /consents/me [get]
func GetMyConsentStatus(cfg *config.App) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Pegar ID do usuário autenticado
		userId, err := getUserIdFromContext(c)
		if err != nil {
			c.JSON(http.StatusUnauthorized, dto.ErrorResponse{
				BaseResponse: dto.BaseResponse{
					Success:   false,
					Timestamp: time.Now(),
				},
				Error:   "Unauthorized",
				Code:    http.StatusUnauthorized,
				Message: "Usuário não autenticado",
			})
			return
		}

		// Buscar termo ativo com itens
		activeTerm, err := cfg.SqlServer.GetActiveTermWithItems(c.Request.Context())
		if err != nil {
			c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
				BaseResponse: dto.BaseResponse{
					Success:   false,
					Timestamp: time.Now(),
				},
				Error:   "Internal Server Error",
				Code:    http.StatusInternalServerError,
				Message: "Falha ao buscar termo ativo",
				Details: err.Error(),
			})
			return
		}

		response := dto.MyConsentStatusResponse{
			UserId:           userId,
			HasActiveConsent: false,
			NeedsNewConsent:  true,
		}

		// Converter termo para DTO
		if activeTerm != nil {
			termResponse := &dto.TermsOfUseResponse{
				Id:            activeTerm.Id,
				Version:       activeTerm.Version,
				Title:         activeTerm.Title,
				Description:   activeTerm.Description,
				Content:       activeTerm.Content,
				IsActive:      activeTerm.IsActive,
				EffectiveDate: activeTerm.EffectiveDate,
				CreatedAt:     activeTerm.CreatedAt,
				Items:         []dto.TermItemResponse{},
			}

			// Adicionar itens do termo
			for _, item := range activeTerm.Items {
				termResponse.Items = append(termResponse.Items, dto.TermItemResponse{
					Id:          item.Id,
					TermId:      item.TermId,
					ItemOrder:   item.ItemOrder,
					Title:       item.Title,
					Content:     item.Content,
					IsMandatory: item.IsMandatory,
					IsActive:    item.IsActive,
				})
			}

			response.Term = termResponse

			// Buscar consentimento do usuário para este termo
			consent, err := cfg.SqlServer.GetUserActiveConsent(c.Request.Context(), userId)
			if err != nil {
				c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
					BaseResponse: dto.BaseResponse{
						Success:   false,
						Timestamp: time.Now(),
					},
					Error:   "Internal Server Error",
					Code:    http.StatusInternalServerError,
					Message: "Falha ao buscar consentimento",
					Details: err.Error(),
				})
				return
			}

			// Se tem consentimento ativo para o termo atual
			if consent != nil && consent.TermId == activeTerm.Id {
				response.HasActiveConsent = true
				response.NeedsNewConsent = false
				response.ConsentDate = &consent.ConsentDate
			}
		}

		c.JSON(http.StatusOK, dto.SuccessResponse{
			BaseResponse: dto.BaseResponse{
				Success:   true,
				Timestamp: time.Now(),
			},
			Data:    response,
			Message: "Status de consentimento recuperado com sucesso",
		})
	}
}

// GetUserConsent retorna o consentimento de um usuário específico (admin)
// @Summary      Obter Consentimento do Usuário
// @Description  Retorna o consentimento ativo de um usuário específico (apenas administradores)
// @Tags         consents
// @Accept       json
// @Produce      json
// @Security 	 BearerAuth
// @Param        userId path int true "ID do usuário"
// @Success      200 {object} dto.SuccessResponse{data=dto.UserConsentResponse}
// @Failure 	 400 {object} dto.ErrorResponse "Bad Request"
// @Failure 	 401 {object} dto.AuthErrorResponse "Unauthorized"
// @Failure 	 403 {object} dto.ErrorResponse "Forbidden"
// @Failure 	 404 {object} dto.ErrorResponse "Not Found"
// @Failure 	 500 {object} dto.ErrorResponse "Internal Server Error"
// @Router       /consents/user/{userId} [get]
func GetUserConsent(cfg *config.App) gin.HandlerFunc {
	return func(c *gin.Context) {
		userIdParam := c.Param("userId")
		userId, err := strconv.Atoi(userIdParam)
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

		consent, err := cfg.SqlServer.GetUserActiveConsent(c.Request.Context(), userId)
		if err != nil {
			c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
				BaseResponse: dto.BaseResponse{
					Success:   false,
					Timestamp: time.Now(),
				},
				Error:   "Internal Server Error",
				Code:    http.StatusInternalServerError,
				Message: "Failed to get user consent",
				Details: err.Error(),
			})
			return
		}

		if consent == nil {
			c.JSON(http.StatusNotFound, dto.ErrorResponse{
				BaseResponse: dto.BaseResponse{
					Success:   false,
					Timestamp: time.Now(),
				},
				Error:   "Not Found",
				Code:    http.StatusNotFound,
				Message: "User has no active consent",
			})
			return
		}

		// Converter para DTO
		response := dto.UserConsentResponse{
			Id:           consent.Id,
			UserId:       consent.UserId,
			TermId:       consent.TermId,
			TermVersion:  consent.Term.Version,
			ConsentDate:  consent.ConsentDate,
			IsActive:     consent.IsActive,
			ItemConsents: []dto.UserItemConsentResponse{},
		}

		for _, itemConsent := range consent.ItemConsents {
			response.ItemConsents = append(response.ItemConsents, dto.UserItemConsentResponse{
				ItemId:      itemConsent.ItemId,
				ItemTitle:   itemConsent.Item.Title,
				Accepted:    itemConsent.Accepted,
				IsMandatory: itemConsent.Item.IsMandatory,
			})
		}

		c.JSON(http.StatusOK, dto.SuccessResponse{
			BaseResponse: dto.BaseResponse{
				Success:   true,
				Timestamp: time.Now(),
			},
			Data:    response,
			Message: "User consent retrieved successfully",
		})
	}
}

// RegisterMyConsent registra o consentimento do usuário logado para um termo
// @Summary      Registrar Consentimento
// @Description  Permite que o usuário autenticado registre o aceite de um termo de uso
// @Tags         consents
// @Accept       json
// @Produce      json
// @Security 	 BearerAuth
// @Param        consent body dto.UserConsentRequest true "Dados do consentimento"
// @Success      200 {object} dto.SuccessResponse
// @Failure 	 400 {object} dto.ErrorResponse "Bad Request"
// @Failure 	 401 {object} dto.AuthErrorResponse "Unauthorized"
// @Failure 	 500 {object} dto.ErrorResponse "Internal Server Error"
// @Router       /consents/me [post]
func RegisterMyConsent(cfg *config.App) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 1. Pegar ID do usuário autenticado do token JWT
		userId, err := getUserIdFromContext(c)
		if err != nil {
			c.JSON(http.StatusUnauthorized, dto.ErrorResponse{
				BaseResponse: dto.BaseResponse{Success: false, Timestamp: time.Now()},
				Error:        "Unauthorized",
				Code:         http.StatusUnauthorized,
				Message:      "Usuário não autenticado",
			})
			return
		}

		// 2. Ler o corpo da requisição
		var req dto.UserConsentRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, dto.ErrorResponse{
				BaseResponse: dto.BaseResponse{Success: false, Timestamp: time.Now()},
				Error:        "Bad Request",
				Code:         http.StatusBadRequest,
				Message:      "Corpo da requisição inválido",
				Details:      err.Error(),
			})
			return
		}

		// 3. Verificar se o termo existe e está ativo
		term, err := cfg.SqlServer.GetTermByID(c.Request.Context(), req.TermId)
		if err != nil || !term.IsActive {
			c.JSON(http.StatusBadRequest, dto.ErrorResponse{
				BaseResponse: dto.BaseResponse{Success: false, Timestamp: time.Now()},
				Error:        "Bad Request",
				Code:         http.StatusBadRequest,
				Message:      "Termo inválido ou inativo",
			})
			return
		}

		// 4. Validar se os itens obrigatórios foram aceitos
		hasAcceptedMandatory := false
		for _, itemConsent := range req.ItemConsents {
			for _, termItem := range term.Items {
				if termItem.Id == itemConsent.ItemId && termItem.IsMandatory && itemConsent.Accepted {
					hasAcceptedMandatory = true
					break
				}
			}
		}

		if !hasAcceptedMandatory {
			c.JSON(http.StatusBadRequest, dto.ErrorResponse{
				BaseResponse: dto.BaseResponse{Success: false, Timestamp: time.Now()},
				Error:        "Bad Request",
				Code:         http.StatusBadRequest,
				Message:      "É necessário aceitar os itens obrigatórios do termo",
			})
			return
		}

		// 5. Preparar o objeto de consentimento (Lógica abstraída do CreateUser)
		ipAddress := c.ClientIP()
		userAgent := c.GetHeader("User-Agent")

		consent := &entities.UserTermConsent{
			UserId:       userId,
			TermId:       req.TermId,
			ConsentDate:  time.Now(),
			IsActive:     true,
			IPAddress:    &ipAddress,
			UserAgent:    &userAgent,
			ItemConsents: []entities.UserItemConsent{},
		}

		for _, itemConsent := range req.ItemConsents {
			consent.ItemConsents = append(consent.ItemConsents, entities.UserItemConsent{
				ItemId:      itemConsent.ItemId,
				Accepted:    itemConsent.Accepted,
				ConsentDate: time.Now(),
			})
		}

		// 6. Salvar no banco
		err = cfg.SqlServer.RegisterUserConsent(c.Request.Context(), consent)
		if err != nil {
			c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
				BaseResponse: dto.BaseResponse{Success: false, Timestamp: time.Now()},
				Error:        "Internal Server Error",
				Code:         http.StatusInternalServerError,
				Message:      "Falha ao registrar consentimento",
				Details:      err.Error(),
			})
			return
		}

		c.JSON(http.StatusOK, dto.SuccessResponse{
			BaseResponse: dto.BaseResponse{Success: true, Timestamp: time.Now()},
			Message:      "Consentimento registrado com sucesso",
		})
	}
}
