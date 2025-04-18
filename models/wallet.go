package models

import (
	"errors"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Wallet struct {
	ID      uuid.UUID `gorm:"type:uuid;primary_key;"`
	Address string    `gorm:"unique;not null"`
	Balance int       `gorm:"not null"`
	Version int       `gorm:"default:1"`
}

func (wallet *Wallet) BeforeCreate(tx *gorm.DB) (err error) {
	wallet.ID = uuid.New()
	return
}

func InitializeWallet(db *gorm.DB, address string, initialBalance int) error {
	if address == "" {
		return errors.New("address cannot be empty")
	}
	if initialBalance < 0 {
		return errors.New("balance cannot be negative")
	}

	var wallet Wallet
	result := db.Where("address = ?", address).First(&wallet)
	if result.Error != nil && result.Error != gorm.ErrRecordNotFound {
		return result.Error
	}
	if result.Error == gorm.ErrRecordNotFound {
		wallet = Wallet{
			Address: address,
			Balance: initialBalance,
		}
		return db.Create(&wallet).Error
	}
	return nil
}
