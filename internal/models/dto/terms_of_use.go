package dto

import "time"

// TermsOfUseResponse representa a resposta com informações de um termo de uso
type TermsOfUseResponse struct {
	Id            int                `json:"id"`
	Version       string             `json:"version"`
	Title         string             `json:"title"`
	Description   *string            `json:"description,omitempty"`
	Content       string             `json:"content"`
	IsActive      bool               `json:"isActive"`
	EffectiveDate time.Time          `json:"effectiveDate"`
	CreatedAt     time.Time          `json:"createdAt"`
	Items         []TermItemResponse `json:"items,omitempty"`
}

// TermItemResponse representa a resposta com informações de um item de termo
type TermItemResponse struct {
	Id          int    `json:"id"`
	TermId      int    `json:"termId"`
	ItemOrder   int    `json:"itemOrder"`
	Title       string `json:"title"`
	Content     string `json:"content"`
	IsMandatory bool   `json:"isMandatory"`
	IsActive    bool   `json:"isActive"`
}

// CreateTermRequest representa a requisição para criar um novo termo
type CreateTermRequest struct {
	Version       string                  `json:"version" binding:"required"`
	Title         string                  `json:"title" binding:"required"`
	Description   *string                 `json:"description,omitempty"`
	Content       string                  `json:"content" binding:"required"`
	EffectiveDate *time.Time              `json:"effectiveDate,omitempty"`
	Items         []CreateTermItemRequest `json:"items" binding:"required,min=1"`
}

// CreateTermItemRequest representa a requisição para criar um item de termo
type CreateTermItemRequest struct {
	ItemOrder   int    `json:"itemOrder" binding:"required,min=1"`
	Title       string `json:"title" binding:"required"`
	Content     string `json:"content" binding:"required"`
	IsMandatory bool   `json:"isMandatory"`
}

// UpdateTermRequest representa a requisição para atualizar um termo
type UpdateTermRequest struct {
	Title         *string    `json:"title,omitempty"`
	Description   *string    `json:"description,omitempty"`
	Content       *string    `json:"content,omitempty"`
	IsActive      *bool      `json:"isActive,omitempty"`
	EffectiveDate *time.Time `json:"effectiveDate,omitempty"`
}

// UserConsentRequest representa a requisição de consentimento do usuário
type UserConsentRequest struct {
	TermId       int                      `json:"termId" binding:"required"`
	ItemConsents []UserItemConsentRequest `json:"itemConsents" binding:"required,min=1"`
}

// UserItemConsentRequest representa o consentimento para um item específico
type UserItemConsentRequest struct {
	ItemId   int  `json:"itemId" binding:"required"`
	Accepted bool `json:"accepted"`
}

// UserConsentResponse representa a resposta do consentimento registrado
type UserConsentResponse struct {
	Id           int                       `json:"id"`
	UserId       int                       `json:"userId"`
	TermId       int                       `json:"termId"`
	TermVersion  string                    `json:"termVersion"`
	ConsentDate  time.Time                 `json:"consentDate"`
	IsActive     bool                      `json:"isActive"`
	ItemConsents []UserItemConsentResponse `json:"itemConsents,omitempty"`
}

// UserItemConsentResponse representa a resposta do consentimento de item
type UserItemConsentResponse struct {
	ItemId      int    `json:"itemId"`
	ItemTitle   string `json:"itemTitle"`
	Accepted    bool   `json:"accepted"`
	IsMandatory bool   `json:"isMandatory"`
}

// RevokeConsentRequest representa a requisição para revogar consentimento
type RevokeConsentRequest struct {
	TermId int     `json:"termId" binding:"required"`
	Reason *string `json:"reason,omitempty"`
}

// UserConsentStatusResponse representa o status de consentimento do usuário
type UserConsentStatusResponse struct {
	UserId             int        `json:"userId"`
	HasActiveConsent   bool       `json:"hasActiveConsent"`
	CurrentTermId      *int       `json:"currentTermId,omitempty"`
	CurrentTermVersion *string    `json:"currentTermVersion,omitempty"`
	CurrentTermTitle   *string    `json:"currentTermTitle,omitempty"`
	ConsentDate        *time.Time `json:"consentDate,omitempty"`
	NeedsNewConsent    bool       `json:"needsNewConsent"`
}

// MyConsentStatusResponse representa o status completo de consentimento do usuário com termo e itens
type MyConsentStatusResponse struct {
	UserId           int                 `json:"userId"`
	HasActiveConsent bool                `json:"hasActiveConsent"`
	NeedsNewConsent  bool                `json:"needsNewConsent"`
	Term             *TermsOfUseResponse `json:"term,omitempty"`
	ConsentDate      *time.Time          `json:"consentDate,omitempty"`
}

// ListTermsResponse representa a lista de termos (para admin)
type ListTermsResponse struct {
	Terms      []TermsOfUseResponse `json:"terms"`
	TotalCount int                  `json:"totalCount"`
	Page       int                  `json:"page"`
	PageSize   int                  `json:"pageSize"`
}

// TermStatisticsResponse representa estatísticas de consentimento de um termo
type TermStatisticsResponse struct {
	TermId              int       `json:"termId"`
	TermVersion         string    `json:"termVersion"`
	TotalUsers          int       `json:"totalUsers"`
	UsersWithConsent    int       `json:"usersWithConsent"`
	UsersWithoutConsent int       `json:"usersWithoutConsent"`
	ConsentRate         float64   `json:"consentRate"` // Porcentagem
	LastUpdate          time.Time `json:"lastUpdate"`
}
