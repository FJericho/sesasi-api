package entity

import "time"

const (
	ADMIN    = "admin"
	VERIFIER = "verifier"
	USER     = "user"
)

type Account struct {
	ID        string    `gorm:"type:uuid;default:uuid_generate_v4();primaryKey"`
	Name      string    `gorm:"not null"`
	Email     string    `gorm:"uniqueIndex;not null"`
	Password  string    `gorm:"not null"`
	Role      string    `gorm:"default:'user'"`
	Verified  bool      `gorm:"default:false"`
	CreatedAt time.Time `gorm:"autoCreateTime"`

	Permissions []Permission `gorm:"foreignKey:AccountID"`
}
