package entity

import "gorm.io/gorm"

type Account struct {
	gorm.Model
	AccountID int64  `gorm:"uniqueIndex;not null"`
	Balance   string `gorm:"type:numeric;not null"`
}
