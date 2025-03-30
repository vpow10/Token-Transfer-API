package graph

import (
	"context"
	"errors"
	"token-transfer-api/graph/generated"
	"token-transfer-api/models"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type Resolver struct {
	DB *gorm.DB
}

// Transfer is the resolver for the transfer field.
func (r *Resolver) Transfer(ctx context.Context, fromAddress string, toAddress string, amount int) (*models.Wallet, error) {
	if amount <= 0 {
		return nil, errors.New("amount must be positive")
	}

	if fromAddress == toAddress {
		return nil, errors.New("cannot transfer to self")
	}

	// Determine lock order (always lock the "lower" address first)
	firstToLock, secondToLock := fromAddress, toAddress
	if fromAddress > toAddress {
		firstToLock, secondToLock = toAddress, fromAddress
	}

	db := r.DB.WithContext(ctx)

	tx := db.Session(&gorm.Session{
		SkipDefaultTransaction: true,
		PrepareStmt:            true,
	}).Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// Lock first wallet
	var firstWallet models.Wallet
	if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
		Where("address = ?", firstToLock).
		First(&firstWallet).Error; err != nil {
		tx.Rollback()
		if errors.Is(err, gorm.ErrRecordNotFound) {
			if firstToLock == fromAddress {
				return nil, errors.New("sender wallet not found")
			}
			return nil, errors.New("receiver wallet not found")
		}
		return nil, err
	}

	// Lock second wallet
	var secondWallet models.Wallet
	if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
		Where("address = ?", secondToLock).
		First(&secondWallet).Error; err != nil {
		tx.Rollback()
		if errors.Is(err, gorm.ErrRecordNotFound) {
			if secondToLock == fromAddress {
				return nil, errors.New("sender wallet not found")
			}
			return nil, errors.New("receiver wallet not found")
		}
		return nil, err
	}

	// Determine which wallet is sender and which is receiver
	var fromWallet, toWallet *models.Wallet
	if firstToLock == fromAddress {
		fromWallet, toWallet = &firstWallet, &secondWallet
	} else {
		fromWallet, toWallet = &secondWallet, &firstWallet
	}

	if fromWallet.Balance < amount {
		tx.Rollback()
		return nil, errors.New("insufficient balance")
	}

	fromWallet.Balance -= amount
	toWallet.Balance += amount

	if err := tx.Save(fromWallet).Error; err != nil {
		tx.Rollback()
		return nil, err
	}
	if err := tx.Save(toWallet).Error; err != nil {
		tx.Rollback()
		return nil, err
	}

	if err := tx.Commit().Error; err != nil {
		return nil, err
	}

	return fromWallet, nil
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
