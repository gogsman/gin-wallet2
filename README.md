# Gin Wallet API

A wallet API service based on the Gin framework, providing user authentication and wallet transaction functionality.

## Features

- User Authentication (Register/Login)
- Wallet Operations:
  - Deposit
  - Withdrawal
  - Transfer
  - Balance Query
  - Transaction History

## Quick Start

### Prerequisites

- Go 1.x
- PostgreSQL
- Docker (optional)

### Installation

1. Clone the repository

```
  git clone <repository-url>
  cd gin-wallet2
```

2. Install dependencies

```
  go mod tidy
```

3. Configure database
- Create PostgreSQL database
- Update database configuration in `.env` file

4. Run service
```
 go run .
```

The service will start on port `:8080`


### Docker Setup

1. Build the Docker image
```
  docker build -t gin-wallet .
```

2. Run with Docker Compose
```
  docker-compose up -d
```


This will start:
- PostgreSQL container
- Gin Wallet API container
- All necessary networking

Example docker-compose.yml:

## API Endpoints

### Authentication Endpoints
- `POST /register` - User registration
- `POST /login` - User login

### Wallet Endpoints (Authentication Required)
- `POST /wallet/deposit` - Deposit funds
- `POST /wallet/withdraw` - Withdraw funds
- `POST /wallet/transfer` - Transfer funds
- `GET /wallet/balance/:userID` - Check balance
- `GET /wallet/transactions/:userID` - View transaction history

## Architecture Decisions

1. **Layered Architecture**
  - handlers: API handling layer
  - middleware: Middleware layer
  - models: Data model layer
  - Clear separation of concerns

2. **Security Considerations**
  - JWT authentication
  - Middleware protection for sensitive routes
  - Parameter validation

3. **Testability**
  - Dependency injection
  - Database mocking for tests
  - Comprehensive test cases

## Code Review Focus

1. **handlers Package**
  - Business logic implementation
  - Error handling approach
  - Parameter validation

2. **middleware Package**
  - JWT authentication implementation
  - Middleware chain

3. **Test Cases**
  - Test methods in `xxxx_test.go`
  - Mock usage patterns
## Future Improvements

- [ ] Add more unit tests
- [ ] Implement transaction rollback mechanism
- [ ] Add rate limiting
- [ ] Enhance logging system
- [ ] Add API documentation

## License

[MIT](LICENSE)