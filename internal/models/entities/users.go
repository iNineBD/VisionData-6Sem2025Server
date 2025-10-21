package entities

import "time"

// User representa um usuário do sistema
type User struct {
	Id           int        `json:"id" gorm:"column:Id;primaryKey;autoIncrement"`
	Name         string     `json:"name" gorm:"column:Name;type:nvarchar(200);not null"`
	Email        string     `json:"email" gorm:"column:Email;type:nvarchar(255);not null;unique"`
	PasswordHash *string    `json:"-" gorm:"column:PasswordHash;type:nvarchar(500)"` // Nunca retornar no JSON
	UserType     string     `json:"userType" gorm:"column:UserType;type:nvarchar(50);not null"`
	MicrosoftId  *string    `json:"microsoftId,omitempty" gorm:"column:MicrosoftId;type:nvarchar(255);unique"`
	IsActive     bool       `json:"isActive" gorm:"column:IsActive;type:bit;not null;default:1"`
	CreatedAt    time.Time  `json:"createdAt" gorm:"column:CreatedAt;type:datetime2;not null;default:GETDATE()"`
	UpdatedAt    *time.Time `json:"updatedAt,omitempty" gorm:"column:UpdatedAt;type:datetime2"`
	LastLoginAt  *time.Time `json:"lastLoginAt,omitempty" gorm:"column:LastLoginAt;type:datetime2"`
	CreatedBy    *int       `json:"createdBy,omitempty" gorm:"column:CreatedBy;type:int"`
	UpdatedBy    *int       `json:"updatedBy,omitempty" gorm:"column:UpdatedBy;type:int"`
}

// TableName especifica o nome da tabela no banco
func (User) TableName() string {
	return "dbo.Users"
}

// UserAuthLog representa um log de autenticação
type UserAuthLog struct {
	Id           int       `json:"id" gorm:"column:Id;primaryKey;autoIncrement"`
	UserId       int       `json:"userId" gorm:"column:UserId;type:int;not null"`
	AuthType     string    `json:"authType" gorm:"column:AuthType;type:nvarchar(50);not null"`
	IPAddress    *string   `json:"ipAddress,omitempty" gorm:"column:IPAddress;type:nvarchar(50)"`
	UserAgent    *string   `json:"userAgent,omitempty" gorm:"column:UserAgent;type:nvarchar(500)"`
	Success      bool      `json:"success" gorm:"column:Success;type:bit;not null"`
	ErrorMessage *string   `json:"errorMessage,omitempty" gorm:"column:ErrorMessage;type:nvarchar(500)"`
	CreatedAt    time.Time `json:"createdAt" gorm:"column:CreatedAt;type:datetime2;not null;default:GETDATE()"`
}

// TableName especifica o nome da tabela no banco
func (UserAuthLog) TableName() string {
	return "dbo.UserAuthLogs"
}
