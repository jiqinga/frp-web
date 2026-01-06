package repository

import (
	"frp-web-panel/internal/model"
	"frp-web-panel/pkg/database"

	"gorm.io/gorm"
)

type AlertRecipientRepo struct{}

func NewAlertRecipientRepo() *AlertRecipientRepo {
	return &AlertRecipientRepo{}
}

// Recipient CRUD
func (r *AlertRecipientRepo) CreateRecipient(recipient *model.AlertRecipient) error {
	return database.DB.Create(recipient).Error
}

func (r *AlertRecipientRepo) GetAllRecipients() ([]model.AlertRecipient, error) {
	var recipients []model.AlertRecipient
	err := database.DB.Order("id desc").Find(&recipients).Error
	return recipients, err
}

func (r *AlertRecipientRepo) GetRecipientByID(id uint) (*model.AlertRecipient, error) {
	var recipient model.AlertRecipient
	err := database.DB.First(&recipient, id).Error
	return &recipient, err
}

func (r *AlertRecipientRepo) GetRecipientsByIDs(ids []uint) ([]model.AlertRecipient, error) {
	var recipients []model.AlertRecipient
	if len(ids) == 0 {
		return recipients, nil
	}
	err := database.DB.Where("id IN ? AND enabled = ?", ids, true).Find(&recipients).Error
	return recipients, err
}

func (r *AlertRecipientRepo) UpdateRecipient(recipient *model.AlertRecipient) error {
	return database.DB.Save(recipient).Error
}

func (r *AlertRecipientRepo) DeleteRecipient(id uint) error {
	return database.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("recipient_id = ?", id).Delete(&model.AlertGroupRecipient{}).Error; err != nil {
			return err
		}
		return tx.Delete(&model.AlertRecipient{}, id).Error
	})
}

// Group CRUD
func (r *AlertRecipientRepo) CreateGroup(group *model.AlertRecipientGroup) error {
	return database.DB.Create(group).Error
}

func (r *AlertRecipientRepo) GetAllGroups() ([]model.AlertRecipientGroup, error) {
	var groups []model.AlertRecipientGroup
	err := database.DB.Order("id desc").Find(&groups).Error
	return groups, err
}

func (r *AlertRecipientRepo) GetGroupByID(id uint) (*model.AlertRecipientGroup, error) {
	var group model.AlertRecipientGroup
	err := database.DB.First(&group, id).Error
	return &group, err
}

func (r *AlertRecipientRepo) UpdateGroup(group *model.AlertRecipientGroup) error {
	return database.DB.Save(group).Error
}

func (r *AlertRecipientRepo) DeleteGroup(id uint) error {
	return database.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("group_id = ?", id).Delete(&model.AlertGroupRecipient{}).Error; err != nil {
			return err
		}
		return tx.Delete(&model.AlertRecipientGroup{}, id).Error
	})
}

// Group-Recipient relations
func (r *AlertRecipientRepo) GetGroupRecipients(groupID uint) ([]model.AlertRecipient, error) {
	var recipients []model.AlertRecipient
	err := database.DB.Raw(`
		SELECT r.* FROM alert_recipients r
		INNER JOIN alert_group_recipients gr ON r.id = gr.recipient_id
		WHERE gr.group_id = ?
	`, groupID).Scan(&recipients).Error
	return recipients, err
}

func (r *AlertRecipientRepo) SetGroupRecipients(groupID uint, recipientIDs []uint) error {
	return database.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("group_id = ?", groupID).Delete(&model.AlertGroupRecipient{}).Error; err != nil {
			return err
		}
		for _, rid := range recipientIDs {
			if err := tx.Create(&model.AlertGroupRecipient{GroupID: groupID, RecipientID: rid}).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

func (r *AlertRecipientRepo) GetRecipientsByGroupIDs(groupIDs []uint) ([]model.AlertRecipient, error) {
	var recipients []model.AlertRecipient
	if len(groupIDs) == 0 {
		return recipients, nil
	}
	err := database.DB.Raw(`
		SELECT DISTINCT r.* FROM alert_recipients r
		INNER JOIN alert_group_recipients gr ON r.id = gr.recipient_id
		INNER JOIN alert_recipient_groups g ON gr.group_id = g.id
		WHERE gr.group_id IN ? AND r.enabled = ? AND g.enabled = ?
	`, groupIDs, true, true).Scan(&recipients).Error
	return recipients, err
}
