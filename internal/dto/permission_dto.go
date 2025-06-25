package dto

import (
	"time"

	"github.com/FJericho/sesasi-api/internal/entity"
)

type PermissionResponse struct {
	ID        string           `json:"id,omitempty"`
	Title     string           `json:"title,omitempty"`
	Reason    string           `json:"reason,omitempty"`
	StartDate time.Time        `json:"start_date,omitempty"`
	EndDate   time.Time        `json:"end_date,omitempty"`
	Comment   string           `json:"comment,omitempty"`
	Status    string           `json:"status,omitempty"`
	CreatedAt time.Time        `json:"created_at,omitempty"`
	UpdatedAt time.Time        `json:"updated_at,omitempty"`
	Account   *AccountResponse `json:"account,omitempty"`
}

type ApprovalRequest struct {
	Comment string `json:"comment" validate:"required"`
}

type CreatePermissionRequest struct {
	Title     string    `json:"title" validate:"required"`
	Reason    string    `json:"reason" validate:"required"`
	StartDate time.Time `json:"start_date" validate:"required"`
	EndDate   time.Time `json:"end_date" validate:"required"`
}

func ToPermissionResponse(p *entity.Permission) *PermissionResponse {
	if p == nil {
		return nil
	}
	return &PermissionResponse{
		ID:        p.Id,
		Title:     p.Title,
		Reason:    p.Reason,
		StartDate: p.StartDate,
		EndDate:   p.EndDate,
		Comment:   p.Comment,
		Status:    p.Status,
		CreatedAt: p.CreatedAt,
		UpdatedAt: p.UpdatedAt,
		Account: &AccountResponse{
			ID:       p.Account.ID,
			Name:     p.Account.Name,
			Email:    p.Account.Email,
			Verified: p.Account.Verified,
		},
	}
}

func ToPermissionResponseList(list []*entity.Permission) []*PermissionResponse {
	var result []*PermissionResponse
	for _, p := range list {
		result = append(result, ToPermissionResponse(p))
	}
	return result
}
