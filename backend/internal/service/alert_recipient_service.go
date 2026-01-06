/*
 * @Author              : 寂情啊
 * @Date                : 2025-12-12 14:20:43
 * @LastEditors         : 寂情啊
 * @LastEditTime        : 2025-12-12 14:20:57
 * @FilePath            : frp-web-testbackendinternalservicealert_recipient_service.go
 * @Description         : 说明
 * 倾尽绿蚁花尽开，问潭底剑仙安在哉
 */
package service

import (
	"frp-web-panel/internal/model"
	"frp-web-panel/internal/repository"
)

type AlertRecipientService struct {
	repo *repository.AlertRecipientRepo
}

func NewAlertRecipientService() *AlertRecipientService {
	return &AlertRecipientService{repo: repository.NewAlertRecipientRepo()}
}

// Recipient methods
func (s *AlertRecipientService) CreateRecipient(r *model.AlertRecipient) error {
	return s.repo.CreateRecipient(r)
}

func (s *AlertRecipientService) GetAllRecipients() ([]model.AlertRecipient, error) {
	return s.repo.GetAllRecipients()
}

func (s *AlertRecipientService) GetRecipientByID(id uint) (*model.AlertRecipient, error) {
	return s.repo.GetRecipientByID(id)
}

func (s *AlertRecipientService) UpdateRecipient(r *model.AlertRecipient) error {
	return s.repo.UpdateRecipient(r)
}

func (s *AlertRecipientService) DeleteRecipient(id uint) error {
	return s.repo.DeleteRecipient(id)
}

// Group methods
func (s *AlertRecipientService) CreateGroup(g *model.AlertRecipientGroup) error {
	return s.repo.CreateGroup(g)
}

func (s *AlertRecipientService) GetAllGroups() ([]model.AlertRecipientGroup, error) {
	groups, err := s.repo.GetAllGroups()
	if err != nil {
		return nil, err
	}
	for i := range groups {
		groups[i].Recipients, _ = s.repo.GetGroupRecipients(groups[i].ID)
	}
	return groups, nil
}

func (s *AlertRecipientService) GetGroupByID(id uint) (*model.AlertRecipientGroup, error) {
	group, err := s.repo.GetGroupByID(id)
	if err != nil {
		return nil, err
	}
	group.Recipients, _ = s.repo.GetGroupRecipients(id)
	return group, nil
}

func (s *AlertRecipientService) UpdateGroup(g *model.AlertRecipientGroup) error {
	return s.repo.UpdateGroup(g)
}

func (s *AlertRecipientService) DeleteGroup(id uint) error {
	return s.repo.DeleteGroup(id)
}

func (s *AlertRecipientService) SetGroupRecipients(groupID uint, recipientIDs []uint) error {
	return s.repo.SetGroupRecipients(groupID, recipientIDs)
}

// GetEmailsByRecipientAndGroupIDs 根据接收人ID和分组ID获取所有邮箱（去重）
func (s *AlertRecipientService) GetEmailsByRecipientAndGroupIDs(recipientIDs, groupIDs []uint) []string {
	emailMap := make(map[string]bool)

	if len(recipientIDs) > 0 {
		recipients, _ := s.repo.GetRecipientsByIDs(recipientIDs)
		for _, r := range recipients {
			emailMap[r.Email] = true
		}
	}

	if len(groupIDs) > 0 {
		recipients, _ := s.repo.GetRecipientsByGroupIDs(groupIDs)
		for _, r := range recipients {
			emailMap[r.Email] = true
		}
	}

	emails := make([]string, 0, len(emailMap))
	for email := range emailMap {
		emails = append(emails, email)
	}
	return emails
}
