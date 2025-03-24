package graph

// THIS CODE WILL BE UPDATED WITH SCHEMA CHANGES. PREVIOUS IMPLEMENTATION FOR SCHEMA CHANGES WILL BE KEPT IN THE COMMENT SECTION. IMPLEMENTATION FOR UNCHANGED SCHEMA WILL BE KEPT.

import (
	"context"
	"errors"
	"token-transfer-api/graph/generated"
	"token-transfer-api/models"

	"gorm.io/gorm"
)

type Resolver struct {
	DB *gorm.DB
}

// Transfer is the resolver for the transfer field.
func (r *Resolver) Transfer(ctx context.Context, fromAddress string, toAddress string, amount int) (*models.Wallet, error) {
	if amount <= 0 {
		return nil, errors.New("amount must be positive")
	}

	tx := r.DB.Begin()
	if tx.Error != nil {
		return nil, tx.Error
	}

	var fromWallet models.Wallet
	err := tx.Where("address =?", fromAddress).First(&fromWallet).Error
	if err != nil {
		tx.Rollback()
		return nil, errors.New("sender wallet not found")
	}

	if fromWallet.Balance < amount {
		tx.Rollback()
		return nil, errors.New("insufficient balance")
	}

	var toWallet models.Wallet
	err = tx.Where("address =?", toAddress).First(&toWallet).Error
	if err != nil {
		return nil, errors.New("receiver wallet not found")
	}

	fromWallet.Balance -= amount
	toWallet.Balance += amount

	err = tx.Save(&fromWallet).Error
	if err != nil {
		tx.Rollback()
		return nil, err
	}

	err = tx.Save(&toWallet).Error
	if err != nil {
		tx.Rollback()
		return nil, err
	}

	err = tx.Commit().Error
	if err != nil {
		tx.Rollback()
		return nil, err
	}

	return &fromWallet, nil
}

// ID is the resolver for the id field.
func (r *walletResolver) ID(ctx context.Context, obj *models.Wallet) (string, error) {
	return obj.ID.String(), nil
}

// Mutation returns generated.MutationResolver implementation.
func (r *Resolver) Mutation() generated.MutationResolver { return &mutationResolver{r} }

// Wallet returns generated.WalletResolver implementation.
func (r *Resolver) Wallet() generated.WalletResolver { return &walletResolver{r} }

type mutationResolver struct{ *Resolver }
type walletResolver struct{ *Resolver }
