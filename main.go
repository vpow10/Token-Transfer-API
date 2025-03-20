package main

import (
	"log"
	"net/http"
	"token-transfer-api/db"
	"token-transfer-api/graph"
	"token-transfer-api/graph/generated"
	"token-transfer-api/models"

	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/playground"
)

func main() {
	database, err := db.Connect()
	if err != nil {
		log.Fatalf("Failed to connect to the database: %v", err)
	}

	defaultAddress := "0x0000"
	initialBalance := 1000000
	err = models.InitializeWallet(database, defaultAddress, initialBalance)
	if err != nil {
		log.Fatalf("Failed to initialize the default wallet: %v", err)
	}

	resolver := &graph.Resolver{DB: database}

	srv := handler.NewDefaultServer(generated.NewExecutableSchema(generated.Config{Resolvers: resolver}))

	http.Handle("/", playground.Handler("GraphQL Playground", "/query"))
	http.Handle("/query", srv)

	log.Println("GraphQL server is running on http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
