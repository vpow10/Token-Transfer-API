package graph

// THIS CODE WILL BE UPDATED WITH SCHEMA CHANGES. PREVIOUS IMPLEMENTATION FOR SCHEMA CHANGES WILL BE KEPT IN THE COMMENT SECTION. IMPLEMENTATION FOR UNCHANGED SCHEMA WILL BE KEPT.

import (
	"context"
	"token-transfer-api/graph/generated"
	"token-transfer-api/models"
)

type Resolver struct{}

// Transfer is the resolver for the transfer field.
func (r *mutationResolver) Transfer(ctx context.Context, fromAddress string, toAddress string, amount int) (*models.Wallet, error) {
	panic("not implemented")
}

// ID is the resolver for the id field.
func (r *walletResolver) ID(ctx context.Context, obj *models.Wallet) (string, error) {
	panic("not implemented")
}

// Mutation returns generated.MutationResolver implementation.
func (r *Resolver) Mutation() generated.MutationResolver { return &mutationResolver{r} }

// Wallet returns generated.WalletResolver implementation.
func (r *Resolver) Wallet() generated.WalletResolver { return &walletResolver{r} }

type mutationResolver struct{ *Resolver }
type walletResolver struct{ *Resolver }

// !!! WARNING !!!
// The code below was going to be deleted when updating resolvers. It has been copied here so you have
// one last chance to move it out of harms way if you want. There are two reasons this happens:
//  - When renaming or deleting a resolver the old code will be put in here. You can safely delete
//    it when you're done.
//  - You have helper methods in this file. Move them out to keep these resolver files clean.
/*
	type Resolver struct{}
func (r *Resolver) Transfer(ctx context.Context, fromAddress string, toAddress string, amount int) (*models.Wallet, error) {
	if amount <= 0 {
		return nil, errors.New("amount must be positive")
	}

	tx := db.DB.Begin()
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
*/
