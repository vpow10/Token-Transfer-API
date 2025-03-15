package main

import (
	"fmt"
	"log"
	"token-transfer-api/db"
)

func main() {
	db, err := db.Connect()
	if err != nil {
		log.Fatalf("Failed to connect to the database: %v", err)
	}
	defer db.Close()

	fmt.Println("Successfully connected to the database!")
}
