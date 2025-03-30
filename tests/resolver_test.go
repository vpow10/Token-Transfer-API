package tests

import (
	"context"
	"fmt"
	"os"
	"testing"
	"token-transfer-api/graph"
	"token-transfer-api/models"

	"github.com/joho/godotenv"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type GraphQLTestSuite struct {
	suite.Suite
	db       *gorm.DB
	resolver *graph.Resolver
}

func (suite *GraphQLTestSuite) SetupSuite() {
	err := godotenv.Load("../.env")
	if err != nil {
		suite.T().Log("Warning: .env file not found")
	}

	user := os.Getenv("POSTGRES_USER")
	password := os.Getenv("POSTGRES_PASSWORD")
	dbname := os.Getenv("POSTGRES_DB")
	host := os.Getenv("POSTGRES_HOST")
	port := os.Getenv("POSTGRES_PORT")

	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable", host, user, password, dbname, port)

	database, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	assert.NoError(suite.T(), err, "Failed to connect to the database")

	err = database.AutoMigrate(&models.Wallet{})
	assert.NoError(suite.T(), err, "Failed to auto-migrate")

	suite.db = database

	suite.resolver = &graph.Resolver{DB: database}
}

// func (suite *GraphQLTestSuite) TearDownSuite() {
// 	suite.db.Migrator().DropTable(&models.Wallet{})
// }

func (suite *GraphQLTestSuite) SetupTest() {
	defaultAddress := "0x1000"
	defaultBalance := 10000
	err := models.InitializeWallet(suite.db, defaultAddress, defaultBalance)
	assert.NoError(suite.T(), err, "Failed to initialize the default test wallet")
}

// using that instead of TearDownSuite() because incorrect receiver address test is causing runtime error otherwise
func (suite *GraphQLTestSuite) TearDownTest() {
	suite.db.Exec("DELETE FROM wallets WHERE address LIKE '0xTEST%' OR address LIKE '0x1000'")
}

func (suite *GraphQLTestSuite) TestTransferSuccessful() {
	fromAddress := "0x1000"
	toAddress := "0xTEST1001"
	amount := 100

	err := models.InitializeWallet(suite.db, toAddress, 0)
	assert.NoError(suite.T(), err, "Failed to initialize receiver wallet")

	var senderWallet models.Wallet
	err = suite.db.Where("address =?", fromAddress).First(&senderWallet).Error
	assert.NoError(suite.T(), err, "Failed to find sender wallet")
	initialSenderBalance := senderWallet.Balance

	wallet, err := suite.resolver.Transfer(context.Background(), fromAddress, toAddress, amount)
	assert.NoError(suite.T(), err, "Failed to transfer funds")
	assert.Equal(suite.T(), initialSenderBalance-amount, wallet.Balance, "Sender balance incorrect")

	var receiverWallet models.Wallet
	err = suite.db.Where("address =?", toAddress).First(&receiverWallet).Error
	assert.NoError(suite.T(), err, "Failed to find receiver wallet")
	assert.Equal(suite.T(), amount, receiverWallet.Balance, "Receiver balance incorrect")
}

func (suite *GraphQLTestSuite) TestTransferInsufficientBalance() {
	fromAddress := "0x1000"
	toAddress := "0xTEST1002"
	amount := 2000000

	err := models.InitializeWallet(suite.db, toAddress, 0)
	assert.NoError(suite.T(), err, "Failed to initialize receiver wallet")

	var senderWallet models.Wallet
	err = suite.db.Where("address =?", fromAddress).First(&senderWallet).Error
	assert.NoError(suite.T(), err, "Failed to find sender wallet")
	initialSenderBalance := senderWallet.Balance

	_, err = suite.resolver.Transfer(context.Background(), fromAddress, toAddress, amount)
	assert.Error(suite.T(), err, "Expected insufficient balance error")
	err = suite.db.Where("address =?", fromAddress).First(&senderWallet).Error
	assert.NoError(suite.T(), err, "Failed to find sender wallet")
	assert.Equal(suite.T(), initialSenderBalance, senderWallet.Balance, "Sender balance is incorrect")

	var receiverWallet models.Wallet
	err = suite.db.Where("address =?", toAddress).First(&receiverWallet).Error
	assert.NoError(suite.T(), err, "Failed to find receiver wallet")
	assert.Equal(suite.T(), 0, receiverWallet.Balance, "Receiver balance incorrect")
}

func (suite *GraphQLTestSuite) TestTransferInvalidAddres1() {
	fromAddress := "0xTEST1009"
	toAddress := "0x1000"
	amount := 1

	err := models.InitializeWallet(suite.db, toAddress, 0)
	assert.NoError(suite.T(), err, "Failed to initialize sender wallet")

	_, err = suite.resolver.Transfer(context.Background(), fromAddress, toAddress, amount)
	assert.Error(suite.T(), err, "Expected sender wallet not found error")
	assert.Equal(suite.T(), "sender wallet not found", err.Error(), "Incorrect error message")

	var wallet models.Wallet
	err = suite.db.Where("address = ?", toAddress).First(&wallet).Error
	assert.NoError(suite.T(), err, "Failed to find receiver wallet")
	assert.Equal(suite.T(), 10000, wallet.Balance, "Receiver balance should not change")
}

func (suite *GraphQLTestSuite) TestTransferInvalidAddres2() {
	fromAddress := "0x1000"
	toAddress := "0xTEST1009"
	amount := 1

	err := models.InitializeWallet(suite.db, fromAddress, 1000)
	assert.NoError(suite.T(), err, "Failed to initialize sender wallet")

	_, err = suite.resolver.Transfer(context.Background(), fromAddress, toAddress, amount)
	assert.Error(suite.T(), err, "Expected receiver wallet not found error")
	assert.Equal(suite.T(), "receiver wallet not found", err.Error(), "Incorrect error message")

	var wallet models.Wallet
	err = suite.db.Where("address = ?", fromAddress).First(&wallet).Error
	assert.NoError(suite.T(), err, "Failed to find sender wallet")
	assert.Equal(suite.T(), 10000, wallet.Balance, "Sender balance should not change")
}

func TestGraphQLSuite(t *testing.T) {
	suite.Run(t, new(GraphQLTestSuite))
}
