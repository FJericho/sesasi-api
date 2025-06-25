package middleware

import (
	"errors"
	"strings"
	"time"

	"github.com/FJericho/sesasi-api/internal/entity"
	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v4"
	"github.com/spf13/viper"
)

type Authentication interface {
	GenerateToken(id, email, role string, name string) (string, error)
	GetCurrentUserAccount(ctx *fiber.Ctx) entity.Account
	Authorize(ctx *fiber.Ctx) error
}

type AuthenticationMiddleware struct {
	Viper *viper.Viper
}

func NewAuthenticationMiddleware(viper *viper.Viper) Authentication {
	return &AuthenticationMiddleware{
		Viper: viper,
	}
}

func (a *AuthenticationMiddleware) GenerateToken(id string, email string, role string, name string) (string, error) {
	if id == "" || email == "" || role == "" {
		return "", fiber.ErrBadRequest
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": id,
		"email":   email,
		"name":    name,
		"role":    role,
		"exp":     time.Now().Add(time.Hour * 8).Unix(),
	})

	tokenStr, err := token.SignedString([]byte(a.Viper.GetString("jwt.secret_key")))
	if err != nil {
		return "", errors.New("unable to sign the token")
	}

	return tokenStr, nil
}

func (a *AuthenticationMiddleware) GetCurrentUserAccount(ctx *fiber.Ctx) entity.Account {
	user, ok := ctx.Locals("user").(entity.Account)

	if !ok {
		return entity.Account{}
	}
	return user
}

func (a *AuthenticationMiddleware) Authorize(ctx *fiber.Ctx) error {
	authHeader := ctx.Get("Authorization")

	if authHeader == "" {
		return ctx.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"message": "missing authorization header",
		})
	}

	user, err := a.verifyToken(authHeader)
	if err != nil || user.ID == "" {
		return ctx.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"message": "authorization failed",
		})
	}

	ctx.Locals("user", user)

	return ctx.Next()
}

func (a *AuthenticationMiddleware) verifyToken(authHeader string) (entity.Account, error) {
	var user entity.Account

	if !strings.HasPrefix(authHeader, "Bearer ") {
		return user, errors.New("authorization token is malformed")
	}

	tokenString := strings.TrimPrefix(authHeader, "Bearer ")

	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return []byte(a.Viper.GetString("jwt.secret_key")), nil
	})

	if err != nil {
		return user, err
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		user.ID = claims["user_id"].(string)
		user.Email = claims["email"].(string)
		user.Role = claims["role"].(string)

		expiration := int64(claims["exp"].(float64))
		if time.Now().Unix() > expiration {
			return user, errors.New("token has expired")
		}

		return user, nil
	}

	return user, errors.New("invalid token")
}
