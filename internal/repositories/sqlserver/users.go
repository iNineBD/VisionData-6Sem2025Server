package sqlserver

import (
	"context"
	"fmt"
	"orderstreamrest/internal/models/entities"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// CreateUser cria um novo usuário
func (s *Internal) CreateUser(ctx context.Context, user *entities.User) (int, error) {
	result := s.db.WithContext(ctx).Table("dbo.tb_users").Create(user)
	if result.Error != nil {
		return 0, fmt.Errorf("failed to create user: %w", result.Error)
	}
	return user.Id, nil
}

// GetUserByID busca um usuário por ID
func (s *Internal) GetUserByID(ctx context.Context, id int) (*entities.User, error) {
	var user entities.User
	err := s.db.WithContext(ctx).
		Table("dbo.tb_users").
		Where("Id = ?", id).
		First(&user).Error

	if err == gorm.ErrRecordNotFound {
		return nil, fmt.Errorf("user not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	return &user, nil
}

// GetUserByEmail busca um usuário por email
func (s *Internal) GetUserByEmail(ctx context.Context, email string) (*entities.User, error) {
	var user entities.User
	err := s.db.WithContext(ctx).
		Table("dbo.tb_users").
		Where("Email = ?", email).
		First(&user).Error

	if err == gorm.ErrRecordNotFound {
		return nil, fmt.Errorf("user not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	return &user, nil
}

// GetUserByMicrosoftID busca um usuário por Microsoft ID
func (s *Internal) GetUserByMicrosoftID(ctx context.Context, microsoftId string) (*entities.User, error) {
	var user entities.User
	err := s.db.WithContext(ctx).
		Table("dbo.tb_users").
		Where("MicrosoftId = ?", microsoftId).
		First(&user).Error

	if err == gorm.ErrRecordNotFound {
		return nil, fmt.Errorf("user not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	return &user, nil
}

// GetAllUsers retorna todos os usuários com paginação
func (s *Internal) GetAllUsers(ctx context.Context, page, pageSize int, onlyActive bool) ([]entities.User, int64, error) {
	offset := (page - 1) * pageSize

	query := s.db.WithContext(ctx).Table("dbo.tb_users")

	if onlyActive {
		query = query.Where("IsActive = ?", true)
	}

	// Contar total
	var totalCount int64
	if err := query.Count(&totalCount).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count users: %w", err)
	}

	// Buscar usuários
	var users []entities.User
	err := query.
		Order("CreatedAt DESC").
		Offset(offset).
		Limit(pageSize).
		Find(&users).Error

	if err != nil {
		return nil, 0, fmt.Errorf("failed to get users: %w", err)
	}

	return users, totalCount, nil
}

// UpdateUser atualiza um usuário
func (s *Internal) UpdateUser(ctx context.Context, id int, user *entities.User) error {
	updates := map[string]interface{}{
		"Name":      user.Name,
		"Email":     user.Email,
		"UserType":  user.UserType,
		"IsActive":  user.IsActive,
		"UpdatedAt": time.Now(),
	}

	result := s.db.WithContext(ctx).
		Table("dbo.tb_users").
		Where("Id = ?", id).
		Updates(updates)

	if result.Error != nil {
		return fmt.Errorf("failed to update user: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("user not found")
	}

	return nil
}

// UpdatePassword atualiza a senha de um usuário
func (s *Internal) UpdatePassword(ctx context.Context, id int, passwordHash string, updatedBy int) error {
	result := s.db.WithContext(ctx).
		Table("dbo.tb_users").
		Where("Id = ?", id).
		Updates(map[string]interface{}{
			"PasswordHash": passwordHash,
			"UpdatedAt":    time.Now(),
		})

	if result.Error != nil {
		return fmt.Errorf("failed to update password: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("user not found")
	}

	return nil
}

// UpdateLastLogin atualiza o último login do usuário
func (s *Internal) UpdateLastLogin(ctx context.Context, id int) error {
	result := s.db.WithContext(ctx).
		Table("dbo.tb_users").
		Where("Id = ?", id).
		Update("LastLoginAt", time.Now())

	if result.Error != nil {
		return fmt.Errorf("failed to update last login: %w", result.Error)
	}

	return nil
}

/*
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
}*/

// DeleteUser deleta um usuário (soft delete - marca como inativo)
func (s *Internal) DeleteUser(ctx context.Context, id int, deletedBy int) error {
	result := s.db.WithContext(ctx).
		Table("dbo.tb_users").
		Where("Id = ?", id).
		Updates(map[string]interface{}{
			"IsActive":     false,
			"UpdatedAt":    time.Now(),
			"Name":         " - ",
			"Email":        uuid.New().String() + "@deleted.local",
			"PasswordHash": uuid.New().String() + "@deleted.local",
			"MicrosoftId":  uuid.New().String() + "@deleted.local",
			"UserType":     " - ",
		})

	if result.Error != nil {
		return fmt.Errorf("failed to delete user: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("user not found")
	}

	// Verifica se já existe o log LGPD para o usuário
	var count int64
	err := s.db_bkp.WithContext(ctx).
		Table("dbo.Log_LGPD").
		Where("UserId = ?", id).
		Count(&count).Error

	if err != nil {
		return fmt.Errorf("failed to check LGPD log: %w", err)
	}

	if count == 0 {
		result2 := s.db_bkp.WithContext(ctx).
			Table("dbo.Log_LGPD").
			Create(map[string]interface{}{
				"UserId": id,
			})

		if result2.Error != nil {
			return fmt.Errorf("failed to create LGPD log: %w", result2.Error)
		}
	}

	return nil
}

// CreateAuthLog cria um log de autenticação
func (s *Internal) CreateAuthLog(ctx context.Context, log *entities.UserAuthLog) error {
	result := s.db.WithContext(ctx).
		Table("dbo.UserAuthLogs").
		Create(log)

	if result.Error != nil {
		return fmt.Errorf("failed to create auth log: %w", result.Error)
	}

	return nil
}
