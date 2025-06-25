package controller

import (
	"strconv"
	"strings"

	"github.com/FJericho/sesasi-api/internal/dto"
	"github.com/FJericho/sesasi-api/internal/entity"
	"github.com/FJericho/sesasi-api/internal/middleware"
	"github.com/FJericho/sesasi-api/internal/service"
	"github.com/gofiber/fiber/v2"
	"github.com/sirupsen/logrus"
)

type AuthController interface {
	Register(ctx *fiber.Ctx) error
	Login(ctx *fiber.Ctx) error

	ResetUserPassword(ctx *fiber.Ctx) error
	UpdatePassword(ctx *fiber.Ctx) error

	GetAllUsers(ctx *fiber.Ctx) error
	RegisterVerificator(ctx *fiber.Ctx) error
	UpdateUserToVerificator(ctx *fiber.Ctx) error
	VerifyUser(ctx *fiber.Ctx) error
	GetAccountById(ctx *fiber.Ctx) error
}

type AuthControllerImpl struct {
	Log            *logrus.Logger
	AuthService    service.AuthService
	Authentication middleware.Authentication
}

func NewAuthController(log *logrus.Logger, authService service.AuthService, authentication middleware.Authentication) AuthController {
	return &AuthControllerImpl{
		Log:            log,
		AuthService:    authService,
		Authentication: authentication,
	}
}

func (c *AuthControllerImpl) Register(ctx *fiber.Ctx) error {
	var payload dto.RegisterRequest

	if err := ctx.BodyParser(&payload); err != nil {
		c.Log.Warnf("Failed to parse request body : %+v", err)
		return fiber.NewError(fiber.StatusBadRequest, "Invalid request body.")
	}

	response, err := c.AuthService.Register(ctx.UserContext(), &payload)
	if err != nil {
		c.Log.Warnf("Failed to register user : %+v", err)
		return err
	}

	return ctx.Status(fiber.StatusCreated).JSON(dto.WebResponse[*dto.AccountResponse]{
		Message: "Register Successfully",
		Data:    response,
	})
}

func (c *AuthControllerImpl) Login(ctx *fiber.Ctx) error {
	var payload dto.LoginRequest

	if err := ctx.BodyParser(&payload); err != nil {
		c.Log.Warnf("Failed to parse request body : %+v", err)
		return fiber.NewError(fiber.StatusBadRequest, "Invalid request body.")
	}

	response, err := c.AuthService.Login(ctx.UserContext(), &payload)

	if err != nil {
		c.Log.Warnf("Failed to login user : %+v", err)
		return err
	}

	return ctx.Status(fiber.StatusOK).JSON(dto.WebResponse[*dto.LoginResponse]{
		Data:    response,
		Message: "Login Successfully",
	})
}

func (c *AuthControllerImpl) ResetUserPassword(ctx *fiber.Ctx) error {
	userID := ctx.Params("id")
	if userID == "" {
		return fiber.NewError(fiber.StatusBadRequest, "User ID is required")
	}

	err := c.AuthService.ResetUserPassword(ctx.UserContext(), userID)
	if err != nil {
		return err
	}

	return ctx.Status(fiber.StatusOK).JSON(dto.WebResponse[any]{
		Message: "Password reset successfully",
	})
}

func (c *AuthControllerImpl) UpdatePassword(ctx *fiber.Ctx) error {
	var payload dto.UpdatePasswordRequest
	if err := ctx.BodyParser(&payload); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid input")
	}

	user := c.Authentication.GetCurrentUserAccount(ctx)
	err := c.AuthService.UpdatePassword(ctx.UserContext(), user.ID, &payload)
	if err != nil {
		return err
	}

	return ctx.Status(fiber.StatusOK).JSON(dto.WebResponse[any]{
		Message: "Password updated successfully",
	})
}

func (c *AuthControllerImpl) GetAllUsers(ctx *fiber.Ctx) error {
	page, err := strconv.Atoi(ctx.Query("page", "1"))
	if err != nil || page < 1 {
		page = 1
	}

	size, err := strconv.Atoi(ctx.Query("size", "10"))
	if err != nil || size < 1 {
		size = 10
	}

	search := ctx.Query("search", "")
	order := strings.ToLower(ctx.Query("order", "desc"))
	if order != "asc" && order != "desc" {
		order = "asc"
	}

	user := c.Authentication.GetCurrentUserAccount(ctx)
	role := user.Role

	var (
		res  []*dto.AccountResponse
		meta *dto.PageMetadata
	)

	switch role {
	case entity.ADMIN:
		roles := []string{entity.USER, entity.VERIFIER}
		res, meta, err = c.AuthService.GetAllUsers(ctx.UserContext(), page, size, search, order, roles, nil)
	case entity.VERIFIER:
		var verified *bool
		verifiedStr := ctx.Query("verified")
		switch verifiedStr {
		case "true":
			val := true
			verified = &val
		case "false":
			val := false
			verified = &val
		}

		roles := []string{entity.USER}
		res, meta, err = c.AuthService.GetAllUsers(ctx.UserContext(), page, size, search, order, roles, verified)
	default:
		return fiber.NewError(fiber.StatusForbidden, "Access denied")
	}

	if err != nil {
		c.Log.Warnf("Failed to get all users: %+v", err)
		return err
	}

	return ctx.Status(fiber.StatusOK).JSON(dto.WebResponse[[]*dto.AccountResponse]{
		Message: "Success get all users",
		Data:    res,
		Paging:  meta,
	})
}

func (c *AuthControllerImpl) RegisterVerificator(ctx *fiber.Ctx) error {
	var payload dto.RegisterRequest
	if err := ctx.BodyParser(&payload); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid request body")
	}

	user, err := c.AuthService.RegisterVerificator(ctx.UserContext(), &payload)
	if err != nil {
		return err
	}

	return ctx.Status(fiber.StatusCreated).JSON(dto.WebResponse[*dto.AccountResponse]{
		Message: "Verificator registered successfully",
		Data:    user,
	})
}

func (c *AuthControllerImpl) UpdateUserToVerificator(ctx *fiber.Ctx) error {
	id := ctx.Params("id")
	if id == "" {
		return fiber.NewError(fiber.StatusBadRequest, "Id is required")
	}

	err := c.AuthService.UpdateUserToVerificator(ctx.UserContext(), id)
	if err != nil {
		return err
	}

	return ctx.Status(fiber.StatusOK).JSON(dto.WebResponse[any]{
		Message: "Updated role successfully",
	})
}

func (c *AuthControllerImpl) VerifyUser(ctx *fiber.Ctx) error {
	id := ctx.Params("id")
	if id == "" {
		return fiber.NewError(fiber.StatusBadRequest, "Id is required")
	}

	err := c.AuthService.UpdateVerifyUser(ctx.UserContext(), id)
	if err != nil {
		return err
	}

	return ctx.Status(fiber.StatusOK).JSON(dto.WebResponse[any]{
		Message: "User verify updated successfully",
	})
}

func (c *AuthControllerImpl) GetAccountById(ctx *fiber.Ctx) error {
	id := ctx.Params("id")
	if id == "" {
		return fiber.NewError(fiber.StatusBadRequest, "Id is required")
	}

	account, err := c.AuthService.FindUserById(ctx.UserContext(), id)
	if err != nil {
		return err
	}

	return ctx.Status(fiber.StatusOK).JSON(dto.WebResponse[*dto.AccountResponse]{
		Message: "User detail fetched successfully",
		Data:    account,
	})
}
