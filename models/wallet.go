package models

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Wallet struct {
	ID      uuid.UUID `gorm:"type:uuid;primary_key;"`
	Address string    `gorm:"unique;not null"`
	Balance int       `gorm:"not null"`
}

func (wallet *Wallet) BeforeCreate(tx *gorm.DB) (err error) {
	wallet.ID = uuid.New()
	return
}

func InitializeWallet(db *gorm.DB, address string, initialBalance int) error {
	var wallet Wallet
	result := db.Where("address = ?", address).First(&wallet)
	if result.Error == gorm.ErrRecordNotFound {
		wallet = Wallet{
			Address: address,
			Balance: initialBalance,
		}
		return db.Create(&wallet).Error
	}
	return result.Error
}
