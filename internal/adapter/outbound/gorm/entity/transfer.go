package entity

import "gorm.io/gorm"

type Transfer struct {
	gorm.Model
	SourceAccountID      int64  `gorm:"not null"`
	DestinationAccountID int64  `gorm:"not null"`
	Amount               string `gorm:"type:numeric;not null"`
}
