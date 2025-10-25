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

type MeanTimeByPriority struct {
	PriorityName string  `json:"priorityName"`
	MeanTimeHour float64 `json:"meanTimeHour"`
	MeanTimeDay  float64 `json:"meanTimeDay"`
}

// MonthlyCounts representa a contagem de tickets para cada mês.
type MonthlyCounts struct {
	Janeiro   int64 `json:"janeiro"`
	Fevereiro int64 `json:"fevereiro"`
	Marco     int64 `json:"marco"`
	Abril     int64 `json:"abril"`
	Maio      int64 `json:"maio"`
	Junho     int64 `json:"junho"`
	Julho     int64 `json:"julho"`
	Agosto    int64 `json:"agosto"`
	Setembro  int64 `json:"setembro"`
	Outubro   int64 `json:"outubro"`
	Novembro  int64 `json:"novembro"`
	Dezembro  int64 `json:"dezembro"`
}

// YearlyData é um mapa de anos (string) para uma lista de contagens mensais.
type YearlyData map[string][]MonthlyCounts

// TicketsByStatusYearMonth é um mapa de status (string) para seus dados anuais.
type TicketsByStatusYearMonth map[string]YearlyData

type Months struct {
	Month string `json:"month"`
	Total int64  `json:"total"`
}

type TicketsByMonth struct {
	Ano          int   `json:"ano"`
	Mes          int   `json:"mes"`
	TotalTickets int64 `json:"totalTickets"`
}
