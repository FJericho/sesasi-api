package repository

import (
	"context"
	"strings"

	"github.com/FJericho/sesasi-api/internal/entity"
	"gorm.io/gorm"
)

type AuthRepository interface {
	CreateAccount(ctx context.Context, a *entity.Account) (*entity.Account, error)
	FindUserByEmail(ctx context.Context, email string) (*entity.Account, error)
	FindUserById(ctx context.Context, id string) (*entity.Account, error)
	CheckEmailIfExist(ctx context.Context, email string) (bool, error)

	GetAccounts(ctx context.Context, size, offset int, search, order string, role []string, verified *bool) ([]*entity.Account, int64, error)
	UpdateUserRole(ctx context.Context, userID string, newRole string) error
	UpdatePassword(ctx context.Context, userID, newHashedPassword string) error
	UpdateUserPassword(ctx context.Context, userID, newPassword string) error
	UpdateUserVerify(ctx context.Context, userId string, newVerifiedStatus bool) error
}

type AuthRepositoryImpl struct {
	DB *gorm.DB
}

func NewAuthRepository(db *gorm.DB) AuthRepository {
	return &AuthRepositoryImpl{
		DB: db,
	}
}

func (r *AuthRepositoryImpl) CreateAccount(ctx context.Context, a *entity.Account) (*entity.Account, error) {
	if err := r.DB.WithContext(ctx).Create(a).Error; err != nil {
		return nil, err
	}
	return a, nil
}

func (r *AuthRepositoryImpl) CheckEmailIfExist(ctx context.Context, email string) (bool, error) {
	var account entity.Account

	result := r.DB.WithContext(ctx).Where("email = ?", email).First(&account)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return false, nil
		}
		return false, result.Error
	}
	return true, nil
}

func (r *AuthRepositoryImpl) FindUserByEmail(ctx context.Context, email string) (*entity.Account, error) {
	var account entity.Account

	err := r.DB.WithContext(ctx).Where("email = ?", email).First(&account).Error

	if err != nil {
		return nil, err
	}

	return &account, nil
}

func (r *AuthRepositoryImpl) FindUserById(ctx context.Context, id string) (*entity.Account, error) {
	var account entity.Account

	err := r.DB.WithContext(ctx).Where("id = ?", id).First(&account).Error

	if err != nil {
		return nil, err
	}

	return &account, nil
}

func (r *AuthRepositoryImpl) GetAccounts(ctx context.Context, size int, offset int, search, order string, role []string, verified *bool) ([]*entity.Account, int64, error) {
	var accounts []*entity.Account
	var total int64

	query := r.DB.WithContext(ctx).Model(&entity.Account{}).Where("role IN ?", role).Preload("Permissions")

	if search != "" {
		query = query.Where("LOWER(name) LIKE ?", "%"+strings.ToLower(search)+"%")
	}

	if verified != nil {
		query = query.Where("verified = ?", *verified)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if strings.ToLower(order) == "asc" {
		query = query.Order("created_at ASC")
	} else {
		query = query.Order("created_at DESC")
	}

	if err := query.Limit(size).Offset(offset).Find(&accounts).Error; err != nil {
		return nil, 0, err
	}

	return accounts, total, nil
}

func (r *AuthRepositoryImpl) UpdateUserRole(ctx context.Context, userID string, newRole string) error {
	err := r.DB.WithContext(ctx).Model(&entity.Account{}).Where("id = ?", userID).Updates(map[string]any{
		"role":     newRole,
		"verified": true,
	}).Error

	if err != nil {
		return err
	}

	return nil
}

func (r *AuthRepositoryImpl) UpdatePassword(ctx context.Context, userID, newHashedPassword string) error {
	err := r.DB.WithContext(ctx).Model(&entity.Account{}).Where("id = ?", userID).Update("password", newHashedPassword).Error

	if err != nil {
		return err
	}

	return nil
}

func (r *AuthRepositoryImpl) UpdateUserPassword(ctx context.Context, userID, newPassword string) error {
	return r.DB.WithContext(ctx).Model(&entity.Account{}).Where("id = ?", userID).Update("password", newPassword).Error
}

func (r *AuthRepositoryImpl) UpdateUserVerify(ctx context.Context, userId string, newVerifiedStatus bool) error {
	err := r.DB.WithContext(ctx).Model(&entity.Account{}).Where("id = ?", userId).Update("verified", newVerifiedStatus).Error
	if err != nil {
		return err
	}
	return nil
}
