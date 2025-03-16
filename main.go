package main

import (
	"fmt"
	"log"
	"token-transfer-api/db"
)

func main() {
	_, err := db.Connect()
	if err != nil {
		log.Fatalf("Failed to connect to the database: %v", &err)
	}

	fmt.Println("Successfully connected to the database.")
}
