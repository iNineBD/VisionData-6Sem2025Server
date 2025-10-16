package sqlserver

import (
	"context"
	"fmt"
	"orderstreamrest/internal/models/entities"
	"time"

	"gorm.io/gorm"
)

// CreateUser cria um novo usuário
func (s *Internal) CreateUser(ctx context.Context, user *entities.User) (int, error) {
	result := s.db.WithContext(ctx).Table("dbusers.Users").Create(user)
	if result.Error != nil {
		return 0, fmt.Errorf("failed to create user: %w", result.Error)
	}
	return user.Id, nil
}

// GetUserByID busca um usuário por ID
func (s *Internal) GetUserByID(ctx context.Context, id int) (*entities.User, error) {
	var user entities.User
	err := s.db.WithContext(ctx).
		Table("dbusers.Users").
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
		Table("dbusers.Users").
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
		Table("dbusers.Users").
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

	query := s.db.WithContext(ctx).Table("dbusers.Users")

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
		"UpdatedBy": user.UpdatedBy,
	}

	result := s.db.WithContext(ctx).
		Table("dbusers.Users").
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
		Table("dbusers.Users").
		Where("Id = ?", id).
		Updates(map[string]interface{}{
			"PasswordHash": passwordHash,
			"UpdatedAt":    time.Now(),
			"UpdatedBy":    updatedBy,
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
		Table("dbusers.Users").
		Where("Id = ?", id).
		Update("LastLoginAt", time.Now())

	if result.Error != nil {
		return fmt.Errorf("failed to update last login: %w", result.Error)
	}

	return nil
}

// DeleteUser deleta um usuário (soft delete - marca como inativo)
func (s *Internal) DeleteUser(ctx context.Context, id int, deletedBy int) error {
	result := s.db.WithContext(ctx).
		Table("dbusers.Users").
		Where("Id = ?", id).
		Updates(map[string]interface{}{
			"IsActive":  false,
			"UpdatedAt": time.Now(),
			"UpdatedBy": deletedBy,
		})

	if result.Error != nil {
		return fmt.Errorf("failed to delete user: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("user not found")
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

// GetUserAuthLogs retorna os logs de autenticação de um usuário
func (s *Internal) GetUserAuthLogs(ctx context.Context, userId int, limit int) ([]entities.UserAuthLog, error) {
	var logs []entities.UserAuthLog
	err := s.db.WithContext(ctx).
		Table("dbo.UserAuthLogs").
		Where("UserId = ?", userId).
		Order("CreatedAt DESC").
		Limit(limit).
		Find(&logs).Error

	if err != nil {
		return nil, fmt.Errorf("failed to get auth logs: %w", err)
	}

	return logs, nil
}
