package tickets

import (
	"context"
	"net/http"
	"orderstreamrest/internal/config"
	"orderstreamrest/internal/models/dto"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

// GetByWord handles the GET /tickets endpoint to search tickets by a query word
// @Summary      Search tickets by query word
// @Description  Returns tickets matching the search query
// @Tags         tickets
// @Accept       json
// @Produce      json
// @Param        q     query     string  true  "Search query"
// @Param        page      query     int     false "Page number" default(1)
// @Param        page_size query     int     false "Number of items per page" default(50) maximum(100)
// @Success 	  200 {object} dto.PaginatedResponse{data=[]dto.Ticket}
// @Failure      400   {object}  dto.ErrorResponse
// @Failure      500   {object}  dto.ErrorResponse
// @Router       /tickets/query [get]
func GetByWord(cfg *config.App) gin.HandlerFunc {
	return func(c *gin.Context) {

		var params dto.SearchParams
		if err := c.ShouldBindQuery(&params); err != nil {
			c.JSON(http.StatusBadRequest, dto.NewErrorResponse(c, http.StatusBadRequest, err.Error(), "Error while searching tickets", nil))
			return
		}

		// Limpar a query
		params.Query = strings.TrimSpace(params.Query)
		if params.Query == "" {
			c.JSON(http.StatusBadRequest, dto.NewErrorResponse(c, http.StatusBadRequest, "Search query 'q' is required", "Error while searching tickets", nil))
			return
		}

		// Executar a busca
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		result, err := cfg.ES.SearchTicketsBySomeWord(ctx, params)
		if err != nil {
			c.JSON(http.StatusInternalServerError, dto.NewErrorResponse(c, http.StatusInternalServerError, err.Error(), "Error while searching tickets", nil))
			return
		}

		c.JSON(http.StatusOK, result)

	}
}
