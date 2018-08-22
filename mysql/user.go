package mysql

import (
	"time"
)

type User struct {
	ID               int        `gorm:"not null;primary_key;AUTO_INCREMENT"`
	CreatedAt        *time.Time `gorm:"not null"`
	UpdatedAt        *time.Time `gorm:"not null"`
	PlayerID         int64      `gorm:"not null"`
	SettingsEmail    string     `gorm:"not null"`
	SettingsPassword string     `gorm:"not null"`
	SettingsHidden   bool       `gorm:"not null"`
	SettingsAlerts   bool       `gorm:"not null"`
}
