# Token Transfer API

A GraphQL API for transferring BTP tokens between wallets.

## Prerequisites

- Docker
- Docker Compose
- Go 1.16+

## 1. Installation & Running

1. Clone the repository:
   ```bash
   git clone https://github.com/vpow10/token-transfer-api.git
   cd token-transfer-api
   ```
2. Create and configure environment variables:
    ```bash
    cp .env.example .env
    ```

    Edit the .env file with your preferred values:

    ```
    POSTGRES_USER=tta_user
    POSTGRES_PASSWORD=secure_password
    POSTGRES_DB=tta_db
    POSTGRES_HOST=localhost
    POSTGRES_PORT=5432
    ```
3. Configure Docker (optional):
    Edit docker/docker-compose.yaml if you need custom container names or ports:
    ```yaml
    services:
      db:
        container_name: my_token_db  # Optional custom name
        ports:
        - "5432:5432"  # Map host port 5432 to container 5432
    ```
4. Start the services:
    ```bash
    cd docker/
    docker-compose up -d
    cd ..
    ```
5. Start the application:
    ```bash
    go run main.go
    ```
The GraphQL server will be available at http://localhost:8080/graphql

### Stopping the application

1. Stop the Go server (Ctrl+C in terminal)
2. Stop Docker services:
    ```bash
    cd docker/
    docker-compose down
    ```

## 2. Database management

To reset the database:

```bash
cd docker/
docker-compose down -v # Wipes all data
docker-compose up -d
cd ..
go run main.go
```

Access database (optional)

```bash
docker exec -it <container_name> psql -U tta_user -d tta_db
```

## 3. Example GraphQL Mutations

### Initial State
- Default wallet: 0x0000 with 1,000,000 BTP tokens

- All other wallets must be created by transfers to them

### Transfer Tokens
```graphql
mutation TransferTokens {
  transfer(
    fromAddress: "0x0000",
    toAddress: "0x1001",
    amount: 100
  ) {
    address
    balance
  }
}
```

### Successful Response:
```json
{
  "data": {
    "transfer": {
      "address": "0x0000",
      "balance": 999900
    }
  }
}
```

### Error cases:
1. Insufficient balance
    ```graphql
    mutation TransferTooMuch {
    transfer(
        fromAddress: "0x0000",
        toAddress: "0x1001",
        amount: 100000000
    ) {
        address
        balance
    }
    }
    ```
    Returns: ```"insufficient balance" ```

2. Nonexistent sender:
    ```graphql
    mutation InvalidSender {
    transfer(
        fromAddress: "0xinvalid",
        toAddress: "0x1001", 
        amount: 100
    ) {
        address
        balance
    }
    }
    ```
    Returns: ```"sender not found" ```
3. Nonexistent receiver
    ```graphql
        mutation InvalidRecipient {
    transfer(
        fromAddress: "0x0000",
        toAddress: "0xinvalid",
        amount: 100
    ) {
        address
        balance
    }
    }
    ```
    Returns: ```"receiver not found" ```

## Development notes
- **Concurrency**: The API safely handles concurrent transfers

- **Persistence**: Data persists across restarts when using Docker volumes

- **Testing**: See tests/ directory for comprehensive test cases. To run them:
    ```bash
    go test ./tests
    ```
