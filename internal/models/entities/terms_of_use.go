package entities

import "time"

// TermsOfUse representa um termo de uso com versionamento
type TermsOfUse struct {
	Id            int        `json:"id" gorm:"column:Id;primaryKey;autoIncrement"`
	Version       string     `json:"version" gorm:"column:Version;type:nvarchar(50);not null;unique"`
	Content       string     `json:"content" gorm:"column:Content;type:text;not null"`
	Title         string     `json:"title" gorm:"column:Title;type:nvarchar(500);not null"`
	Description   *string    `json:"description,omitempty" gorm:"column:Description;type:nvarchar(max)"`
	IsActive      bool       `json:"isActive" gorm:"column:IsActive;type:bit;not null;default:1"`
	EffectiveDate time.Time  `json:"effectiveDate" gorm:"column:EffectiveDate;type:datetime2;not null;default:GETDATE()"`
	CreatedAt     time.Time  `json:"createdAt" gorm:"column:CreatedAt;type:datetime2;not null;default:GETDATE()"`
	CreatedBy     *int       `json:"createdBy,omitempty" gorm:"column:CreatedBy;type:int"`
	UpdatedAt     *time.Time `json:"updatedAt,omitempty" gorm:"column:UpdatedAt;type:datetime2"`
	UpdatedBy     *int       `json:"updatedBy,omitempty" gorm:"column:UpdatedBy;type:int"`

	// Relacionamentos
	Items []TermItem `json:"items,omitempty" gorm:"foreignKey:TermId"`
}

// TableName especifica o nome da tabela no banco
func (TermsOfUse) TableName() string {
	return "dbo.TermsOfUse"
}

// TermItem representa um item de um termo (obrigatório ou opcional)
type TermItem struct {
	Id          int        `json:"id" gorm:"column:Id;primaryKey;autoIncrement"`
	TermId      int        `json:"termId" gorm:"column:TermId;type:int;not null"`
	ItemOrder   int        `json:"itemOrder" gorm:"column:ItemOrder;type:int;not null;default:1"`
	Title       string     `json:"title" gorm:"column:Title;type:nvarchar(500);not null"`
	Content     string     `json:"content" gorm:"column:Content;type:text;not null"`
	IsMandatory bool       `json:"isMandatory" gorm:"column:IsMandatory;type:bit;not null;default:0"`
	IsActive    bool       `json:"isActive" gorm:"column:IsActive;type:bit;not null;default:1"`
	CreatedAt   time.Time  `json:"createdAt" gorm:"column:CreatedAt;type:datetime2;not null;default:GETDATE()"`
	UpdatedAt   *time.Time `json:"updatedAt,omitempty" gorm:"column:UpdatedAt;type:datetime2"`
}

// TableName especifica o nome da tabela no banco
func (TermItem) TableName() string {
	return "dbo.TermItems"
}

// UserTermConsent representa o consentimento de um usuário para um termo
type UserTermConsent struct {
	Id            int        `json:"id" gorm:"column:Id;primaryKey;autoIncrement"`
	UserId        int        `json:"userId" gorm:"column:UserId;type:int;not null"`
	TermId        int        `json:"termId" gorm:"column:TermId;type:int;not null"`
	ConsentDate   time.Time  `json:"consentDate" gorm:"column:ConsentDate;type:datetime2;not null;default:GETDATE()"`
	IsActive      bool       `json:"isActive" gorm:"column:IsActive;type:bit;not null;default:1"`
	IPAddress     *string    `json:"ipAddress,omitempty" gorm:"column:IPAddress;type:nvarchar(50)"`
	UserAgent     *string    `json:"userAgent,omitempty" gorm:"column:UserAgent;type:nvarchar(500)"`
	RevokedAt     *time.Time `json:"revokedAt,omitempty" gorm:"column:RevokedAt;type:datetime2"`
	RevokedReason *string    `json:"revokedReason,omitempty" gorm:"column:RevokedReason;type:nvarchar(max)"`

	// Relacionamentos
	User       User              `json:"user,omitempty" gorm:"foreignKey:UserId"`
	Term       TermsOfUse        `json:"term,omitempty" gorm:"foreignKey:TermId"`
	ItemConsents []UserItemConsent `json:"itemConsents,omitempty" gorm:"foreignKey:UserConsentId"`
}

// TableName especifica o nome da tabela no banco
func (UserTermConsent) TableName() string {
	return "dbo.UserTermConsents"
}

// UserItemConsent representa o consentimento de um usuário para um item específico
type UserItemConsent struct {
	Id            int       `json:"id" gorm:"column:Id;primaryKey;autoIncrement"`
	UserConsentId int       `json:"userConsentId" gorm:"column:UserConsentId;type:int;not null"`
	ItemId        int       `json:"itemId" gorm:"column:ItemId;type:int;not null"`
	Accepted      bool      `json:"accepted" gorm:"column:Accepted;type:bit;not null"`
	ConsentDate   time.Time `json:"consentDate" gorm:"column:ConsentDate;type:datetime2;not null;default:GETDATE()"`

	// Relacionamentos
	Item TermItem `json:"item,omitempty" gorm:"foreignKey:ItemId"`
}

// TableName especifica o nome da tabela no banco
func (UserItemConsent) TableName() string {
	return "dbo.UserItemConsents"
}
