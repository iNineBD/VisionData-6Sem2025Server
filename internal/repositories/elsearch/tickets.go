package elsearch

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"orderstreamrest/internal/models/dto"
	"time"

	"github.com/elastic/go-elasticsearch/esapi"
	"github.com/google/uuid"
)

// SearchTicketsBySomeWord realiza uma busca paginada de tickets com base nos parâmetros fornecidos
func (es *Client) SearchTicketsBySomeWord(ctx context.Context, params dto.SearchParams) (*dto.PaginatedResponse, error) {
	// Configurar paginação
	if params.Page < 1 {
		params.Page = 1
	}
	if params.PageSize < 1 || params.PageSize > 100 {
		params.PageSize = 50
	}

	from := (params.Page - 1) * params.PageSize

	// Construir a query
	searchQuery := es.buildSearchQuery(params.Query, from, params.PageSize)

	// Converter query para JSON
	queryJSON, err := json.Marshal(searchQuery)
	if err != nil {
		return nil, fmt.Errorf("erro ao serializar query: %v", err)
	}

	// Executar a busca
	req := esapi.SearchRequest{
		Index: []string{es.config.IndexName},
		Body:  bytes.NewReader(queryJSON),
	}

	res, err := req.Do(ctx, es.ES)
	if err != nil {
		return nil, fmt.Errorf("erro na execução da busca: %v", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		body, _ := io.ReadAll(res.Body)
		return nil, fmt.Errorf("erro na busca: %s - %s", res.Status(), string(body))
	}

	// Ler resposta
	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, fmt.Errorf("erro ao ler resposta: %v", err)
	}

	// Parse da resposta
	var esResponse dto.ESResponse
	if err := json.Unmarshal(body, &esResponse); err != nil {
		return nil, fmt.Errorf("erro ao deserializar resposta: %v", err)
	}

	// Processar resultados
	tickets := make([]map[string]interface{}, 0, len(esResponse.Hits.Hits))
	for _, hit := range esResponse.Hits.Hits {
		var ticket map[string]interface{}
		if err := json.Unmarshal(hit.Source, &ticket); err != nil {
			log.Printf("Erro ao deserializar ticket: %v", err)
			continue
		}
		tickets = append(tickets, ticket)
	}

	// Calcular paginação
	totalPages := int((esResponse.Hits.Total.Value + int64(params.PageSize) - 1) / int64(params.PageSize))

	return &dto.PaginatedResponse{
		BaseResponse: dto.BaseResponse{
			Success:   true,
			Timestamp: time.Now(),
			RequestID: uuid.New().String(),
		},
		Data: tickets,
		Pagination: dto.Pagination{
			CurrentPage:  params.Page,
			TotalRecords: esResponse.Hits.Total.Value,
			PerPage:      params.PageSize,
			TotalPages:   totalPages,
			HasNext:      from+params.PageSize < int(esResponse.Hits.Total.Value),
			HasPrev:      from > 0,
		},
		Message: "200 OK",
	}, nil
}

// SearchTicketByID busca um ticket pelo ticket_id e retorna todas as informações do ticket
func (es *Client) SearchTicketByID(ctx context.Context, ticketID string) (*map[string]interface{}, error) {
	// Montar a query para buscar pelo ticket_id
	query := map[string]interface{}{
		"query": map[string]interface{}{
			"term": map[string]interface{}{
				"ticket_id": ticketID,
			},
		},
		"size": 1,
	}

	queryJSON, err := json.Marshal(query)
	if err != nil {
		return nil, fmt.Errorf("erro ao serializar query: %v", err)
	}

	req := esapi.SearchRequest{
		Index: []string{es.config.IndexName},
		Body:  bytes.NewReader(queryJSON),
	}

	res, err := req.Do(ctx, es.ES)
	if err != nil {
		return nil, fmt.Errorf("erro na execução da busca: %v", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		body, _ := io.ReadAll(res.Body)
		return nil, fmt.Errorf("erro na busca: %s - %s", res.Status(), string(body))
	}

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, fmt.Errorf("erro ao ler resposta: %v", err)
	}

	var esResponse dto.ESResponse
	if err := json.Unmarshal(body, &esResponse); err != nil {
		return nil, fmt.Errorf("erro ao deserializar resposta: %v", err)
	}

	if len(esResponse.Hits.Hits) == 0 {
		return nil, nil // Não encontrado
	}

	var ticket map[string]interface{}
	if err := json.Unmarshal(esResponse.Hits.Hits[0].Source, &ticket); err != nil {
		return nil, fmt.Errorf("erro ao deserializar ticket: %v", err)
	}

	return &ticket, nil
}
