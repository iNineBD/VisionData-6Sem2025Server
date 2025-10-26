package metrics

import (
	"net/http"
	"orderstreamrest/internal/config"
	"orderstreamrest/internal/models/dto"
	"sort"
	"strconv"
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

// MeanTimeByPriority Tempo médio por prioridade
// @Summary      Tempo Médio de Resolução por Prioridade
// @Description  Retorna o tempo médio de resolução dos tickets, agrupado por prioridade, em horas e dias.
// @Tags         metrics
// @Accept       json
// @Produce      json
// @Security 	 BearerAuth
// @Success      200 {object} dto.SuccessResponse{data=[]dto.MeanTimeByPriority} "Mean time by priority retrieved successfully"
// @Failure 	 400 {object} dto.ErrorResponse "Bad Request"
// @Failure 	 401 {object} dto.AuthErrorResponse "Unauthorized - Invalid token"
// @Failure 	 403 {object} dto.ErrorResponse "Forbidden - No permission"
// @Failure 	 429 {object} dto.RateLimitErrorResponse "Rate limit exceeded"
// @Failure 	 500 {object} dto.ErrorResponse "Internal Server Error"
// @Header       200 {string} X-RateLimit-Limit "Requests per minute limit"
// @Header       200 {string} X-RateLimit-Remaining "Remaining requests in the period"
// @Header       200 {string} X-RateLimit-Reset "Rate limit reset timestamp"
// @Router       /metrics/tickets/mean-time-resolution-by-priority [get]
func MeanTimeByPriority(cfg *config.App) gin.HandlerFunc {
	return func(c *gin.Context) {

		meanTimeByPriority, err := cfg.SqlServer.GetAverageResolutionTime()
		if err != nil {
			c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
				BaseResponse: dto.BaseResponse{
					Success:   false,
					Timestamp: c.GetTime("request_start_time"),
				},
				Error:   "Internal Server Error",
				Code:    http.StatusInternalServerError,
				Message: "Failed to retrieve mean time by priority",
				Details: err.Error(),
			})
			return
		}

		var metrics []dto.MeanTimeByPriority
		for _, item := range meanTimeByPriority {
			metrics = append(metrics, dto.MeanTimeByPriority{
				PriorityName: item.NomePrioridade,
				MeanTimeHour: item.MediaResolucaoHoras,
				MeanTimeDay:  item.MediaResolucaoDias,
			})
		}

		c.JSON(http.StatusOK, dto.SuccessResponse{
			BaseResponse: dto.BaseResponse{
				Success:   true,
				Timestamp: time.Now(),
			},
			Data:    metrics,
			Message: "Mean time by priority retrieved successfully",
		})
	}
}

// QtdTicketsByStatusYearMonth retorna a quantidade de tickets por status, ano e mês
// @Summary      Quantidade de Tickets por Status, Ano e Mês
// @Description  Retorna a contagem de tickets agrupados por status, ano e mês.
// @Tags         metrics
// @Accept       json
// @Produce      json
// @Security 	 BearerAuth
// @Success      200 {object} dto.SuccessResponse{data=dto.TicketsByStatusYearMonth} "Tickets by status and month retrieved successfully"
// @Failure 	 400 {object} dto.ErrorResponse "Bad Request"
// @Failure 	 401 {object} dto.AuthErrorResponse "Unauthorized - Invalid token"
// @Failure 	 403 {object} dto.ErrorResponse "Forbidden - No permission"
// @Failure 	 429 {object} dto.RateLimitErrorResponse "Rate limit exceeded"
// @Failure 	 500 {object} dto.ErrorResponse "Internal Server Error"
// @Header       200 {string} X-RateLimit-Limit "Requests per minute limit"
// @Header       200 {string} X-RateLimit-Remaining "Remaining requests in the period"
// @Header       200 {string} X-RateLimit-Reset "Rate limit reset timestamp"
// @Router       /metrics/tickets/qtd-tickets-by-status-year-month [get]
func QtdTicketsByStatusYearMonth(cfg *config.App) gin.HandlerFunc {
	return func(c *gin.Context) {
		data, err := cfg.SqlServer.GetTicketsByStatusAndMonth()
		if err != nil {
			c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
				BaseResponse: dto.BaseResponse{
					Success:   false,
					Timestamp: c.GetTime("request_start_time"),
				},
				Error:   "Internal Server Error",
				Code:    http.StatusInternalServerError,
				Message: "Failed to retrieve tickets by status and month",
				Details: err.Error(),
			})
			return
		}

		result := make(dto.TicketsByStatusYearMonth) // map[string]YearlyData

		for _, item := range data {
			status := item.NomeStatus // ou o campo correto do seu struct
			year := strconv.Itoa(item.Ano)
			monthly := dto.MonthlyCounts{
				Janeiro:   int64(item.Janeiro),
				Fevereiro: int64(item.Fevereiro),
				Marco:     int64(item.Marco),
				Abril:     int64(item.Abril),
				Maio:      int64(item.Maio),
				Junho:     int64(item.Junho),
				Julho:     int64(item.Julho),
				Agosto:    int64(item.Agosto),
				Setembro:  int64(item.Setembro),
				Outubro:   int64(item.Outubro),
				Novembro:  int64(item.Novembro),
				Dezembro:  int64(item.Dezembro),
			}
			if _, ok := result[status]; !ok {
				result[status] = make(dto.YearlyData)
			}
			result[status][year] = append(result[status][year], monthly)
		}

		c.JSON(http.StatusOK, dto.SuccessResponse{
			BaseResponse: dto.BaseResponse{
				Success:   true,
				Timestamp: time.Now(),
			},
			Data:    result,
			Message: "Tickets by status and month retrieved successfully",
		})
	}
}

// TicketsByMonth retorna a quantidade de tickets por mês
// @Summary      Quantidade de Tickets por ano e mês
// @Description  Retorna a quantidade de tickets agrupados por ano e mês.
// @Tags         metrics
// @Accept       json
// @Produce      json
// @Security 	 BearerAuth
// @Success      200 {object} dto.SuccessResponse{data=dto.TicketsByStatusYearMonth} "Tickets by status and month retrieved successfully"
// @Failure 	 400 {object} dto.ErrorResponse "Bad Request"
// @Failure 	 401 {object} dto.AuthErrorResponse "Unauthorized - Invalid token"
// @Failure 	 403 {object} dto.ErrorResponse "Forbidden - No permission"
// @Failure 	 429 {object} dto.RateLimitErrorResponse "Rate limit exceeded"
// @Failure 	 500 {object} dto.ErrorResponse "Internal Server Error"
// @Header       200 {string} X-RateLimit-Limit "Requests per minute limit"
// @Header       200 {string} X-RateLimit-Remaining "Remaining requests in the period"
// @Header       200 {string} X-RateLimit-Reset "Rate limit reset timestamp"
// @Router       /metrics/tickets/qtd-tickets-by-month [get]
func TicketsByMonth(cfg *config.App) gin.HandlerFunc {
	return func(c *gin.Context) {
		data, err := cfg.SqlServer.GetTicketsByMonth()
		if err != nil {
			c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
				BaseResponse: dto.BaseResponse{
					Success:   false,
					Timestamp: c.GetTime("request_start_time"),
				},
				Error:   "Internal Server Error",
				Code:    http.StatusInternalServerError,
				Message: "Failed to retrieve tickets by month",
				Details: err.Error(),
			})
			return
		}

		// transforma os dados para formato dto.YearlyData
		var convertedData []dto.TicketsByMonth
		for _, item := range data {
			convertedData = append(convertedData, dto.TicketsByMonth{
				Ano:          item.Ano,
				Mes:          item.Mes,
				TotalTickets: int64(item.TotalTickets),
			})
		}
		formattedData := transformToYearlyData(convertedData)

		c.JSON(http.StatusOK, dto.SuccessResponse{
			BaseResponse: dto.BaseResponse{
				Success:   true,
				Timestamp: time.Now(),
			},
			Data:    formattedData,
			Message: "Tickets by month retrieved successfully",
		})

	}
}

// transformToYearlyData converte a lista plana de contagens de tickets por mês em um mapa aninhado por ano.
func transformToYearlyData(data []dto.TicketsByMonth) dto.YearlyData {
	yearlyData := make(map[int]*dto.MonthlyCounts)

	for _, item := range data {
		if _, ok := yearlyData[item.Ano]; !ok {
			yearlyData[item.Ano] = &dto.MonthlyCounts{}
		}
		monthData := yearlyData[item.Ano]
		switch item.Mes {
		case 1:
			monthData.Janeiro = item.TotalTickets
		case 2:
			monthData.Fevereiro = item.TotalTickets
		case 3:
			monthData.Marco = item.TotalTickets
		case 4:
			monthData.Abril = item.TotalTickets
		case 5:
			monthData.Maio = item.TotalTickets
		case 6:
			monthData.Junho = item.TotalTickets
		case 7:
			monthData.Julho = item.TotalTickets
		case 8:
			monthData.Agosto = item.TotalTickets
		case 9:
			monthData.Setembro = item.TotalTickets
		case 10:
			monthData.Outubro = item.TotalTickets
		case 11:
			monthData.Novembro = item.TotalTickets
		case 12:
			monthData.Dezembro = item.TotalTickets
		}
	}

	result := make(dto.YearlyData)
	for year, counts := range yearlyData {
		yearStr := strconv.Itoa(year)
		result[yearStr] = []dto.MonthlyCounts{*counts}
	}

	return result
}

// TicketsByPriorityAndMonth retorna a quantidade de tickets por prioridade, ano e mês
// @Summary      Quantidade de Tickets por Prioridade, Ano e Mês
// @Description  Retorna a contagem de tickets agrupados por prioridade, ano e mês.
// @Tags         metrics
// @Accept       json
// @Produce      json
// @Security 	 BearerAuth
// @Success      200 {object} dto.SuccessResponse{data=dto.TicketsByStatusYearMonth} "Tickets by priority and month retrieved successfully"
// @Failure 	 400 {object} dto.ErrorResponse "Bad Request"
// @Failure 	 401 {object} dto.AuthErrorResponse "Unauthorized - Invalid token"
// @Failure 	 403 {object} dto.ErrorResponse "Forbidden - No permission"
// @Failure 	 429 {object} dto.RateLimitErrorResponse "Rate limit exceeded"
// @Failure 	 500 {object} dto.ErrorResponse "Internal Server Error"
// @Header       200 {string} X-RateLimit-Limit "Requests per minute limit"
// @Header       200 {string} X-RateLimit-Remaining "Remaining requests in the period"
// @Header       200 {string} X-RateLimit-Reset "Rate limit reset timestamp"
// @Router       /metrics/tickets/qtd-tickets-by-priority-year-month [get]
func TicketsByPriorityAndMonth(cfg *config.App) gin.HandlerFunc {
	return func(c *gin.Context) {
		data, err := cfg.SqlServer.GetTicketsByPriorityAndMonth()
		if err != nil {
			c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
				BaseResponse: dto.BaseResponse{
					Success:   false,
					Timestamp: c.GetTime("request_start_time"),
				},
				Error:   "Internal Server Error",
				Code:    http.StatusInternalServerError,
				Message: "Failed to retrieve tickets by priority and month",
				Details: err.Error(),
			})
			return
		}

		result := make(dto.TicketsByStatusYearMonth) // map[string]YearlyData

		for _, item := range data {
			// Use o campo correto conforme o struct retornado pelo seu repositório:
			priority := item.NomePrioridades // ou item.NomeStatus, se for esse o nome correto
			year := strconv.Itoa(item.Ano)
			monthly := dto.MonthlyCounts{
				Janeiro:   int64(item.Janeiro),
				Fevereiro: int64(item.Fevereiro),
				Marco:     int64(item.Marco),
				Abril:     int64(item.Abril),
				Maio:      int64(item.Maio),
				Junho:     int64(item.Junho),
				Julho:     int64(item.Julho),
				Agosto:    int64(item.Agosto),
				Setembro:  int64(item.Setembro),
				Outubro:   int64(item.Outubro),
				Novembro:  int64(item.Novembro),
				Dezembro:  int64(item.Dezembro),
			}
			if _, ok := result[priority]; !ok {
				result[priority] = make(dto.YearlyData)
			}
			result[priority][year] = append(result[priority][year], monthly)
		}

		c.JSON(http.StatusOK, dto.SuccessResponse{
			BaseResponse: dto.BaseResponse{
				Success:   true,
				Timestamp: time.Now(),
			},
			Data:    result,
			Message: "Tickets by priority and month retrieved successfully",
		})
	}
}
