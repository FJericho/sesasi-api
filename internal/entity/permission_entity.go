package entity

import "time"

const (
	StatusPending  = "pending"
	StatusApproved = "approved"
	StatusRejected = "rejected"
	StatusRevised  = "revised"
	StatusCancelled  = "cancelled"
)

type Permission struct {
	Id        string    `gorm:"type:uuid;default:uuid_generate_v4();primaryKey"`
	Title     string    `gorm:"not null"`
	Reason    string    `gorm:"not null"`
	StartDate time.Time `gorm:"not null"`
	EndDate   time.Time `gorm:"not null"`
	Comment   string
	Status    string    `gorm:"default:'pending'"`
	CreatedAt time.Time `gorm:"autoCreateTime"`
	UpdatedAt time.Time `gorm:"autoUpdateTime"`
	AccountID string    `gorm:"type:uuid;not null"`

	Account Account `gorm:"foreignKey:AccountID;references:ID"`
}
