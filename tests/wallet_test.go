package tests

import (
	"fmt"
	"os"
	"testing"
	"token-transfer-api/models"

	"github.com/joho/godotenv"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type WalletTestSuite struct {
	suite.Suite
	db *gorm.DB
}

func (suite *WalletTestSuite) SetupSuite() {
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

	suite.db = database

	err = suite.db.AutoMigrate(&models.Wallet{})
	assert.NoError(suite.T(), err, "Failed to auto-migrate Wallet model")
}

func (suite *WalletTestSuite) TearDownSuite() {
	suite.db.Migrator().DropTable(&models.Wallet{})
}

func (suite *WalletTestSuite) TestInitWallet() {
	address := "0x0001"
	initialBalance := 20
	err := models.InitializeWallet(suite.db, address, initialBalance)
	assert.NoError(suite.T(), err, "Failed to initialize wallet")

	var wallet models.Wallet
	result := suite.db.Where("address =?", address).First(&wallet)
	assert.NoError(suite.T(), result.Error, "Wallet not found in database")
	assert.Equal(suite.T(), address, wallet.Address, "Wallet address is not correct")
	assert.Equal(suite.T(), initialBalance, wallet.Balance, "Wallet balance is not correct")
}

func (suite *WalletTestSuite) TestInitDuplicetWallet() {
	address := "0x0002"
	initialBalance := 10
	err := models.InitializeWallet(suite.db, address, initialBalance)
	assert.NoError(suite.T(), err, "Failed to initialize wallet")
	var wallet models.Wallet

	err = models.InitializeWallet(suite.db, address, 20)
	assert.NoError(suite.T(), err, "Failed to handle duplicate wallet creation")

	result := suite.db.Where("address =?", address).First(&wallet)
	assert.NoError(suite.T(), result.Error, "Failed to find wallet in database")
	assert.Equal(suite.T(), initialBalance, wallet.Balance, "Wallet balance updated during duplication handling")
}

func (suite *WalletTestSuite) TestNegativeBalanceInit() {
	address := "0x0010"
	initialBalance := -10
	err := models.InitializeWallet(suite.db, address, initialBalance)
	assert.Error(suite.T(), err, "Expected error for negative balance initialization")
}

func (suite *WalletTestSuite) TestEmptyAddressInit() {
	address := ""
	initialBalance := 10
	err := models.InitializeWallet(suite.db, address, initialBalance)
	assert.Error(suite.T(), err, "Expected error for empty address initialization")
}

func TestWalletSuite(t *testing.T) {
	suite.Run(t, new(WalletTestSuite))
}
