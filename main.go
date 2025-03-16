package main

import (
	"fmt"
	"log"
	"token-transfer-api/db"
	"token-transfer-api/models"
)

func main() {
	database, err := db.Connect()
	if err != nil {
		log.Fatalf("Failed to connect to the database: %v", &err)
	}

	defaultAddress := "0x0000"
	initialBalance := 1000000
	err = models.InitializeWallet(database, defaultAddress, initialBalance)
	if err != nil {
		log.Fatalf("Failed to initialize the default wallet: %v", err)
	}

	fmt.Printf("Default wallet initialized with address: %s and balance %d\n", defaultAddress, initialBalance)
}
