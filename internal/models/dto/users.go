package dto

import "time"

// ============================================
// USER REQUEST DTOs
// ============================================

// CreateUserRequest representa a requisição de criação de usuário
type CreateUserRequest struct {
	Name     string  `json:"name" binding:"required,min=3,max=200" example:"João Silva"`
	Email    string  `json:"email" binding:"required,email,max=255" example:"joao.silva@example.com"`
	Password *string `json:"password,omitempty" binding:"omitempty,min=8,max=100" example:"SenhaSegura@123"`
	UserType string  `json:"userType" binding:"required,oneof=ADMIN MANAGER SUPPORT" example:"SUPPORT" enums:"ADMIN,MANAGER,SUPPORT"`
	// MicrosoftId *string `json:"microsoftId,omitempty" binding:"omitempty,max=255" example:"a1b2c3d4-e5f6-7890-abcd-ef1234567890"`
}

// UpdateUserRequest representa a requisição de atualização de usuário
type UpdateUserRequest struct {
	Name     *string `json:"name,omitempty" binding:"omitempty,min=3,max=200" example:"João Silva Atualizado"`
	Email    *string `json:"email,omitempty" binding:"omitempty,email,max=255" example:"joao.novo@example.com"`
	Password *string `json:"password,omitempty" binding:"omitempty,min=8,max=100" example:"NovaSenha@456"`
	UserType *string `json:"userType,omitempty" binding:"omitempty,oneof=ADMIN MANAGER AGENT VIEWER" example:"MANAGER" enums:"ADMIN,MANAGER,AGENT,VIEWER"`
	IsActive *bool   `json:"isActive,omitempty" example:"true"`
}

// ChangePasswordRequest representa a requisição de mudança de senha
type ChangePasswordRequest struct {
	CurrentPassword string `json:"currentPassword" binding:"required" example:"SenhaAtual@123"`
	NewPassword     string `json:"newPassword" binding:"required,min=8,max=100" example:"NovaSenha@456"`
}

// ============================================
// AUTH REQUEST DTOs
// ============================================

// LoginRequest representa a requisição de login
type LoginRequest struct {
	Email            string `json:"email" binding:"required,email" example:"joao.silva@example.com"`
	Password         string `json:"password" binding:"required" example:"SenhaSegura@123"`
	LoginType        string `json:"login_type" binding:"required,oneof=password microsoft" example:"password"`
	MicrosoftIDToken string `json:"microsoft_id_token,omitempty" example:"eyJhbGciOi..."` // optional for microsoft flow when front handles OAuth; not needed when backend-only
}

// MicrosoftAuthRequest representa a requisição de autenticação Microsoft
type MicrosoftAuthRequest struct {
	AccessToken string `json:"accessToken" binding:"required" example:"EwAoA8l6BAAURSN/FjAGe3BVB..."`
}

// ============================================
// USER RESPONSE DTOs
// ============================================

// UserResponse representa um usuário na resposta
type UserResponse struct {
	Id          int        `json:"id" example:"1"`
	Name        string     `json:"name" example:"João Silva"`
	Email       string     `json:"email" example:"joao.silva@example.com"`
	UserType    string     `json:"userType" example:"AGENT" enums:"ADMIN,MANAGER,AGENT,VIEWER"`
	MicrosoftId *string    `json:"microsoftId,omitempty" example:"a1b2c3d4-e5f6-7890-abcd-ef1234567890"`
	IsActive    bool       `json:"isActive" example:"true"`
	CreatedAt   time.Time  `json:"createdAt" example:"2025-10-16T10:30:00Z"`
	UpdatedAt   *time.Time `json:"updatedAt,omitempty" example:"2025-10-16T15:45:00Z"`
	LastLoginAt *time.Time `json:"lastLoginAt,omitempty" example:"2025-10-16T14:20:00Z"`
}

// UsersListResponse representa a lista de usuários
type UsersListResponse struct {
	Users      []UserResponse `json:"users"`
	TotalCount int            `json:"totalCount" example:"50"`
	Page       int            `json:"page" example:"1"`
	PageSize   int            `json:"pageSize" example:"10"`
}

// UserCreatedResponse representa a resposta de criação de usuário
type UserCreatedResponse struct {
	Id      int    `json:"id" example:"1"`
	Message string `json:"message" example:"User created successfully"`
}

// ============================================
// AUTH RESPONSE DTOs
// ============================================

// LoginResponse representa a resposta de login bem-sucedida
type LoginResponse struct {
	Token     string       `json:"token" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."`
	TokenType string       `json:"token_type" example:"Bearer"`
	ExpiresIn int          `json:"expires_in" example:"3600"`
	ExpiresAt time.Time    `json:"expires_at" example:"2025-10-23T15:30:00Z"`
	User      UserResponse `json:"user"`
}

// UserAuthLogResponse representa um log de autenticação
type UserAuthLogResponse struct {
	Id           int       `json:"id" example:"1"`
	UserId       int       `json:"userId" example:"1"`
	AuthType     string    `json:"authType" example:"JWT" enums:"JWT,MICROSOFT"`
	IPAddress    *string   `json:"ipAddress,omitempty" example:"192.168.1.100"`
	UserAgent    *string   `json:"userAgent,omitempty" example:"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36"`
	Success      bool      `json:"success" example:"true"`
	ErrorMessage *string   `json:"errorMessage,omitempty" example:"Invalid credentials"`
	CreatedAt    time.Time `json:"createdAt" example:"2025-10-16T10:30:00Z"`
}

// UserAuthLogsResponse representa a lista de logs de autenticação
type UserAuthLogsResponse struct {
	Logs       []UserAuthLogResponse `json:"logs"`
	TotalCount int                   `json:"totalCount" example:"25"`
}

// ValidationError representa um erro de validação específico de campo
type ValidationError struct {
	Field   string `json:"field" example:"email"`
	Message string `json:"message" example:"Invalid email format"`
}

// ValidationErrorResponse representa uma resposta com múltiplos erros de validação
type ValidationErrorResponse struct {
	BaseResponse
	Error   string            `json:"error" example:"Validation Failed"`
	Code    int               `json:"code" example:"422"`
	Message string            `json:"message" example:"Request validation failed"`
	Errors  []ValidationError `json:"errors"`
}
