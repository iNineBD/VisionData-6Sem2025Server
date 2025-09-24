package dto

// TicketsMetricsResponse representa a resposta das métricas de tickets
type MetricValue struct {
	Name  string `json:"name"`
	Value int64  `json:"value"`
}

// Metric representa uma métrica com seus valores
type TypeMetric struct {
	Name   string        `json:"name"`
	Values []MetricValue `json:"values"`
}

// MetricValue representa um valor individual de métrica
type TicketsMetricsResponse struct {
	TotalTickets int64        `json:"totalTickets"`
	Metrics      []TypeMetric `json:"metrics"`
}
