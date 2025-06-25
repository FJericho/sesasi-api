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

type PermissionController interface {
	GetAllPermissionRequests(ctx *fiber.Ctx) error
	ApprovePermission(ctx *fiber.Ctx) error
	RejectPermission(ctx *fiber.Ctx) error
	RevisionPermission(ctx *fiber.Ctx) error

	CreatePermission(ctx *fiber.Ctx) error
	GetUserPermissions(ctx *fiber.Ctx) error
	GetPermissionByID(ctx *fiber.Ctx) error
	UpdatePermission(ctx *fiber.Ctx) error
	CancelPermission(ctx *fiber.Ctx) error
	DeletePermission(ctx *fiber.Ctx) error
}

type PermissionControllerImpl struct {
	Log               *logrus.Logger
	PermissionService service.PermissionService
	Authentication    middleware.Authentication
}

func NewPermissionController(log *logrus.Logger, permissionService service.PermissionService, authentication middleware.Authentication) PermissionController {
	return &PermissionControllerImpl{
		Log:               log,
		PermissionService: permissionService,
		Authentication:    authentication,
	}
}

func (c *PermissionControllerImpl) GetAllPermissionRequests(ctx *fiber.Ctx) error {
	page, err := strconv.Atoi(ctx.Query("page", "1"))
	if err != nil || page < 1 {
		page = 1
	}

	size, err := strconv.Atoi(ctx.Query("size", "10"))
	if err != nil || size < 1 {
		size = 10
	}

	order := strings.ToLower(ctx.Query("order", "desc"))
	if order != "asc" && order != "desc" {
		order = "desc"
	}

	user := c.Authentication.GetCurrentUserAccount(ctx)
	role := user.Role

	var status *string
	if role == entity.VERIFIER {
		statusStr := ctx.Query("status")
		if statusStr != "" {
			status = &statusStr
		}
	} else if role != entity.ADMIN {
		return fiber.NewError(fiber.StatusForbidden, "Access denied")
	}

	res, meta, err := c.PermissionService.GetAllPermissionRequests(ctx.UserContext(), page, size, status, order)
	if err != nil {
		c.Log.Warnf("Failed to retrieve permission requests: %+v", err)
		return err
	}

	return ctx.Status(fiber.StatusOK).JSON(dto.WebResponse[[]*dto.PermissionResponse]{
		Message: "Success retrieve permission requests",
		Data:    res,
		Paging:  meta,
	})
}

func (c *PermissionControllerImpl) ApprovePermission(ctx *fiber.Ctx) error {
	return c.changePermissionStatus(ctx, entity.StatusApproved)
}

func (c *PermissionControllerImpl) RejectPermission(ctx *fiber.Ctx) error {
	return c.changePermissionStatus(ctx, entity.StatusRejected)
}

func (c *PermissionControllerImpl) RevisionPermission(ctx *fiber.Ctx) error {
	return c.changePermissionStatus(ctx, entity.StatusRevised)
}

func (c *PermissionControllerImpl) changePermissionStatus(ctx *fiber.Ctx, status string) error {
	var payload dto.ApprovalRequest
	id := ctx.Params("id")

	if err := ctx.BodyParser(&payload); err != nil {
		c.Log.Warnf("Failed to parse request body: %+v", err)
		return fiber.NewError(fiber.StatusBadRequest, "Invalid input.")
	}

	err := c.PermissionService.ChangePermissionStatus(ctx.UserContext(), id, status, &payload)
	if err != nil {
		return err
	}

	message := "Permission " + status + " successfully"
	return ctx.Status(fiber.StatusOK).JSON(dto.WebResponse[any]{
		Message: message,
	})
}

func (c *PermissionControllerImpl) CreatePermission(ctx *fiber.Ctx) error {
	var payload dto.CreatePermissionRequest
	if err := ctx.BodyParser(&payload); err != nil {
		c.Log.Warnf("Invalid create request: %+v", err)
		return fiber.NewError(fiber.StatusBadRequest, "Invalid request")
	}

	user := c.Authentication.GetCurrentUserAccount(ctx)
	res, err := c.PermissionService.CreatePermission(ctx.UserContext(), user.ID, &payload)
	if err != nil {
		c.Log.Warnf("Failed to create permission: %+v", err)
		return err
	}

	return ctx.Status(fiber.StatusCreated).JSON(dto.WebResponse[*dto.PermissionResponse]{
		Message: "Permission created successfully",
		Data:    res,
	})
}

func (c *PermissionControllerImpl) GetUserPermissions(ctx *fiber.Ctx) error {
	user := c.Authentication.GetCurrentUserAccount(ctx)

	res, err := c.PermissionService.GetUserPermissions(ctx.UserContext(), user.ID)
	if err != nil {
		c.Log.Warnf("Failed to get user permissions: %+v", err)
		return err
	}

	return ctx.Status(fiber.StatusOK).JSON(dto.WebResponse[[]*dto.PermissionResponse]{
		Message: "User permissions retrieved",
		Data:    res,
	})
}

func (c *PermissionControllerImpl) GetPermissionByID(ctx *fiber.Ctx) error {
	id := ctx.Params("id")

	res, err := c.PermissionService.GetPermissionByID(ctx.UserContext(), id)
	if err != nil {
		c.Log.Warnf("Failed to get permission by ID: %+v", err)
		return err
	}

	return ctx.Status(fiber.StatusOK).JSON(dto.WebResponse[*dto.PermissionResponse]{
		Message: "Permission retrieved",
		Data:    res,
	})
}

func (c *PermissionControllerImpl) UpdatePermission(ctx *fiber.Ctx) error {
	id := ctx.Params("id")

	var payload dto.CreatePermissionRequest
	if err := ctx.BodyParser(&payload); err != nil {
		c.Log.Warnf("Invalid update request: %+v", err)
		return fiber.NewError(fiber.StatusBadRequest, "Invalid request")
	}

	user := c.Authentication.GetCurrentUserAccount(ctx)

	err := c.PermissionService.UpdatePermission(ctx.UserContext(), user.ID, id, &payload)
	if err != nil {
		c.Log.Warnf("Failed to update permission: %+v", err)
		return err
	}

	return ctx.Status(fiber.StatusOK).JSON(dto.WebResponse[any]{
		Message: "Permission updated successfully",
	})
}

func (c *PermissionControllerImpl) CancelPermission(ctx *fiber.Ctx) error {
	id := ctx.Params("id")
	user := c.Authentication.GetCurrentUserAccount(ctx)

	err := c.PermissionService.CancelPermission(ctx.UserContext(), user.ID, id)
	if err != nil {
		c.Log.Warnf("Failed to cancel permission: %+v", err)
		return err
	}

	return ctx.Status(fiber.StatusOK).JSON(dto.WebResponse[any]{
		Message: "Permission cancelled successfully",
	})
}

func (c *PermissionControllerImpl) DeletePermission(ctx *fiber.Ctx) error {
	id := ctx.Params("id")
	user := c.Authentication.GetCurrentUserAccount(ctx)

	err := c.PermissionService.DeletePermission(ctx.UserContext(), user.ID, id)
	if err != nil {
		c.Log.Warnf("Failed to delete permission: %+v", err)
		return err
	}

	return ctx.Status(fiber.StatusOK).JSON(dto.WebResponse[any]{
		Message: "Permission deleted successfully",
	})
}
