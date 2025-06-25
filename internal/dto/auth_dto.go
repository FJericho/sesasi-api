package dto

import "time"

type AccountResponse struct {
	ID          string                `json:"id,omitempty"`
	Name        string                `json:"name,omitempty"`
	Email       string                `json:"email,omitempty"`
	Role        string                `json:"role,omitempty"`
	Verified    bool                  `json:"verified"`
	CreatedAt   *time.Time            `json:"created_at,omitempty"`
	Permissions []*PermissionResponse `json:"permissions,omitempty"`
}

type RegisterRequest struct {
	Name     string `json:"name" validate:"required,min=3,max=32"`
	Email    string `json:"email" validate:"required"`
	Password string `json:"password" validate:"required"`
}

type LoginRequest struct {
	Email    string `json:"email" validate:"required"`
	Password string `json:"password" validate:"required"`
}

type LoginResponse struct {
	Token string `json:"token,omitempty"`
}

type UpdatePasswordRequest struct {
	OldPassword string `json:"old_password" validate:"required,min=6"`
	NewPassword string `json:"new_password" validate:"required,min=6"`
}
