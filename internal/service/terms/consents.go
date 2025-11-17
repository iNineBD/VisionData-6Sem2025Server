package terms

import (
	"errors"
	"fmt"
	"net/http"
	"orderstreamrest/internal/config"
	"orderstreamrest/internal/models/dto"
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

// GetMyConsentStatus retorna o status de consentimento do usuário autenticado
// @Summary      Status do Consentimento
// @Description  Retorna o status de consentimento do usuário autenticado (ADMIN ou SUPPORT)
// @Tags         consents
// @Accept       json
// @Produce      json
// @Security 	 BearerAuth
// @Success      200 {object} dto.SuccessResponse{data=dto.UserConsentStatusResponse}
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
				Message: "User not authenticated",
			})
			return
		}

		// Buscar consentimento ativo do usuário
		consent, err := cfg.SqlServer.GetUserActiveConsent(c.Request.Context(), userId)
		if err != nil {
			c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
				BaseResponse: dto.BaseResponse{
					Success:   false,
					Timestamp: time.Now(),
				},
				Error:   "Internal Server Error",
				Code:    http.StatusInternalServerError,
				Message: "Failed to get consent status",
				Details: err.Error(),
			})
			return
		}

		response := dto.UserConsentStatusResponse{
			UserId:           userId,
			HasActiveConsent: consent != nil,
		}

		if consent != nil {
			response.CurrentTermId = &consent.TermId
			response.CurrentTermVersion = &consent.Term.Version
			response.CurrentTermTitle = &consent.Term.Title
			response.ConsentDate = &consent.ConsentDate
			response.NeedsNewConsent = false
		} else {
			response.NeedsNewConsent = true
		}

		c.JSON(http.StatusOK, dto.SuccessResponse{
			BaseResponse: dto.BaseResponse{
				Success:   true,
				Timestamp: time.Now(),
			},
			Data:    response,
			Message: "Consent status retrieved successfully",
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
