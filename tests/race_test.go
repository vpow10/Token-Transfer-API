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

	fromAddress := "0xTEST2000"
	toAddress := "0xTEST2001"

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

	walletA := "0xTEST3000"
	walletB := "0xTEST3001"

	err := models.InitializeWallet(suite.db, walletA, initialBalance)
	assert.NoError(suite.T(), err, "Failed to initalize sender wallet")
	err = models.InitializeWallet(suite.db, walletB, initialBalance)
	assert.NoError(suite.T(), err, "Failed to initialize receiver wallet")

	start := make(chan struct{})
	results := make(chan error, num*2)
	var wg sync.WaitGroup

	// A to B transfers
	for i := 0; i < num; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			<-start
			_, err := suite.resolver.Transfer(context.Background(), walletA, walletB, amount)
			results <- err
		}()
	}

	// B to A transfers
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

func (suite *GraphQLTestSuite) TestConcurrentInsufficientFunds() {
	initialBalance := 10
	num := 50
	amount := 1

	fromAddress := "0xTEST4000"
	toAddress := "0xTEST4001"

	err := models.InitializeWallet(suite.db, fromAddress, initialBalance)
	assert.NoError(suite.T(), err, "Failed to initalize sender wallet")
	err = models.InitializeWallet(suite.db, toAddress, 0)
	assert.NoError(suite.T(), err, "Failed to initialize receiver wallet")

	start := make(chan struct{})
	results := make(chan error, num)
	var wg sync.WaitGroup

	for i := 0; i < num; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			<-start
			_, err := suite.resolver.Transfer(context.Background(), fromAddress, toAddress, amount)
			results <- err
		}()
	}

	close(start)
	wg.Wait()
	close(results)

	successCount := 0
	failCount := 0
	for err := range results {
		if err == nil {
			successCount++
		} else {
			assert.Equal(suite.T(), "insufficient balance", err.Error())
			failCount++
		}
	}

	assert.Equal(suite.T(), initialBalance, successCount, "Success transfers not correct")
	assert.Equal(suite.T(), num-initialBalance, failCount, "Failed transfers not correct")

	var finalSender, finalReceiver models.Wallet
	suite.db.Where("address =?", fromAddress).First(&finalSender)
	suite.db.Where("address =?", toAddress).First(&finalReceiver)

	assert.Equal(suite.T(), 0, finalSender.Balance)
	assert.Equal(suite.T(), initialBalance, finalReceiver.Balance)
}

func (suite *GraphQLTestSuite) TestConcurrentHighVolume() {
	initialBalance := 10000
	num := 500
	amount := 1

	fromAddress := "0xTEST5000"
	toAddress := "0xTEST5001"

	err := models.InitializeWallet(suite.db, fromAddress, initialBalance)
	assert.NoError(suite.T(), err)
	err = models.InitializeWallet(suite.db, toAddress, 0)
	assert.NoError(suite.T(), err)

	start := make(chan struct{})
	results := make(chan error, num)
	var wg sync.WaitGroup

	sem := make(chan struct{}, 50)

	for i := 0; i < num; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()

			<-start
			_, err := suite.resolver.Transfer(context.Background(), fromAddress, toAddress, amount)
			results <- err
		}()
	}

	close(start)
	wg.Wait()
	close(results)

	for err := range results {
		assert.NoError(suite.T(), err)
	}

	var finalSender, finalReceiver models.Wallet
	suite.db.Where("address = ?", fromAddress).First(&finalSender)
	suite.db.Where("address = ?", toAddress).First(&finalReceiver)

	assert.Equal(suite.T(), initialBalance-num, finalSender.Balance)
	assert.Equal(suite.T(), num, finalReceiver.Balance)
}

func (suite *GraphQLTestSuite) TestConcurrentVariousAmounts() {
	initialBalance := 10000
	num := 100

	walletA := "0xTEST6000"
	walletB := "0xTEST6001"

	err := models.InitializeWallet(suite.db, walletA, initialBalance)
	assert.NoError(suite.T(), err)
	err = models.InitializeWallet(suite.db, walletB, initialBalance)
	assert.NoError(suite.T(), err)

	type transferResult struct {
		from   string
		to     string
		amount int
		err    error
	}

	start := make(chan struct{})
	results := make(chan transferResult, num*2)
	var wg sync.WaitGroup

	sem := make(chan struct{}, 20)

	for i := 0; i < num; i++ {
		amount := 1 + (i % 5) // 1-5

		// A to B
		wg.Add(1)
		go func(amt int) {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()

			<-start
			_, err := suite.resolver.Transfer(context.Background(), walletA, walletB, amt)
			results <- transferResult{walletA, walletB, amt, err}
		}(amount)

		// B to A
		wg.Add(1)
		go func(amt int) {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()

			<-start
			_, err := suite.resolver.Transfer(context.Background(), walletB, walletA, amt)
			results <- transferResult{walletB, walletA, amt, err}
		}(amount)
	}

	close(start)
	wg.Wait()
	close(results)

	netFlowA := 0
	netFlowB := 0
	for res := range results {
		assert.NoError(suite.T(), res.err)
		if res.from == walletA {
			netFlowA -= res.amount
			netFlowB += res.amount
		} else {
			netFlowA += res.amount
			netFlowB -= res.amount
		}
	}

	var finalA, finalB models.Wallet
	suite.db.Where("address = ?", walletA).First(&finalA)
	suite.db.Where("address = ?", walletB).First(&finalB)

	assert.Equal(suite.T(), 0, netFlowA)
	assert.Equal(suite.T(), 0, netFlowB)

	assert.Equal(suite.T(), initialBalance, finalA.Balance)
	assert.Equal(suite.T(), initialBalance, finalB.Balance)
}

func (suite *GraphQLTestSuite) TestConcurrentSelfTransfers() {
	initialBalance := 1000
	num := 50
	amount := 1

	wallet := "0xTEST7000"

	err := models.InitializeWallet(suite.db, wallet, initialBalance)
	assert.NoError(suite.T(), err)

	start := make(chan struct{})
	results := make(chan error, num)
	var wg sync.WaitGroup

	sem := make(chan struct{}, 20)

	for i := 0; i < num; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()

			<-start
			_, err := suite.resolver.Transfer(context.Background(), wallet, wallet, amount)
			results <- err
		}()
	}

	close(start)
	wg.Wait()
	close(results)

	for err := range results {
		assert.Error(suite.T(), err)
		assert.Contains(suite.T(), err.Error(), "cannot transfer to self")
	}

	var finalWallet models.Wallet
	suite.db.Where("address = ?", wallet).First(&finalWallet)
	assert.Equal(suite.T(), initialBalance, finalWallet.Balance)
}
