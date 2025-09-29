package tickets

import (
	"context"
	"net/http"
	"orderstreamrest/internal/config"
	"orderstreamrest/internal/models/dto"
	"time"

	"github.com/gin-gonic/gin"
)

// SearchTicketByID handles the GET /tickets/:id endpoint to fetch a ticket by its ID
// @Summary      Get ticket by ID
// @Description  Returns a single ticket matching the provided ID
// @Tags         tickets
// @Accept       json
// @Produce      json
// @Param        id   path      string  true  "Ticket ID"
// @Success      200  {object}  dto.Ticket
// @Failure      400  {object}  dto.ErrorResponse
// @Failure      404  {object}  dto.ErrorResponse
// @Failure      500  {object}  dto.ErrorResponse
// @Router       /tickets/{id} [get]
func SearchTicketByID(cfg *config.App) gin.HandlerFunc {
	return func(c *gin.Context) {
		ticketID := c.Param("id")
		if ticketID == "" {
			c.JSON(http.StatusBadRequest, dto.NewErrorResponse(c, http.StatusBadRequest, "Ticket ID is required", "Error while fetching ticket", nil))
			return
		}

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		ticket, err := cfg.ES.SearchTicketByID(ctx, ticketID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, dto.NewErrorResponse(c, http.StatusInternalServerError, err.Error(), "Error while fetching ticket", nil))
			return
		}
		if ticket == nil {
			c.JSON(http.StatusNotFound, dto.NewErrorResponse(c, http.StatusNotFound, "Ticket not found", "Error while fetching ticket", nil))
			return
		}

		c.JSON(http.StatusOK, ticket)
	}
}
