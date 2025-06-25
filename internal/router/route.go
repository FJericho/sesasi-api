package router

import (
	"github.com/FJericho/sesasi-api/internal/controller"
	"github.com/FJericho/sesasi-api/internal/middleware"
	"github.com/gofiber/fiber/v2"
)

type RouteConfig struct {
	App                      *fiber.App
	AuthController           controller.AuthController
	PermissionController     controller.PermissionController
	AuthenticationMiddleware middleware.Authentication
	AuthorizationMiddleware  middleware.Authorization
}

func (r *RouteConfig) Setup() {
	r.SetupPublicRoute()
	r.SetupAdminRoute()
	r.SetupVerifierRoute()
	r.SetupUserRoute()
}

func (r *RouteConfig) SetupPublicRoute() {
	r.App.Post("/api/v1/login", r.AuthController.Login)
	r.App.Post("/api/v1/register", r.AuthController.Register)

}

func (r *RouteConfig) SetupAdminRoute() {
	admin := r.App.Group("/api/v1/admin", r.AuthenticationMiddleware.Authorize, r.AuthorizationMiddleware.AuthorizeAdmin)

	admin.Get("/users", r.AuthController.GetAllUsers)
	admin.Get("/user/:id", r.AuthController.GetAccountById)
	admin.Post("/verificator", r.AuthController.RegisterVerificator)
	admin.Patch("/users/:id/verify", r.AuthController.UpdateUserToVerificator)
	admin.Patch("/users/:id/reset-password", r.AuthController.ResetUserPassword)

	admin.Get("/permissions", r.PermissionController.GetAllPermissionRequests)
	admin.Get("/permission/:id", r.PermissionController.GetPermissionByID)
	

}

func (r *RouteConfig) SetupVerifierRoute() {
	verif := r.App.Group("/api/v1/verificator", r.AuthenticationMiddleware.Authorize, r.AuthorizationMiddleware.AuthorizeVerifier)

	verif.Get("/users", r.AuthController.GetAllUsers)
	verif.Patch("/users/:id/verify", r.AuthController.VerifyUser)

	verif.Get("/permissions", r.PermissionController.GetAllPermissionRequests) 

	verif.Patch("/permissions/:id/approve", r.PermissionController.ApprovePermission)
	verif.Patch("/permissions/:id/reject", r.PermissionController.RejectPermission)
	verif.Patch("/permissions/:id/revision", r.PermissionController.RevisionPermission)
	verif.Get("/permissions/:id", r.PermissionController.GetPermissionByID)

	verif.Get("/user/:id", r.AuthController.GetAccountById)
}

func (r *RouteConfig) SetupUserRoute() {
	user := r.App.Group("/api/v1/user", r.AuthenticationMiddleware.Authorize, r.AuthorizationMiddleware.AuthorizeUser)

	user.Patch("/password", r.AuthController.UpdatePassword)

	user.Post("/permissions", r.PermissionController.CreatePermission)
	user.Get("/permissions", r.PermissionController.GetUserPermissions)
	user.Get("/permissions/:id", r.PermissionController.GetPermissionByID)
	user.Put("/permissions/:id", r.PermissionController.UpdatePermission)          
	user.Patch("/permissions/:id/cancel", r.PermissionController.CancelPermission)   
	user.Delete("/permissions/:id", r.PermissionController.DeletePermission)
}
