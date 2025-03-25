package tests

import (
	"context"
	"sync"
	"token-transfer-api/models"

	"github.com/stretchr/testify/assert"
)

func (suite *GraphQLTestSuite) TestConcurrentTransfers() {
	initialBalance := 1000
	num := 100
	amount := 1

	fromAddress := "0x2000"
	toAddress := "0x2001"

	err := models.InitializeWallet(suite.db, fromAddress, initialBalance)
	assert.NoError(suite.T(), err, "Failed to initialize sender wallet")
	err = models.InitializeWallet(suite.db, toAddress, 0)
	assert.NoError(suite.T(), err, "Failed to initialize receiver wallet")

	results := make(chan error, num)
	var wg sync.WaitGroup

	for i := 0; i < num; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, err := suite.resolver.Transfer(context.Background(), fromAddress, toAddress, amount)
			results <- err
		}()
	}

	wg.Wait()
	close(results)

	for err := range results {
		assert.NoError(suite.T(), err, "Transfer failed")
	}

	var finalSender, finalReceiver models.Wallet
	suite.db.Where("address =?", fromAddress).First(&finalSender)
	suite.db.Where("address =?", toAddress).First(&finalReceiver)

	assert.Equal(suite.T(), initialBalance-(num*amount), finalSender.Balance, "Wrong final sender balance")
	assert.Equal(suite.T(), num*amount, finalReceiver.Balance, "Wrong final receiver balance")
}

func (suite *GraphQLTestSuite) TestConcurrentMixedTransfers() {
	initialBalance := 1000
	num := 50
	amount := 1

	walletA := "0x3000"
	walletB := "0x3001"

	err := models.InitializeWallet(suite.db, walletA, initialBalance)
	assert.NoError(suite.T(), err, "Failed to initalize sender wallet")
	err = models.InitializeWallet(suite.db, walletB, initialBalance)
	assert.NoError(suite.T(), err, "Failed to initialize receiver wallet")

	start := make(chan struct{})
	results := make(chan error, num*2)
	var wg sync.WaitGroup

	for i := 0; i < num; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			<-start
			_, err := suite.resolver.Transfer(context.Background(), walletA, walletB, amount)
			results <- err
		}()
	}

	for i := 0; i < num; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			<-start
			_, err := suite.resolver.Transfer(context.Background(), walletB, walletA, amount)
			results <- err
		}()
	}

	close(start)
	wg.Wait()
	close(results)

	for err := range results {
		assert.NoError(suite.T(), err, "Transfer failed")
	}

	var finalA, finalB models.Wallet
	suite.db.Where("address =?", walletA).First(&finalA)
	suite.db.Where("address =?", walletB).First(&finalB)

	assert.Equal(suite.T(), initialBalance, finalA.Balance, "Wrong final wallet A balance")
	assert.Equal(suite.T(), initialBalance, finalB.Balance, "Wrong final wallet B balance")
}
