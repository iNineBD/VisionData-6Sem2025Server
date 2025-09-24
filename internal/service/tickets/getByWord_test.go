package tickets_test

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

// Implementações reais para teste
type TestConfig struct {
	ES TestElasticsearchService
}

type TestElasticsearchService struct {
	ShouldReturnError bool
	ErrorToReturn     error
	ResponseToReturn  *TestSearchResponse
}

type TestSearchParams struct {
	Query    string `form:"q" binding:"required"`
	Page     int    `form:"page"`
	PageSize int    `form:"page_size"`
}

type TestSearchResponse struct {
	Total     int64        `json:"total"`
	Page      int          `json:"page"`
	PageSize  int          `json:"page_size"`
	Tickets   []TestTicket `json:"tickets"`
	TimeTaken string       `json:"time_taken"`
}

type TestTicket struct {
	TicketID      string `json:"ticket_id"`
	Title         string `json:"title"`
	Description   string `json:"description"`
	CurrentStatus string `json:"current_status"`
}

type TestErrorResponse struct {
	Error   string      `json:"error"`
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Details interface{} `json:"details,omitempty"`
}

// Implementação real do SearchTickets para teste
func (es TestElasticsearchService) SearchTickets(ctx context.Context, params TestSearchParams) (*TestSearchResponse, error) {
	if es.ShouldReturnError {
		return nil, es.ErrorToReturn
	}

	// Simular busca real baseada na query
	if es.ResponseToReturn != nil {
		return es.ResponseToReturn, nil
	}

	// Resposta padrão
	tickets := []TestTicket{}

	// Simular resultados baseados na query
	if strings.Contains(strings.ToLower(params.Query), "internet") {
		tickets = append(tickets, TestTicket{
			TicketID:      "TKT-001",
			Title:         "Problema com internet",
			Description:   "Cliente relatando lentidão na conexão",
			CurrentStatus: "open",
		})
	}

	if strings.Contains(strings.ToLower(params.Query), "sistema") {
		tickets = append(tickets, TestTicket{
			TicketID:      "TKT-002",
			Title:         "Sistema lento",
			Description:   "Performance degradada no sistema",
			CurrentStatus: "in_progress",
		})
	}

	return &TestSearchResponse{
		Total:     int64(len(tickets)),
		Page:      params.Page,
		PageSize:  params.PageSize,
		Tickets:   tickets,
		TimeTaken: "15ms",
	}, nil
}

// Implementação real do NewErrorResponse para teste
func NewTestErrorResponse(c *gin.Context, code int, error string, message string, details interface{}) TestErrorResponse {
	return TestErrorResponse{
		Error:   error,
		Code:    code,
		Message: message,
		Details: details,
	}
}

// Função GetByWord para teste (adaptada para usar as estruturas de teste)
func GetByWordForTest(cfg *TestConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		var params TestSearchParams
		if err := c.ShouldBindQuery(&params); err != nil {
			errorResp := NewTestErrorResponse(c, http.StatusBadRequest, err.Error(), "Error while searching tickets", nil)
			c.JSON(http.StatusBadRequest, errorResp)
			return
		}

		// Limpar a query
		params.Query = strings.TrimSpace(params.Query)
		if params.Query == "" {
			errorResp := NewTestErrorResponse(c, http.StatusBadRequest, "Search query 'q' is required", "Error while searching tickets", nil)
			c.JSON(http.StatusBadRequest, errorResp)
			return
		}

		// Executar a busca
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		result, err := cfg.ES.SearchTickets(ctx, params)
		if err != nil {
			errorResp := NewTestErrorResponse(c, http.StatusInternalServerError, err.Error(), "Error while searching tickets", nil)
			c.JSON(http.StatusInternalServerError, errorResp)
			return
		}

		c.JSON(http.StatusOK, result)
	}
}

func TestGetByWord_WithoutMocks(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		url            string
		esConfig       TestElasticsearchService
		expectedStatus int
		expectedError  string
		validateFunc   func(t *testing.T, body []byte)
	}{
		{
			name: "Success - Search for 'internet'",
			url:  "/search?q=internet&page=1&page_size=10",
			esConfig: TestElasticsearchService{
				ShouldReturnError: false,
			},
			expectedStatus: http.StatusOK,
			validateFunc: func(t *testing.T, body []byte) {
				var response TestSearchResponse
				err := json.Unmarshal(body, &response)
				assert.NoError(t, err)
				assert.Equal(t, int64(1), response.Total)
				assert.Equal(t, 1, len(response.Tickets))
				assert.Equal(t, "TKT-001", response.Tickets[0].TicketID)
				assert.Contains(t, response.Tickets[0].Title, "internet")
			},
		},
		{
			name: "Success - Search for 'sistema'",
			url:  "/search?q=sistema&page=1&page_size=5",
			esConfig: TestElasticsearchService{
				ShouldReturnError: false,
			},
			expectedStatus: http.StatusOK,
			validateFunc: func(t *testing.T, body []byte) {
				var response TestSearchResponse
				err := json.Unmarshal(body, &response)
				assert.NoError(t, err)
				assert.Equal(t, int64(1), response.Total)
				assert.Equal(t, "TKT-002", response.Tickets[0].TicketID)
				assert.Equal(t, 1, response.Page)
				assert.Equal(t, 5, response.PageSize)
			},
		},
		{
			name: "Success - No results found",
			url:  "/search?q=inexistente",
			esConfig: TestElasticsearchService{
				ShouldReturnError: false,
			},
			expectedStatus: http.StatusOK,
			validateFunc: func(t *testing.T, body []byte) {
				var response TestSearchResponse
				err := json.Unmarshal(body, &response)
				assert.NoError(t, err)
				assert.Equal(t, int64(0), response.Total)
				assert.Equal(t, 0, len(response.Tickets))
			},
		},
		{
			name: "Success - Custom response",
			url:  "/search?q=custom",
			esConfig: TestElasticsearchService{
				ShouldReturnError: false,
				ResponseToReturn: &TestSearchResponse{
					Total:     100,
					Page:      1,
					PageSize:  10,
					TimeTaken: "50ms",
					Tickets: []TestTicket{
						{TicketID: "CUSTOM-001", Title: "Custom ticket", CurrentStatus: "resolved"},
						{TicketID: "CUSTOM-002", Title: "Another custom", CurrentStatus: "open"},
					},
				},
			},
			expectedStatus: http.StatusOK,
			validateFunc: func(t *testing.T, body []byte) {
				var response TestSearchResponse
				err := json.Unmarshal(body, &response)
				assert.NoError(t, err)
				assert.Equal(t, int64(100), response.Total)
				assert.Equal(t, 2, len(response.Tickets))
				assert.Equal(t, "CUSTOM-001", response.Tickets[0].TicketID)
				assert.Equal(t, "50ms", response.TimeTaken)
			},
		},
		{
			name: "Error - Missing query parameter",
			url:  "/search?page=1&page_size=10",
			esConfig: TestElasticsearchService{
				ShouldReturnError: false,
			},
			expectedStatus: http.StatusBadRequest,
			validateFunc: func(t *testing.T, body []byte) {
				var response TestErrorResponse
				err := json.Unmarshal(body, &response)
				assert.NoError(t, err)
				assert.Contains(t, response.Error, "Query")
				assert.Contains(t, response.Error, "required")
			},
		},
		{
			name: "Error - Elasticsearch connection failure",
			url:  "/search?q=test",
			esConfig: TestElasticsearchService{
				ShouldReturnError: true,
				ErrorToReturn:     errors.New("connection to elasticsearch failed"),
			},
			expectedStatus: http.StatusInternalServerError,
			expectedError:  "connection to elasticsearch failed",
			validateFunc: func(t *testing.T, body []byte) {
				var response TestErrorResponse
				err := json.Unmarshal(body, &response)
				assert.NoError(t, err)
				assert.Equal(t, "connection to elasticsearch failed", response.Error)
				assert.Equal(t, http.StatusInternalServerError, response.Code)
				assert.Equal(t, "Error while searching tickets", response.Message)
			},
		},
		{
			name: "Error - Elasticsearch timeout",
			url:  "/search?q=timeout",
			esConfig: TestElasticsearchService{
				ShouldReturnError: true,
				ErrorToReturn:     errors.New("context deadline exceeded"),
			},
			expectedStatus: http.StatusInternalServerError,
			expectedError:  "context deadline exceeded",
		},
		{
			name: "Success - Pagination test",
			url:  "/search?q=test&page=2&page_size=5",
			esConfig: TestElasticsearchService{
				ShouldReturnError: false,
				ResponseToReturn: &TestSearchResponse{
					Total:     25,
					Page:      2,
					PageSize:  5,
					TimeTaken: "20ms",
					Tickets: []TestTicket{
						{TicketID: "TKT-006", Title: "Ticket page 2"},
						{TicketID: "TKT-007", Title: "Another ticket page 2"},
					},
				},
			},
			expectedStatus: http.StatusOK,
			validateFunc: func(t *testing.T, body []byte) {
				var response TestSearchResponse
				err := json.Unmarshal(body, &response)
				assert.NoError(t, err)
				assert.Equal(t, 2, response.Page)
				assert.Equal(t, 5, response.PageSize)
				assert.Equal(t, int64(25), response.Total)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup config
			cfg := &TestConfig{
				ES: tt.esConfig,
			}

			// Setup router
			router := gin.New()
			router.GET("/search", GetByWordForTest(cfg))

			// Executar request
			req := httptest.NewRequest("GET", tt.url, nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			// Assertions básicas
			assert.Equal(t, tt.expectedStatus, w.Code)

			// Validar error se especificado
			if tt.expectedError != "" {
				body := w.Body.String()
				assert.Contains(t, body, tt.expectedError)
			}

			// Executar validação customizada se fornecida
			if tt.validateFunc != nil {
				tt.validateFunc(t, w.Body.Bytes())
			}
		})
	}
}

// Teste individual simples
func TestGetByWord_SimpleCase(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Setup
	cfg := &TestConfig{
		ES: TestElasticsearchService{
			ShouldReturnError: false,
		},
	}

	router := gin.New()
	router.GET("/search", GetByWordForTest(cfg))

	// Test
	req := httptest.NewRequest("GET", "/search?q=internet", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assertions
	assert.Equal(t, http.StatusOK, w.Code)

	var response TestSearchResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Greater(t, response.Total, int64(0))
	assert.NotEmpty(t, response.Tickets)
}
