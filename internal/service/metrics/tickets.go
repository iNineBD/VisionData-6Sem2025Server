package metrics

import (
	"net/http"
	"orderstreamrest/internal/config"
	"orderstreamrest/internal/models/dto"
	"sort"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

// GetTicketsMetrics retorna métricas dos tickets
// @Summary      Métricas de Tickets
// @Description  Retorna métricas agregadas dos tickets por categoria, prioridade, canal e tag
// @Tags         metrics
// @Accept       json
// @Produce      json
// @Security 	 BearerAuth
// @Success      200 {object} dto.TicketsMetricsResponse "Tickets metrics retrieved successfully"
// @Failure 	 400 {object} dto.ErrorResponse "Bad Request"
// @Failure 	 401 {object} dto.AuthErrorResponse "Unauthorized - Invalid token"
// @Failure 	 403 {object} dto.ErrorResponse "Forbidden - No permission"
// @Failure 	 429 {object} dto.RateLimitErrorResponse "Rate limit exceeded"
// @Failure 	 500 {object} dto.ErrorResponse "Internal Server Error"
// @Header       200 {string} X-RateLimit-Limit "Requests per minute limit"
// @Header       200 {string} X-RateLimit-Remaining "Remaining requests in the period"
// @Header       200 {string} X-RateLimit-Reset "Rate limit reset timestamp"
// @Router       /metrics/tickets [get]
func GetTicketsMetrics(cfg *config.App) gin.HandlerFunc {
	return func(c *gin.Context) {

		// total de tickets
		total, err := cfg.SqlServer.GetTotalTickets()
		if err != nil {
			c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
				BaseResponse: dto.BaseResponse{
					Success:   false,
					Timestamp: c.GetTime("request_start_time"),
				},
				Error:   "Internal Server Error",
				Code:    http.StatusInternalServerError,
				Message: "Failed to retrieve total tickets",
				Details: err.Error(),
			})
			return
		}

		var metrics []dto.TypeMetric

		// total de tickets por categoria
		ticketsByCategory, err := cfg.SqlServer.GetTicketsByCategory()
		if err == nil {
			var categoryMetrics []dto.MetricValue
			for _, item := range ticketsByCategory {
				categoryMetrics = append(categoryMetrics, dto.MetricValue{
					Name:  item.CategoryName,
					Value: item.Total,
				})
			}
			metrics = append(metrics, dto.TypeMetric{
				Name:   "TicketsByCategory",
				Values: categoryMetrics,
			})
		}

		// total de tickets por prioridade
		ticketsByPriority, err := cfg.SqlServer.GetTicketsByPriority()
		if err == nil {
			// Ordena as prioridades: CRÍTICA, ALTA, MÉDIA, BAIXA
			priorityOrder := map[string]int{
				"CRÍTICA": 1,
				"ALTA":    2,
				"MÉDIA":   3,
				"BAIXA":   4,
			}
			sort.Slice(ticketsByPriority, func(i, j int) bool {
				return priorityOrder[strings.ToUpper(ticketsByPriority[i].Name)] < priorityOrder[strings.ToUpper(ticketsByPriority[j].Name)]
			})
			var priorityMetrics []dto.MetricValue
			for _, item := range ticketsByPriority {
				priorityMetrics = append(priorityMetrics, dto.MetricValue{
					Name:  item.Name,
					Value: item.Total,
				})
			}
			metrics = append(metrics, dto.TypeMetric{
				Name:   "TicketsByPriority",
				Values: priorityMetrics,
			})
		}

		// total de tickets por canal
		ticketsByChannel, err := cfg.SqlServer.GetTicketsByChannel()
		if err == nil {
			var channelMetrics []dto.MetricValue
			for _, item := range ticketsByChannel {
				channelMetrics = append(channelMetrics, dto.MetricValue{
					Name:  item.ChannelName,
					Value: item.Total,
				})
			}
			metrics = append(metrics, dto.TypeMetric{
				Name:   "TicketsByChannel",
				Values: channelMetrics,
			})
		}

		// total de tickets por Tag
		ticketsByTag, err := cfg.SqlServer.GetTicketsByTag()
		if err == nil {
			var tagMetrics []dto.MetricValue
			for _, item := range ticketsByTag {
				tagMetrics = append(tagMetrics, dto.MetricValue{
					Name:  item.Name,
					Value: item.Total,
				})
			}
			metrics = append(metrics, dto.TypeMetric{
				Name:   "TicketsByTag",
				Values: tagMetrics,
			})
		}

		// total de tickets por departamento
		ticketsByDepartment, err := cfg.SqlServer.GetTicketsByDepartment()
		if err == nil {
			var departmentMetrics []dto.MetricValue
			for _, item := range ticketsByDepartment {
				departmentMetrics = append(departmentMetrics, dto.MetricValue{
					Name:  item.Name,
					Value: item.Total,
				})
			}
			metrics = append(metrics, dto.TypeMetric{
				Name:   "TicketsByDepartment",
				Values: departmentMetrics,
			})
		}

		response := dto.TicketsMetricsResponse{
			TotalTickets: total,
			Metrics:      metrics,
		}

		// montando o json de response
		c.JSON(http.StatusOK, dto.SuccessResponse{
			BaseResponse: dto.BaseResponse{
				Success:   true,
				Timestamp: time.Now(),
			},
			Data:    response,
			Message: "Tickets metrics retrieved successfully",
		})

	}
}
