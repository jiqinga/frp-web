/*
 * @Author              : 寂情啊
 * @Date                : 2025-12-22 15:45:57
 * @LastEditors         : 寂情啊
 * @LastEditTime        : 2025-12-22 15:46:08
 * @FilePath            : frp-web-testbackendinternalrepositorydns_record_repo.go
 * @Description         : 说明
 * 倾尽绿蚁花尽开，问潭底剑仙安在哉
 */
package repository

import (
	"frp-web-panel/internal/model"
	"frp-web-panel/pkg/database"
)

type DNSRecordRepository struct{}

func NewDNSRecordRepository() *DNSRecordRepository {
	return &DNSRecordRepository{}
}

func (r *DNSRecordRepository) Create(record *model.DNSRecord) error {
	return database.DB.Create(record).Error
}

func (r *DNSRecordRepository) FindByID(id uint) (*model.DNSRecord, error) {
	var record model.DNSRecord
	err := database.DB.First(&record, id).Error
	return &record, err
}

func (r *DNSRecordRepository) FindByProxyID(proxyID uint) (*model.DNSRecord, error) {
	var record model.DNSRecord
	err := database.DB.Where("proxy_id = ?", proxyID).First(&record).Error
	return &record, err
}

func (r *DNSRecordRepository) FindAll() ([]model.DNSRecord, error) {
	var records []model.DNSRecord
	err := database.DB.Find(&records).Error
	return records, err
}

func (r *DNSRecordRepository) Update(record *model.DNSRecord) error {
	return database.DB.Save(record).Error
}

func (r *DNSRecordRepository) Delete(id uint) error {
	return database.DB.Delete(&model.DNSRecord{}, id).Error
}

func (r *DNSRecordRepository) DeleteByProxyID(proxyID uint) error {
	return database.DB.Where("proxy_id = ?", proxyID).Delete(&model.DNSRecord{}).Error
}
