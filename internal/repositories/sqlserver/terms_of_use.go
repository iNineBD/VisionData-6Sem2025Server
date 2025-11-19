package sqlserver

import (
	"context"
	"errors"
	"fmt"
	"orderstreamrest/internal/models/entities"
	"time"

	"gorm.io/gorm"
)

// GetActiveTermWithItems retorna o termo ativo baseado na data de vigência
func (i *Internal) GetActiveTermWithItems(ctx context.Context) (*entities.TermsOfUse, error) {
	var term entities.TermsOfUse

	// Buscar o termo que está marcado como ativo
	// A flag IsActive é gerenciada automaticamente baseada na data de vigência
	err := i.db.WithContext(ctx).
		Where("IsActive = ?", true).
		First(&term).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("nenhum termo ativo encontrado")
		}
		return nil, err
	}

	// Buscar itens do termo
	err = i.db.WithContext(ctx).
		Where("TermId = ? AND IsActive = ?", term.Id, true).
		Order("ItemOrder ASC, Id ASC").
		Find(&term.Items).Error

	if err != nil {
		return nil, err
	}

	return &term, nil
}

// GetTermByID retorna um termo específico com seus itens
func (i *Internal) GetTermByID(ctx context.Context, termId int) (*entities.TermsOfUse, error) {
	var term entities.TermsOfUse

	err := i.db.WithContext(ctx).
		Where("Id = ?", termId).
		First(&term).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("termo não encontrado")
		}
		return nil, err
	}

	// Buscar itens do termo
	err = i.db.WithContext(ctx).
		Where("TermId = ? AND IsActive = ?", term.Id, true).
		Order("ItemOrder ASC, Id ASC").
		Find(&term.Items).Error

	if err != nil {
		return nil, err
	}

	return &term, nil
}

// GetTermByVersion retorna um termo específico pela versão
func (i *Internal) GetTermByVersion(ctx context.Context, version string) (*entities.TermsOfUse, error) {
	var term entities.TermsOfUse

	err := i.db.WithContext(ctx).
		Where("Version = ?", version).
		First(&term).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("termo com versão %s não encontrado", version)
		}
		return nil, err
	}

	// Buscar itens do termo
	err = i.db.WithContext(ctx).
		Where("TermId = ? AND IsActive = ?", term.Id, true).
		Order("ItemOrder ASC, Id ASC").
		Find(&term.Items).Error

	if err != nil {
		return nil, err
	}

	return &term, nil
}

// ListAllTerms retorna todos os termos (para admin)
func (i *Internal) ListAllTerms(ctx context.Context, page, pageSize int) ([]entities.TermsOfUse, int64, error) {
	var terms []entities.TermsOfUse
	var total int64

	// Contar total
	err := i.db.WithContext(ctx).
		Model(&entities.TermsOfUse{}).
		Count(&total).Error

	if err != nil {
		return nil, 0, err
	}

	// Buscar termos com paginação
	offset := (page - 1) * pageSize
	err = i.db.WithContext(ctx).
		Order("Id ASC").
		Limit(pageSize).
		Offset(offset).
		Find(&terms).Error

	if err != nil {
		return nil, 0, err
	}

	// Buscar itens de cada termo
	for idx := range terms {
		err = i.db.WithContext(ctx).
			Where("TermId = ? AND IsActive = ?", terms[idx].Id, true).
			Order("ItemOrder ASC, Id ASC").
			Find(&terms[idx].Items).Error

		if err != nil {
			return nil, 0, err
		}
	}

	return terms, total, nil
}

// CreateTerm cria um novo termo com seus itens
func (i *Internal) CreateTerm(ctx context.Context, term *entities.TermsOfUse) error {
	return i.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Verificar se versão já existe
		var count int64
		err := tx.Model(&entities.TermsOfUse{}).
			Where("Version = ?", term.Version).
			Count(&count).Error

		if err != nil {
			return err
		}

		if count > 0 {
			return fmt.Errorf("já existe um termo com a versão %s", term.Version)
		}

		// Verificar se há apenas 1 item obrigatório
		mandatoryCount := 0
		for _, item := range term.Items {
			if item.IsMandatory {
				mandatoryCount++
			}
		}

		if mandatoryCount == 0 {
			return fmt.Errorf("é necessário ter pelo menos 1 item obrigatório")
		}

		if mandatoryCount > 1 {
			return fmt.Errorf("cada termo pode ter apenas 1 item obrigatório")
		}

		// Salvar os itens temporariamente
		items := term.Items
		term.Items = nil

		// Criar termo sem os itens (evita problema com OUTPUT clause)
		if err := tx.Omit("Items").Create(term).Error; err != nil {
			return err
		}

		// Agora criar os itens separadamente usando SQL direto para evitar OUTPUT clause
		for idx := range items {
			items[idx].TermId = term.Id

			// Inserir usando SQL direto para evitar problema com triggers
			result := tx.Exec(`
				INSERT INTO dbo.TermItems (TermId, ItemOrder, Title, Content, IsMandatory, IsActive, CreatedAt)
				VALUES (?, ?, ?, ?, ?, ?, GETDATE())
			`, items[idx].TermId, items[idx].ItemOrder, items[idx].Title,
				items[idx].Content, items[idx].IsMandatory, items[idx].IsActive)

			if result.Error != nil {
				return result.Error
			}
		}

		// Buscar os itens criados para retornar com IDs
		if err := tx.Where("TermId = ?", term.Id).Find(&items).Error; err != nil {
			return err
		}

		// Restaurar os itens no objeto original
		term.Items = items

		// Recalcular qual termo deve estar ativo baseado na data de vigência
		// Garantir que sempre exista pelo menos 1 termo ativo
		now := time.Now()

		// Buscar todos os termos ativos no momento
		var activeTerms []entities.TermsOfUse
		err = tx.Model(&entities.TermsOfUse{}).
			Where("IsActive = ?", true).
			Find(&activeTerms).Error

		if err != nil {
			return fmt.Errorf("falha ao buscar termos ativos: %w", err)
		}

		// Se já existe algum termo ativo, comparar datas de vigência
		if len(activeTerms) > 0 {
			// Desativar todos os termos primeiro
			err = tx.Model(&entities.TermsOfUse{}).
				Where("1 = 1").
				Update("IsActive", false).Error

			if err != nil {
				return fmt.Errorf("falha ao desativar termos: %w", err)
			}

			// Encontrar o termo que deve estar ativo:
			// - Data de vigência <= agora (já entrou em vigor)
			// - Ordenado por data de vigência DESC (mais recente primeiro)
			// Usar CAST para comparar apenas a data sem hora
			var activeTermId int
			err = tx.Model(&entities.TermsOfUse{}).
				Select("Id").
				Where("CAST(EffectiveDate AS DATE) <= CAST(? AS DATE)", now).
				Order("EffectiveDate DESC, CreatedAt DESC").
				Limit(1).
				Pluck("Id", &activeTermId).Error

			if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
				return fmt.Errorf("falha ao buscar termo ativo: %w", err)
			}

			// Se encontrou um termo válido, ativá-lo
			if activeTermId > 0 {
				err = tx.Model(&entities.TermsOfUse{}).
					Where("Id = ?", activeTermId).
					Update("IsActive", true).Error

				if err != nil {
					return fmt.Errorf("falha ao ativar termo: %w", err)
				}

				// Atualizar o objeto term se for o termo que acabou de ser criado
				if term.Id == activeTermId {
					term.IsActive = true
				} else {
					term.IsActive = false
				}
			} else {
				// Nenhum termo com data válida encontrado, ativar o mais recente de todos
				err = tx.Model(&entities.TermsOfUse{}).
					Select("Id").
					Order("CreatedAt DESC").
					Limit(1).
					Pluck("Id", &activeTermId).Error

				if err != nil {
					return fmt.Errorf("falha ao buscar termo mais recente: %w", err)
				}

				if activeTermId > 0 {
					err = tx.Model(&entities.TermsOfUse{}).
						Where("Id = ?", activeTermId).
						Update("IsActive", true).Error

					if err != nil {
						return fmt.Errorf("falha ao ativar termo: %w", err)
					}

					term.IsActive = (term.Id == activeTermId)
				}
			}
		}
		// Se não existe nenhum termo ativo, o novo termo permanece ativo (já foi criado com IsActive = true)

		return nil
	})
}

// UpdateTerm atualiza um termo existente
func (i *Internal) UpdateTerm(ctx context.Context, termId int, updates map[string]interface{}) error {
	result := i.db.WithContext(ctx).
		Model(&entities.TermsOfUse{}).
		Where("Id = ?", termId).
		Updates(updates)

	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("termo não encontrado")
	}

	return nil
}

// ============================================================================
// CONSENTIMENTOS
// ============================================================================

// GetUserActiveConsent retorna o consentimento ativo do usuário
func (i *Internal) GetUserActiveConsent(ctx context.Context, userId int) (*entities.UserTermConsent, error) {
	var consent entities.UserTermConsent

	err := i.db.WithContext(ctx).
		Where("UserId = ? AND IsActive = ?", userId, true).
		Order("ConsentDate DESC").
		First(&consent).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil // Não é erro, apenas não tem consentimento
		}
		return nil, err
	}

	// Buscar termo relacionado
	err = i.db.WithContext(ctx).
		Where("Id = ?", consent.TermId).
		First(&consent.Term).Error

	if err != nil {
		return nil, err
	}

	// Buscar consentimentos dos itens
	err = i.db.WithContext(ctx).
		Preload("Item").
		Where("UserConsentId = ?", consent.Id).
		Find(&consent.ItemConsents).Error

	if err != nil {
		return nil, err
	}

	return &consent, nil
}

// CheckUserHasMandatoryConsent verifica se usuário aceitou o item obrigatório
func (i *Internal) CheckUserHasMandatoryConsent(ctx context.Context, userId, termId int) (bool, error) {
	var count int64

	err := i.db.WithContext(ctx).
		Table("dbo.UserTermConsents utc").
		Joins("INNER JOIN dbo.UserItemConsents uic ON utc.Id = uic.UserConsentId").
		Joins("INNER JOIN dbo.TermItems ti ON uic.ItemId = ti.Id").
		Where("utc.UserId = ? AND utc.TermId = ? AND utc.IsActive = ? AND ti.IsMandatory = ? AND uic.Accepted = ?",
			userId, termId, true, true, true).
		Count(&count).Error

	if err != nil {
		return false, err
	}

	return count > 0, nil
}

// RegisterUserConsent registra o consentimento completo do usuário
func (i *Internal) RegisterUserConsent(ctx context.Context, consent *entities.UserTermConsent) error {
	return i.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Desativar consentimentos anteriores para o mesmo termo
		err := tx.Model(&entities.UserTermConsent{}).
			Where("UserId = ? AND TermId = ? AND IsActive = ?", consent.UserId, consent.TermId, true).
			Updates(map[string]interface{}{
				"IsActive":      false,
				"RevokedAt":     time.Now(),
				"RevokedReason": "Nova versão aceita",
			}).Error

		if err != nil {
			return err
		}

		// Criar novo consentimento
		if err := tx.Create(consent).Error; err != nil {
			return err
		}

		// Verificar se o item obrigatório foi aceito
		hasMandatory := false
		for _, itemConsent := range consent.ItemConsents {
			// Buscar informações do item
			var item entities.TermItem
			err := tx.Where("Id = ?", itemConsent.ItemId).First(&item).Error
			if err != nil {
				return err
			}

			if item.IsMandatory && itemConsent.Accepted {
				hasMandatory = true
				break
			}
		}

		if !hasMandatory {
			return fmt.Errorf("o item obrigatório do termo não foi aceito")
		}

		return nil
	})
}

// RevokeUserConsent revoga o consentimento do usuário
func (i *Internal) RevokeUserConsent(ctx context.Context, userId, termId int, reason string) error {
	result := i.db.WithContext(ctx).
		Model(&entities.UserTermConsent{}).
		Where("UserId = ? AND TermId = ? AND IsActive = ?", userId, termId, true).
		Updates(map[string]interface{}{
			"IsActive":      false,
			"RevokedAt":     time.Now(),
			"RevokedReason": reason,
		})

	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("nenhum consentimento ativo encontrado para revogar")
	}

	return nil
}

// RecalculateActiveTerm recalcula qual termo deve estar ativo baseado na data de vigência
// Útil para ser chamado por schedulers ou tarefas de manutenção
func (i *Internal) RecalculateActiveTerm(ctx context.Context) error {
	return i.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		now := time.Now()

		// Desativar todos os termos
		err := tx.Model(&entities.TermsOfUse{}).
			Where("1 = 1").
			Update("IsActive", false).Error

		if err != nil {
			return fmt.Errorf("falha ao desativar termos: %w", err)
		}

		// Encontrar o termo que deve estar ativo:
		// - Data de vigência <= agora (já entrou em vigor)
		// - Ordenado por data de vigência DESC (mais recente primeiro)
		// Usar CAST para comparar apenas a data sem hora
		var activeTermId int
		err = tx.Model(&entities.TermsOfUse{}).
			Select("Id").
			Where("CAST(EffectiveDate AS DATE) <= CAST(? AS DATE)", now).
			Order("EffectiveDate DESC, CreatedAt DESC").
			Limit(1).
			Pluck("Id", &activeTermId).Error

		if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			return fmt.Errorf("falha ao buscar termo ativo: %w", err)
		}

		// Se encontrou um termo válido, ativá-lo
		if activeTermId > 0 {
			err = tx.Model(&entities.TermsOfUse{}).
				Where("Id = ?", activeTermId).
				Update("IsActive", true).Error

			if err != nil {
				return fmt.Errorf("falha ao ativar termo: %w", err)
			}
		}

		return nil
	})
}

// GetUserConsentHistory retorna o histórico de consentimentos do usuário
func (i *Internal) GetUserConsentHistory(ctx context.Context, userId int) ([]entities.UserTermConsent, error) {
	var consents []entities.UserTermConsent

	err := i.db.WithContext(ctx).
		Preload("Term").
		Preload("ItemConsents.Item").
		Where("UserId = ?", userId).
		Order("ConsentDate DESC").
		Find(&consents).Error

	if err != nil {
		return nil, err
	}

	return consents, nil
}
