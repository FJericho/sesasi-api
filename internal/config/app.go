package config

import (
	"github.com/FJericho/sesasi-api/internal/controller"
	"github.com/FJericho/sesasi-api/internal/middleware"
	"github.com/FJericho/sesasi-api/internal/repository"
	"github.com/FJericho/sesasi-api/internal/router"
	"github.com/FJericho/sesasi-api/internal/service"
	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"gorm.io/gorm"
)

type AppConfig struct {
	DB       *gorm.DB
	App      *fiber.App
	Log      *logrus.Logger
	Validate *validator.Validate
	Config   *viper.Viper
}

func StartServer(config *AppConfig) {
	authRepository := repository.NewAuthRepository(config.DB)
	permissionRepository := repository.NewPermissionRepository(config.DB)

	authenticationMiddleware := middleware.NewAuthenticationMiddleware(config.Config)
	authorizationMiddleware := middleware.NewAuthorizationMiddleware(authenticationMiddleware)

	authService := service.NewAuthService(config.Log, config.Config, config.Validate, authRepository, authenticationMiddleware)
	permissionService := service.NewPermissionService(config.Log, config.Config, config.Validate, permissionRepository)

	authController := controller.NewAuthController(config.Log, authService, authenticationMiddleware)
	permissionController := controller.NewPermissionController(config.Log, permissionService, authenticationMiddleware)

	routeConfig := router.RouteConfig{
		App:                      config.App,
		AuthController:           authController,
		AuthenticationMiddleware: authenticationMiddleware,
		AuthorizationMiddleware:  authorizationMiddleware,
		PermissionController:     permissionController,
	}

	routeConfig.Setup()
}
