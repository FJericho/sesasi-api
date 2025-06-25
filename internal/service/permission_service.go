package service

import (
	"context"

	"github.com/FJericho/sesasi-api/internal/dto"
	"github.com/FJericho/sesasi-api/internal/entity"
	"github.com/FJericho/sesasi-api/internal/helper"
	"github.com/FJericho/sesasi-api/internal/repository"
	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

type PermissionService interface {
	GetAllPermissionRequests(ctx context.Context, page, size int, status *string, order string) ([]*dto.PermissionResponse, *dto.PageMetadata, error)
	ChangePermissionStatus(ctx context.Context, id string, status string, payload *dto.ApprovalRequest) error

	CreatePermission(ctx context.Context, userID string, payload *dto.CreatePermissionRequest) (*dto.PermissionResponse, error)
	GetUserPermissions(ctx context.Context, userID string) ([]*dto.PermissionResponse, error)
	GetPermissionByID(ctx context.Context, id string) (*dto.PermissionResponse, error)
	UpdatePermission(ctx context.Context, userID, id string, payload *dto.CreatePermissionRequest) error
	CancelPermission(ctx context.Context, userID, id string) error
	DeletePermission(ctx context.Context, userID, id string) error
}

type PermissionServiceImpl struct {
	Log            *logrus.Logger
	Viper          *viper.Viper
	Validate       *validator.Validate
	PermissionRepo repository.PermissionRepository
}

func NewPermissionService(log *logrus.Logger, viper *viper.Viper, validate *validator.Validate, permissionRepo repository.PermissionRepository) PermissionService {
	return &PermissionServiceImpl{
		Log:            log,
		Validate:       validate,
		PermissionRepo: permissionRepo,
	}
}

func (s *PermissionServiceImpl) GetAllPermissionRequests(ctx context.Context, page, size int, status *string, order string) ([]*dto.PermissionResponse, *dto.PageMetadata, error) {
	offset := (page - 1) * size

	permissions, total, err := s.PermissionRepo.FindAllPermissions(ctx, size, offset, status, order)
	if err != nil {
		s.Log.Warnf("Failed to get permissions: %+v", err)
		return nil, nil, fiber.NewError(fiber.StatusInternalServerError, "Failed to get permission requests")
	}

	var responses []*dto.PermissionResponse
	for _, p := range permissions {
		responses = append(responses, &dto.PermissionResponse{
			ID:        p.Id,
			Title:     p.Title,
			Reason:    p.Reason,
			StartDate: p.StartDate,
			EndDate:   p.EndDate,
			Comment:   p.Comment,
			Status:    p.Status,
			CreatedAt: p.CreatedAt,
			UpdatedAt: p.UpdatedAt,
			Account: &dto.AccountResponse{
				Name:  p.Account.Name,
				Email: p.Account.Email,
				Role:  p.Account.Role,
			},
		})
	}

	totalPages := int((total + int64(size) - 1) / int64(size))
	meta := &dto.PageMetadata{
		Page:        page,
		Size:        size,
		TotalItem:   total,
		TotalPage:   int64(totalPages),
		HasNext:     page < totalPages,
		HasPrevious: page > 1,
	}

	return responses, meta, nil
}

func (s *PermissionServiceImpl) ChangePermissionStatus(ctx context.Context, id string, status string, payload *dto.ApprovalRequest) error {
	if err := s.Validate.Struct(payload); err != nil {
		s.Log.Warnf("Invalid validation: %+v", err)
		errorResponse := helper.GenerateValidationErrors(err)
		return fiber.NewError(fiber.StatusBadRequest, errorResponse)
	}

	_, err := s.PermissionRepo.FindByID(ctx, id)
	if err != nil {
		s.Log.Warnf("Permission not found: %+v", err)
		return fiber.NewError(fiber.StatusNotFound, "Permission not found")
	}

	if err := s.PermissionRepo.UpdatePermissionStatus(ctx, id, status, payload.Comment); err != nil {
		s.Log.Warnf("Failed to update permission status: %+v", err)
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to update permission status")
	}

	return nil
}

func (s *PermissionServiceImpl) CreatePermission(ctx context.Context, userID string, payload *dto.CreatePermissionRequest) (*dto.PermissionResponse, error) {
	if err := s.Validate.Struct(payload); err != nil {
		s.Log.Warnf("Invalid validation: %+v", err)
		errorResponse := helper.GenerateValidationErrors(err)
		return nil, fiber.NewError(fiber.StatusBadRequest, errorResponse)
	}

	entity := &entity.Permission{
		Title:     payload.Title,
		Reason:    payload.Reason,
		StartDate: payload.StartDate,
		EndDate:   payload.EndDate,
		AccountID: userID,
	}

	created, err := s.PermissionRepo.Create(ctx, entity)
	if err != nil {
		s.Log.Warnf("Failed to create permission: %+v", err)
		return nil, fiber.NewError(fiber.StatusInternalServerError, "Failed to create permission")
	}

	return dto.ToPermissionResponse(created), nil
}

func (s *PermissionServiceImpl) GetUserPermissions(ctx context.Context, userID string) ([]*dto.PermissionResponse, error) {
	permissions, err := s.PermissionRepo.FindByAccountID(ctx, userID)
	if err != nil {
		s.Log.Warnf("Failed to get user permissions: %+v", err)
		return nil, fiber.NewError(fiber.StatusInternalServerError, "Failed to get permissions")
	}
	return dto.ToPermissionResponseList(permissions), nil
}

func (s *PermissionServiceImpl) GetPermissionByID(ctx context.Context, id string) (*dto.PermissionResponse, error) {
	perm, err := s.PermissionRepo.FindByID(ctx, id)
	if err != nil {
		s.Log.Warnf("Permission not found: %+v", err)
		return nil, fiber.NewError(fiber.StatusNotFound, "Permission not found")
	}
	return dto.ToPermissionResponse(perm), nil
}

func (s *PermissionServiceImpl) UpdatePermission(ctx context.Context, userID, id string, payload *dto.CreatePermissionRequest) error {
	if err := s.Validate.Struct(payload); err != nil {
		s.Log.Warnf("Invalid input: %+v", err)
		errorResponse := helper.GenerateValidationErrors(err)
		return fiber.NewError(fiber.StatusBadRequest, errorResponse)
	}

	perm, err := s.PermissionRepo.FindByID(ctx, id)
	if err != nil {
		s.Log.Warnf("Permission not found: %+v", err)
		return fiber.NewError(fiber.StatusNotFound, "Permission not found")
	}

	if perm.AccountID != userID {
		return fiber.NewError(fiber.StatusForbidden, "You can only update your own permission")
	}

	if perm.Status != entity.StatusPending && perm.Status != entity.StatusRevised {
		return fiber.NewError(fiber.StatusForbidden, "Permission cannot be updated in current status")
	}

	updated := &entity.Permission{
		Title:     payload.Title,
		Reason:    payload.Reason,
		StartDate: payload.StartDate,
		EndDate:   payload.EndDate,
	}

	return s.PermissionRepo.Update(ctx, id, updated)
}

func (s *PermissionServiceImpl) CancelPermission(ctx context.Context, userID, id string) error {
	perm, err := s.PermissionRepo.FindByID(ctx, id)
	if err != nil {
		return fiber.NewError(fiber.StatusNotFound, "Permission not found")
	}

	if perm.AccountID != userID {
		return fiber.NewError(fiber.StatusForbidden, "You can only cancel your own permission")
	}

	if perm.Status != entity.StatusPending && perm.Status != entity.StatusRevised {
		return fiber.NewError(fiber.StatusForbidden, "Permission can`t be cancelled")
	}

	return s.PermissionRepo.UpdateStatus(ctx, id, entity.StatusCancelled, "Cancelled by user")
}

func (s *PermissionServiceImpl) DeletePermission(ctx context.Context, userID, id string) error {
	perm, err := s.PermissionRepo.FindByID(ctx, id)
	if err != nil {
		return fiber.NewError(fiber.StatusNotFound, "Permission not found")
	}

	if perm.AccountID != userID {
		return fiber.NewError(fiber.StatusForbidden, "You can only delete your own permission")
	}

	if perm.Status != entity.StatusPending {
		return fiber.NewError(fiber.StatusForbidden, "Only pending permission can be deleted")
	}

	return s.PermissionRepo.Delete(ctx, id)
}
