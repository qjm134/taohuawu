package database

import (
	"time"

	"gorm.io/gorm"
)

// AuditRepository 审计日志仓储接口
type AuditRepository interface {
	Create(log *AuditLog) error
	GetByTenantID(tenantID string, startDate, endDate time.Time, page, pageSize int) ([]*AuditLog, int64, error)
	GetCountByTenantID(tenantID string, since time.Time) (int64, error)
}

type auditRepository struct {
	db *gorm.DB
}

// NewAuditRepository 创建审计日志仓储
func NewAuditRepository(db *gorm.DB) AuditRepository {
	return &auditRepository{db: db}
}

func (r *auditRepository) Create(log *AuditLog) error {
	return r.db.Create(log).Error
}

func (r *auditRepository) GetByTenantID(tenantID string, startDate, endDate time.Time, page, pageSize int) ([]*AuditLog, int64, error) {
	var logs []*AuditLog
	var total int64

	query := r.db.Model(&AuditLog{}).Where("tenant_id = ?", tenantID)

	if !startDate.IsZero() {
		query = query.Where("created_at >= ?", startDate)
	}
	if !endDate.IsZero() {
		query = query.Where("created_at <= ?", endDate)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	err := query.Order("created_at DESC").
		Limit(pageSize).
		Offset(offset).
		Find(&logs).Error

	return logs, total, err
}

func (r *auditRepository) GetCountByTenantID(tenantID string, since time.Time) (int64, error) {
	var count int64
	query := r.db.Model(&AuditLog{}).Where("tenant_id = ?", tenantID)
	if !since.IsZero() {
		query = query.Where("created_at >= ?", since)
	}
	return count, query.Count(&count).Error
}