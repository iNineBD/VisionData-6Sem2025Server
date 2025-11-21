package terms

import (
	"fmt"
	"net/http"
	"orderstreamrest/internal/config"
	"orderstreamrest/internal/models/dto"
	"orderstreamrest/internal/models/entities"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

// GetActiveTerm retorna o termo ativo atual com seus itens
// @Summary      Obter Termo Ativo
// @Description  Retorna o termo de uso ativo atual com todos os seus itens (endpoint público para uso no cadastro)
// @Tags         terms
// @Accept       json
// @Produce      json
// @Success      200 {object} dto.SuccessResponse{data=dto.TermsOfUseResponse}
// @Failure 	 404 {object} dto.ErrorResponse "Not Found"
// @Failure 	 500 {object} dto.ErrorResponse "Internal Server Error"
// @Router       /auth/terms/active [get]
func GetActiveTerm(cfg *config.App) gin.HandlerFunc {
	return func(c *gin.Context) {
		term, err := cfg.SqlServer.GetActiveTermWithItems(c.Request.Context())
		if err != nil {
			c.JSON(http.StatusNotFound, dto.ErrorResponse{
				BaseResponse: dto.BaseResponse{
					Success:   false,
					Timestamp: time.Now(),
				},
				Error:   "Not Found",
				Code:    http.StatusNotFound,
				Message: "Nenhum termo ativo encontrado",
				Details: err.Error(),
			})
			return
		}

		// Converter para DTO
		response := dto.TermsOfUseResponse{
			Id:            term.Id,
			Version:       term.Version,
			Title:         term.Title,
			Description:   term.Description,
			Content:       term.Content,
			IsActive:      term.IsActive,
			EffectiveDate: term.EffectiveDate,
			CreatedAt:     term.CreatedAt,
			Items:         []dto.TermItemResponse{},
		}

		for _, item := range term.Items {
			response.Items = append(response.Items, dto.TermItemResponse{
				Id:          item.Id,
				TermId:      item.TermId,
				ItemOrder:   item.ItemOrder,
				Title:       item.Title,
				Content:     item.Content,
				IsMandatory: item.IsMandatory,
				IsActive:    item.IsActive,
			})
		}

		c.JSON(http.StatusOK, dto.SuccessResponse{
			BaseResponse: dto.BaseResponse{
				Success:   true,
				Timestamp: time.Now(),
			},
			Data:    response,
			Message: "Termo ativo recuperado com sucesso",
		})
	}
}

// ListTerms lista todos os termos (somente para admin)
// @Summary      Listar Termos
// @Description  Lista todos os termos de uso (apenas para administradores)
// @Tags         terms
// @Accept       json
// @Produce      json
// @Security 	 BearerAuth
// @Param        page query int false "Número da página" default(1)
// @Param        pageSize query int false "Tamanho da página" default(10)
// @Success      200 {object} dto.SuccessResponse{data=dto.ListTermsResponse}
// @Failure 	 401 {object} dto.AuthErrorResponse "Unauthorized"
// @Failure 	 403 {object} dto.ErrorResponse "Forbidden"
// @Failure 	 500 {object} dto.ErrorResponse "Internal Server Error"
// @Router       /terms [get]
func ListTerms(cfg *config.App) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Paginação
		page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
		pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "10"))

		if page < 1 {
			page = 1
		}
		if pageSize < 1 || pageSize > 100 {
			pageSize = 10
		}

		terms, total, err := cfg.SqlServer.ListAllTerms(c.Request.Context(), page, pageSize)
		if err != nil {
			c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
				BaseResponse: dto.BaseResponse{
					Success:   false,
					Timestamp: time.Now(),
				},
				Error:   "Internal Server Error",
				Code:    http.StatusInternalServerError,
				Message: "Falha ao listar termos",
				Details: err.Error(),
			})
			return
		}

		// Converter para DTO
		termsResponse := []dto.TermsOfUseResponse{}
		for _, term := range terms {
			termResp := dto.TermsOfUseResponse{
				Id:            term.Id,
				Version:       term.Version,
				Title:         term.Title,
				Description:   term.Description,
				Content:       term.Content,
				IsActive:      term.IsActive,
				EffectiveDate: term.EffectiveDate,
				CreatedAt:     term.CreatedAt,
				Items:         []dto.TermItemResponse{},
			}

			for _, item := range term.Items {
				termResp.Items = append(termResp.Items, dto.TermItemResponse{
					Id:          item.Id,
					TermId:      item.TermId,
					ItemOrder:   item.ItemOrder,
					Title:       item.Title,
					Content:     item.Content,
					IsMandatory: item.IsMandatory,
					IsActive:    item.IsActive,
				})
			}

			termsResponse = append(termsResponse, termResp)
		}

		response := dto.ListTermsResponse{
			Terms:      termsResponse,
			TotalCount: int(total),
			Page:       page,
			PageSize:   pageSize,
		}

		c.JSON(http.StatusOK, dto.SuccessResponse{
			BaseResponse: dto.BaseResponse{
				Success:   true,
				Timestamp: time.Now(),
			},
			Data:    response,
			Message: "Termos listados com sucesso",
		})
	}
}

// CreateTerm cria um novo termo (somente admin)
// @Summary      Criar Termo
// @Description  Cria um novo termo de uso com versionamento (apenas administradores)
// @Tags         terms
// @Accept       json
// @Produce      json
// @Security 	 BearerAuth
// @Param        term body dto.CreateTermRequest true "Dados do termo"
// @Success      201 {object} dto.SuccessResponse{data=dto.TermsOfUseResponse}
// @Failure 	 400 {object} dto.ErrorResponse "Bad Request"
// @Failure 	 401 {object} dto.AuthErrorResponse "Unauthorized"
// @Failure 	 403 {object} dto.ErrorResponse "Forbidden"
// @Failure 	 409 {object} dto.ErrorResponse "Conflict - Versão já existe"
// @Failure 	 500 {object} dto.ErrorResponse "Internal Server Error"
// @Router       /terms [post]
func CreateTerm(cfg *config.App) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req dto.CreateTermRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, dto.ErrorResponse{
				BaseResponse: dto.BaseResponse{
					Success:   false,
					Timestamp: time.Now(),
				},
				Error:   "Bad Request",
				Code:    http.StatusBadRequest,
				Message: "Corpo da requisição inválido",
				Details: err.Error(),
			})
			return
		}

		// Validar que a versão não está vazia
		if req.Version == "" {
			c.JSON(http.StatusBadRequest, dto.ErrorResponse{
				BaseResponse: dto.BaseResponse{
					Success:   false,
					Timestamp: time.Now(),
				},
				Error:   "Bad Request",
				Code:    http.StatusBadRequest,
				Message: "A versão não pode estar vazia",
			})
			return
		}

		// Validar que o título não está vazio
		if req.Title == "" {
			c.JSON(http.StatusBadRequest, dto.ErrorResponse{
				BaseResponse: dto.BaseResponse{
					Success:   false,
					Timestamp: time.Now(),
				},
				Error:   "Bad Request",
				Code:    http.StatusBadRequest,
				Message: "O título não pode estar vazio",
			})
			return
		}

		// Validar que o conteúdo não está vazio
		if req.Content == "" {
			c.JSON(http.StatusBadRequest, dto.ErrorResponse{
				BaseResponse: dto.BaseResponse{
					Success:   false,
					Timestamp: time.Now(),
				},
				Error:   "Bad Request",
				Code:    http.StatusBadRequest,
				Message: "O conteúdo não pode estar vazio",
			})
			return
		}

		// Validar que existe pelo menos 1 item
		if len(req.Items) == 0 {
			c.JSON(http.StatusBadRequest, dto.ErrorResponse{
				BaseResponse: dto.BaseResponse{
					Success:   false,
					Timestamp: time.Now(),
				},
				Error:   "Bad Request",
				Code:    http.StatusBadRequest,
				Message: "É necessário pelo menos um item",
			})
			return
		}

		// Verificar se tem pelo menos 1 item obrigatório
		mandatoryCount := 0
		itemOrders := make(map[int]bool)

		for _, item := range req.Items {
			// Validar que itemOrder é positivo
			if item.ItemOrder <= 0 {
				c.JSON(http.StatusBadRequest, dto.ErrorResponse{
					BaseResponse: dto.BaseResponse{
						Success:   false,
						Timestamp: time.Now(),
					},
					Error:   "Bad Request",
					Code:    http.StatusBadRequest,
					Message: fmt.Sprintf("A ordem do item deve ser maior que 0, encontrado: %d", item.ItemOrder),
				})
				return
			}

			// Verificar duplicidade de itemOrder
			if itemOrders[item.ItemOrder] {
				c.JSON(http.StatusBadRequest, dto.ErrorResponse{
					BaseResponse: dto.BaseResponse{
						Success:   false,
						Timestamp: time.Now(),
					},
					Error:   "Bad Request",
					Code:    http.StatusBadRequest,
					Message: fmt.Sprintf("Ordem de item duplicada encontrada: %d", item.ItemOrder),
				})
				return
			}
			itemOrders[item.ItemOrder] = true

			// Validar que título do item não está vazio
			if item.Title == "" {
				c.JSON(http.StatusBadRequest, dto.ErrorResponse{
					BaseResponse: dto.BaseResponse{
						Success:   false,
						Timestamp: time.Now(),
					},
					Error:   "Bad Request",
					Code:    http.StatusBadRequest,
					Message: fmt.Sprintf("O item na ordem %d deve ter um título", item.ItemOrder),
				})
				return
			}

			// Validar que conteúdo do item não está vazio
			if item.Content == "" {
				c.JSON(http.StatusBadRequest, dto.ErrorResponse{
					BaseResponse: dto.BaseResponse{
						Success:   false,
						Timestamp: time.Now(),
					},
					Error:   "Bad Request",
					Code:    http.StatusBadRequest,
					Message: fmt.Sprintf("O item '%s' deve ter conteúdo", item.Title),
				})
				return
			}

			if item.IsMandatory {
				mandatoryCount++
			}
		}

		if mandatoryCount == 0 {
			c.JSON(http.StatusBadRequest, dto.ErrorResponse{
				BaseResponse: dto.BaseResponse{
					Success:   false,
					Timestamp: time.Now(),
				},
				Error:   "Bad Request",
				Code:    http.StatusBadRequest,
				Message: "É necessário pelo menos um item obrigatório",
			})
			return
		}

		if mandatoryCount > 1 {
			c.JSON(http.StatusBadRequest, dto.ErrorResponse{
				BaseResponse: dto.BaseResponse{
					Success:   false,
					Timestamp: time.Now(),
				},
				Error:   "Bad Request",
				Code:    http.StatusBadRequest,
				Message: "Apenas um item obrigatório é permitido por termo",
			})
			return
		}

		// Pegar ID do usuário autenticado
		currentUserId, exists := c.Get("user_id")
		var createdBy *int
		if exists {
			if id, ok := currentUserId.(int); ok {
				createdBy = &id
			}
		}

		createdAt := time.Now()
		effectiveDate := createdAt
		if req.EffectiveDate != nil {
			effectiveDate = *req.EffectiveDate

			// Validar que a data de vigência não pode ser menor que a data de criação
			// Compara apenas a data (sem hora) para evitar problemas com timestamps
			createdDate := time.Date(createdAt.Year(), createdAt.Month(), createdAt.Day(), 0, 0, 0, 0, createdAt.Location())
			effectiveOnlyDate := time.Date(effectiveDate.Year(), effectiveDate.Month(), effectiveDate.Day(), 0, 0, 0, 0, effectiveDate.Location())

			if effectiveOnlyDate.Before(createdDate) {
				c.JSON(http.StatusBadRequest, dto.ErrorResponse{
					BaseResponse: dto.BaseResponse{
						Success:   false,
						Timestamp: time.Now(),
					},
					Error:   "Bad Request",
					Code:    http.StatusBadRequest,
					Message: "A data de vigência não pode ser anterior à data de criação",
				})
				return
			}
		}

		term := &entities.TermsOfUse{
			Version:       req.Version,
			Title:         req.Title,
			Description:   req.Description,
			Content:       req.Content,
			IsActive:      true, // Sempre criar como ativo, repositório decide depois se deve desativar
			EffectiveDate: effectiveDate,
			CreatedAt:     createdAt,
			CreatedBy:     createdBy,
			Items:         []entities.TermItem{},
		}

		for _, itemReq := range req.Items {
			term.Items = append(term.Items, entities.TermItem{
				ItemOrder:   itemReq.ItemOrder,
				Title:       itemReq.Title,
				Content:     itemReq.Content,
				IsMandatory: itemReq.IsMandatory,
				IsActive:    true,
				CreatedAt:   createdAt,
			})
		}

		err := cfg.SqlServer.CreateTerm(c.Request.Context(), term)
		if err != nil {
			statusCode := http.StatusInternalServerError
			message := "Falha ao criar termo"

			// Tratamento de erros específicos
			errorMsg := err.Error()

			if errorMsg == fmt.Sprintf("já existe um termo com a versão %s", req.Version) {
				statusCode = http.StatusConflict
				message = fmt.Sprintf("Já existe um termo com a versão '%s'", req.Version)
			} else if errorMsg == "é necessário ter pelo menos 1 item obrigatório" {
				statusCode = http.StatusBadRequest
				message = "É necessário pelo menos um item obrigatório"
			} else if errorMsg == "cada termo pode ter apenas 1 item obrigatório" {
				statusCode = http.StatusBadRequest
				message = "Apenas um item obrigatório é permitido por termo"
			} else if strings.Contains(errorMsg, "falha ao desativar termos") {
				statusCode = http.StatusInternalServerError
				message = "Falha ao atualizar status do termo"
			} else if strings.Contains(errorMsg, "falha ao buscar termo ativo") {
				statusCode = http.StatusInternalServerError
				message = "Falha ao determinar termo ativo"
			} else if strings.Contains(errorMsg, "falha ao ativar termo") {
				statusCode = http.StatusInternalServerError
				message = "Falha ao ativar termo"
			}

			c.JSON(statusCode, dto.ErrorResponse{
				BaseResponse: dto.BaseResponse{
					Success:   false,
					Timestamp: time.Now(),
				},
				Error:   http.StatusText(statusCode),
				Code:    statusCode,
				Message: message,
				Details: errorMsg,
			})
			return
		}

		// Converter para DTO
		response := dto.TermsOfUseResponse{
			Id:            term.Id,
			Version:       term.Version,
			Title:         term.Title,
			Description:   term.Description,
			Content:       term.Content,
			IsActive:      term.IsActive,
			EffectiveDate: term.EffectiveDate,
			CreatedAt:     term.CreatedAt,
			Items:         []dto.TermItemResponse{},
		}

		for _, item := range term.Items {
			response.Items = append(response.Items, dto.TermItemResponse{
				Id:          item.Id,
				TermId:      item.TermId,
				ItemOrder:   item.ItemOrder,
				Title:       item.Title,
				Content:     item.Content,
				IsMandatory: item.IsMandatory,
				IsActive:    item.IsActive,
			})
		}

		c.JSON(http.StatusCreated, dto.SuccessResponse{
			BaseResponse: dto.BaseResponse{
				Success:   true,
				Timestamp: time.Now(),
			},
			Data:    response,
			Message: "Termo criado com sucesso",
		})
	}
}
