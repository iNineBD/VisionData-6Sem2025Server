package dto

type Ticket struct {
	AssignedAgent AssignedAgent `json:"assigned_agent,omitempty"`
	Attachments   []interface{} `json:"attachments,omitempty"`
	AuditLogs     []interface{} `json:"audit_logs,omitempty"`
	Category      Category      `json:"category,omitempty"`
	Channel       string        `json:"channel,omitempty"`
	Company       Company       `json:"company,omitempty"`
	CreatedByUser CreatedByUser `json:"created_by_user,omitempty"`
	CurrentStatus int64         `json:"current_status,omitempty"`
	Dates         Dates         `json:"dates,omitempty"`
	Description   string        `json:"description,omitempty"`
	Device        string        `json:"device,omitempty"`
	Priority      string        `json:"priority,omitempty"`
	Product       Product       `json:"product,omitempty"`
	SearchText    string        `json:"search_text,omitempty"`
	SLAMetrics    SLAMetrics    `json:"sla_metrics,omitempty"`
	SLAPlan       int64         `json:"sla_plan,omitempty"`
	StatusHistory []interface{} `json:"status_history,omitempty"`
	Subcategory   Category      `json:"subcategory,omitempty"`
	Tags          []interface{} `json:"tags,omitempty"`
	TicketID      string        `json:"ticket_id,omitempty"`
	Title         string        `json:"title,omitempty"`
}

type AssignedAgent struct {
	Department int64  `json:"department,omitempty"`
	Email      string `json:"email,omitempty"`
	FullName   string `json:"full_name,omitempty"`
	ID         int64  `json:"id,omitempty"`
}

type Category struct {
	ID   int64  `json:"id,omitempty"`
	Name string `json:"name,omitempty"`
}

type Company struct {
	Cnpj    string `json:"cnpj,omitempty"`
	ID      int64  `json:"id,omitempty"`
	Name    string `json:"name,omitempty"`
	Segment string `json:"segment,omitempty"`
}

type CreatedByUser struct {
	Cpf      string `json:"cpf,omitempty"`
	Email    string `json:"email,omitempty"`
	FullName string `json:"full_name,omitempty"`
	ID       int64  `json:"id,omitempty"`
	IsVip    bool   `json:"is_vip,omitempty"`
	Phone    string `json:"phone,omitempty"`
}

type Dates struct {
	ClosedAt        interface{} `json:"closed_at,omitempty"`
	CreatedAt       interface{} `json:"created_at,omitempty"`
	FirstResponseAt interface{} `json:"first_response_at,omitempty"`
}

type Product struct {
	Code        int64  `json:"code,omitempty"`
	Description string `json:"description,omitempty"`
	ID          int64  `json:"id,omitempty"`
	Name        string `json:"name,omitempty"`
}

type SLAMetrics struct {
	FirstResponseSLABreached bool        `json:"first_response_sla_breached,omitempty"`
	FirstResponseTimeMinutes interface{} `json:"first_response_time_minutes,omitempty"`
	ResolutionSLABreached    bool        `json:"resolution_sla_breached,omitempty"`
	ResolutionTimeMinutes    interface{} `json:"resolution_time_minutes,omitempty"`
}
