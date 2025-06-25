package repository

import (
	"context"
	"strings"

	"github.com/FJericho/sesasi-api/internal/entity"
	"gorm.io/gorm"
)

type PermissionRepository interface {
	FindAllPermissions(ctx context.Context, size, offset int, status *string, order string) ([]*entity.Permission, int64, error)
	FindByID(ctx context.Context, id string) (*entity.Permission, error)
	UpdatePermissionStatus(ctx context.Context, id string, status, comment string) error

	Create(ctx context.Context, p *entity.Permission) (*entity.Permission, error)
	FindByAccountID(ctx context.Context, userID string) ([]*entity.Permission, error)
	Update(ctx context.Context, id string, e *entity.Permission) error
	UpdateStatus(ctx context.Context, id, status, comment string) error
	Delete(ctx context.Context, id string) error
}

type PermissionRepositoryImpl struct {
	DB *gorm.DB
}

func NewPermissionRepository(db *gorm.DB) PermissionRepository {
	return &PermissionRepositoryImpl{
		DB: db,
	}
}

func (r *PermissionRepositoryImpl) FindAllPermissions(ctx context.Context, size, offset int, status *string, order string) ([]*entity.Permission, int64, error) {
	var permissions []*entity.Permission
	var total int64

	query := r.DB.WithContext(ctx).Model(&entity.Permission{}).Preload("Account")

	if status != nil && *status != "" {
		query = query.Where("LOWER(status) = ?", strings.ToLower(*status))
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if strings.ToLower(order) == "asc" {
		query = query.Order("created_at ASC")
	} else {
		query = query.Order("created_at DESC")
	}

	if err := query.Limit(size).Offset(offset).Find(&permissions).Error; err != nil {
		return nil, 0, err
	}

	return permissions, total, nil
}

func (r *PermissionRepositoryImpl) FindByID(ctx context.Context, id string) (*entity.Permission, error) {
	var permission entity.Permission

	err := r.DB.WithContext(ctx).Preload("Account").First(&permission, "id = ?", id).Error

	if err != nil {
		return nil, err
	}

	return &permission, nil
}

func (r *PermissionRepositoryImpl) UpdatePermissionStatus(ctx context.Context, id string, status string, comment string) error {
	return r.DB.WithContext(ctx).Model(&entity.Permission{}).Where("id = ?", id).
		Updates(map[string]any{
			"status":  status,
			"comment": comment,
		}).Error
}

func (r *PermissionRepositoryImpl) Create(ctx context.Context, p *entity.Permission) (*entity.Permission, error) {
	if err := r.DB.WithContext(ctx).Create(p).Error; err != nil {
		return nil, err
	}

	if err := r.DB.WithContext(ctx).Preload("Account").First(p, "id = ?", p.Id).Error; err != nil {
		return nil, err
	}
	return p, nil
}

func (r *PermissionRepositoryImpl) FindByAccountID(ctx context.Context, userID string) ([]*entity.Permission, error) {
	var permissions []*entity.Permission
	if err := r.DB.WithContext(ctx).Where("account_id = ?", userID).Find(&permissions).Error; err != nil {
		return nil, err
	}
	return permissions, nil
}

func (r *PermissionRepositoryImpl) Update(ctx context.Context, id string, e *entity.Permission) error {
	return r.DB.WithContext(ctx).Model(&entity.Permission{}).Where("id = ?", id).Updates(e).Error
}

func (r *PermissionRepositoryImpl) UpdateStatus(ctx context.Context, id, status, comment string) error {
	return r.DB.WithContext(ctx).Model(&entity.Permission{}).Where("id = ?", id).Updates(entity.Permission{
		Status:  status,
		Comment: comment,
	}).Error
}

func (r *PermissionRepositoryImpl) Delete(ctx context.Context, id string) error {
	return r.DB.WithContext(ctx).Delete(&entity.Permission{}, "id = ?", id).Error
}
