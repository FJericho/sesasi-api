package service

import (
	"context"

	"github.com/FJericho/sesasi-api/internal/dto"
	"github.com/FJericho/sesasi-api/internal/entity"
	"github.com/FJericho/sesasi-api/internal/helper"
	"github.com/FJericho/sesasi-api/internal/middleware"
	"github.com/FJericho/sesasi-api/internal/repository"
	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

type AuthService interface {
	Register(ctx context.Context, payload *dto.RegisterRequest) (*dto.AccountResponse, error)
	Login(ctx context.Context, payload *dto.LoginRequest) (*dto.LoginResponse, error)
	ResetUserPassword(ctx context.Context, userId string) error
	UpdatePassword(ctx context.Context, userID string, payload *dto.UpdatePasswordRequest) error

	GetAllUsers(ctx context.Context, page, size int, search, order string, roles []string, verified *bool) ([]*dto.AccountResponse, *dto.PageMetadata, error)
	RegisterVerificator(ctx context.Context, payload *dto.RegisterRequest) (*dto.AccountResponse, error)
	UpdateUserToVerificator(ctx context.Context, userID string) error
	UpdateVerifyUser(ctx context.Context, userId string) error
	FindUserById(ctx context.Context, userId string) (*dto.AccountResponse, error)
}

type AuthServiceImpl struct {
	Log            *logrus.Logger
	Viper          *viper.Viper
	Validate       *validator.Validate
	AuthRepository repository.AuthRepository
	Authentication middleware.Authentication
}

func NewAuthService(log *logrus.Logger, viper *viper.Viper, validate *validator.Validate, authRepository repository.AuthRepository, authentication middleware.Authentication) AuthService {
	return &AuthServiceImpl{
		Log:            log,
		Viper:          viper,
		Validate:       validate,
		AuthRepository: authRepository,
		Authentication: authentication,
	}
}

func (s *AuthServiceImpl) Register(ctx context.Context, payload *dto.RegisterRequest) (*dto.AccountResponse, error) {
	err := s.Validate.Struct(payload)
	if err != nil {
		s.Log.Warnf("Invalid validation: %+v", err)
		errorResponse := helper.GenerateValidationErrors(err)
		return nil, fiber.NewError(fiber.StatusBadRequest, errorResponse)
	}

	exist, err := s.AuthRepository.CheckEmailIfExist(ctx, payload.Email)
	if exist {
		s.Log.Warn("Email already in use")
		return nil, fiber.NewError(fiber.StatusUnprocessableEntity, "Email already in use")
	}

	hashPassword, err := helper.HashPassword(payload.Password)
	if err != nil {
		s.Log.Warnf("Failed to generate bcrype hash : %+v", err)
		return nil, fiber.NewError(fiber.StatusInternalServerError, "Failed to create hash password")
	}

	user, err := s.AuthRepository.CreateAccount(ctx, &entity.Account{
		Name:     payload.Name,
		Email:    payload.Email,
		Password: hashPassword,
		Role:     entity.USER,
	})

	if err != nil {
		s.Log.Warnf("Failed create user account to database : %+v", err)
		return nil, fiber.NewError(fiber.StatusInternalServerError, "Failed to register a user account")
	}

	return &dto.AccountResponse{
		ID:        user.ID,
		Name:      user.Name,
		Email:     user.Email,
		Role:      user.Role,
		CreatedAt: &user.CreatedAt,
	}, nil
}

func (s *AuthServiceImpl) Login(ctx context.Context, payload *dto.LoginRequest) (*dto.LoginResponse, error) {
	err := s.Validate.Struct(payload)
	if err != nil {
		s.Log.Warnf("Invalid validation: %+v", err)
		errorResponse := helper.GenerateValidationErrors(err)
		return nil, fiber.NewError(fiber.StatusBadRequest, errorResponse)
	}

	exist, err := s.AuthRepository.CheckEmailIfExist(ctx, payload.Email)
	if !exist {
		s.Log.Warn("Email not found")
		return nil, fiber.NewError(fiber.StatusUnauthorized, "Incorrect email or password")
	}

	user, _ := s.AuthRepository.FindUserByEmail(ctx, payload.Email)

	if err := helper.ComparePassword(user.Password, payload.Password); err != nil {
		s.Log.Warnf("Failed to compare user password with bcrype hash : %+v", err)
		return nil, fiber.NewError(fiber.StatusUnauthorized, "Incorrect email or password")
	}

	token, err := s.Authentication.GenerateToken(user.ID, user.Email, user.Role, user.Name)
	if err != nil {
		s.Log.Warnf("Failed generated token : %+v", err)
		return nil, fiber.NewError(fiber.StatusInternalServerError, "Failed generated token")
	}

	return &dto.LoginResponse{
		Token: token,
	}, nil
}

func (s *AuthServiceImpl) ResetUserPassword(ctx context.Context, userId string) error {
	hashed, err := helper.HashPassword(s.Viper.GetString("pass.default"))
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to hash password")
	}

	err = s.AuthRepository.UpdatePassword(ctx, userId, hashed)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to reset password")
	}

	return nil
}

func (s *AuthServiceImpl) UpdatePassword(ctx context.Context, userID string, payload *dto.UpdatePasswordRequest) error {
	err := s.Validate.Struct(payload)
	if err != nil {
		errorResponse := helper.GenerateValidationErrors(err)
		return fiber.NewError(fiber.StatusBadRequest, errorResponse)
	}

	account, err := s.AuthRepository.FindUserById(ctx, userID)
	if err != nil {
		return fiber.NewError(fiber.StatusNotFound, "User not found")
	}

	if err := helper.ComparePassword(account.Password, payload.OldPassword); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Old password is incorrect")
	}

	hashed, err := helper.HashPassword(payload.NewPassword)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to hash password")
	}

	if err := s.AuthRepository.UpdateUserPassword(ctx, userID, hashed); err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to update password")
	}

	return nil
}

func (s *AuthServiceImpl) GetAllUsers(ctx context.Context, page, size int, search, order string, roles []string, verified *bool) ([]*dto.AccountResponse, *dto.PageMetadata, error) {
	offset := (page - 1) * size

	users, total, err := s.AuthRepository.GetAccounts(ctx, size, offset, search, order, roles, verified)
	if err != nil {
		s.Log.Warnf("Failed to retrieve users: %+v", err)
		return nil, nil, fiber.NewError(fiber.StatusInternalServerError, "Failed to retrieve users")
	}

	var responses []*dto.AccountResponse
	for _, a := range users {
		var permissionResponses []*dto.PermissionResponse
		for _, p := range a.Permissions {
			permissionResponses = append(permissionResponses, &dto.PermissionResponse{
				ID:        p.Id,
				Title:     p.Title,
				Reason:    p.Reason,
				StartDate: p.StartDate,
				EndDate:   p.EndDate,
				Comment:   p.Comment,
				Status:    p.Status,
				CreatedAt: p.CreatedAt,
				UpdatedAt: p.UpdatedAt,
			})
		}

		responses = append(responses, &dto.AccountResponse{
			ID:          a.ID,
			Name:        a.Name,
			Email:       a.Email,
			Verified:    a.Verified,
			Role:        a.Role,
			CreatedAt:   &a.CreatedAt,
			Permissions: permissionResponses,
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

func (s *AuthServiceImpl) RegisterVerificator(ctx context.Context, payload *dto.RegisterRequest) (*dto.AccountResponse, error) {
	if err := s.Validate.Struct(payload); err != nil {
		s.Log.Warnf("Validation error: %+v", err)
		return nil, fiber.NewError(fiber.StatusBadRequest, helper.GenerateValidationErrors(err))
	}

	exists, err := s.AuthRepository.CheckEmailIfExist(ctx, payload.Email)
	if exists {
		return nil, fiber.NewError(fiber.StatusConflict, "Email already exists")
	}

	hashedPassword, err := helper.HashPassword(payload.Password)
	if err != nil {
		return nil, fiber.NewError(fiber.StatusInternalServerError, "Failed to hash password")
	}

	user, err := s.AuthRepository.CreateAccount(ctx, &entity.Account{
		Name:     payload.Name,
		Email:    payload.Email,
		Password: hashedPassword,
		Role:     entity.VERIFIER,
		Verified: true,
	})

	if err != nil {
		return nil, fiber.NewError(fiber.StatusInternalServerError, "Failed to register user")
	}

	return &dto.AccountResponse{
		ID:        user.ID,
		Name:      user.Name,
		Email:     user.Email,
		Role:      user.Role,
		Verified:  user.Verified,
		CreatedAt: &user.CreatedAt,
	}, nil
}

func (s *AuthServiceImpl) UpdateUserToVerificator(ctx context.Context, userID string) error {
	err := s.AuthRepository.UpdateUserRole(ctx, userID, entity.VERIFIER)
	if err != nil {
		s.Log.Warnf("Failed to update role: %+v", err)
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to update user role")
	}

	return nil
}

func (s *AuthServiceImpl) UpdateVerifyUser(ctx context.Context, userId string) error {
	user, err := s.AuthRepository.FindUserById(ctx, userId)
	if err != nil {
		return fiber.NewError(fiber.StatusNotFound, "User not found")
	}

	newVerifiedStatus := !user.Verified

	err = s.AuthRepository.UpdateUserVerify(ctx, userId, newVerifiedStatus)
	if err != nil {
		s.Log.Warnf("Failed to update verified status: %+v", err)
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to update verified status")
	}

	return nil
}

func (s *AuthServiceImpl) FindUserById(ctx context.Context, userId string) (*dto.AccountResponse, error) {
	user, err := s.AuthRepository.FindUserById(ctx, userId)
	if err != nil {
		return nil, fiber.NewError(fiber.StatusNotFound, "User not found")
	}

	return &dto.AccountResponse{
		Name:      user.Name,
		Email:     user.Email,
		Role:      user.Role,
		Verified:  user.Verified,
		CreatedAt: &user.CreatedAt,
	}, nil
}
