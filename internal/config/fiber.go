package config

import (
	"github.com/FJericho/sesasi-api/internal/dto"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/spf13/viper"
)

func NewFiberConfig(viper *viper.Viper) *fiber.App {
	var app = fiber.New(fiber.Config{
		AppName:      viper.GetString("app.name"),
		ErrorHandler: NewErrorHandler(),
		Prefork:      viper.GetBool("web.prefork"),
	})

	app.Use(cors.New(cors.Config{
		AllowOrigins: "*", 
		AllowHeaders: "Origin, Content-Type, Accept, Authorization",
		AllowMethods: "GET, POST, PUT, DELETE, PATCH",
	}))

	return app
}

func NewErrorHandler() fiber.ErrorHandler {
	return func(ctx *fiber.Ctx, err error) error {
		code := fiber.StatusInternalServerError

		if fiberErr, ok := err.(*fiber.Error); ok {
			code = fiberErr.Code
		}

		return ctx.Status(code).JSON(dto.WebResponse[any]{
			Errors: &dto.ErrorResponse{Message: err.Error()} ,
		})
	}
}