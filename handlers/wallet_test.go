package handlers

import (
	"bytes"
	"database/sql"
	"encoding/json"

	// "gin-wallet2/models"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestWalletHandler_Deposit(t *testing.T) {
	gin.SetMode(gin.TestMode)
	// Create a new mock database
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to create mock database: %v", err)
	}
	defer db.Close()

	// Create handler with mock DB
	handler := NewWalletHandler(db)

	// Create test cases
	tests := []struct {
		name           string
		requestBody    map[string]interface{}
		setupMock      func(sqlmock.Sqlmock)
		expectedStatus int
		expectedBody   map[string]interface{}
	}{
		{
			name: "successful deposit",
			requestBody: map[string]interface{}{
				"user_id": 1,
				"amount":  100.0,
			},
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectBegin()
				mock.ExpectExec("UPDATE users").
					WithArgs(100.0, 1).
					WillReturnResult(sqlmock.NewResult(0, 1))
				mock.ExpectExec("INSERT INTO transactions").
					WithArgs(1, "deposit", 100.0, "Deposit to wallet").
					WillReturnResult(sqlmock.NewResult(1, 1))
				mock.ExpectCommit()
			},
			expectedStatus: http.StatusOK,
			expectedBody: map[string]interface{}{
				"message": "Deposit successful",
			},
		},
		{
			name: "invalid amount",
			requestBody: map[string]interface{}{
				"user_id": 1,
				"amount":  -100.0,
			},
			setupMock:      func(_ sqlmock.Sqlmock) {},
			expectedStatus: http.StatusBadRequest,
			expectedBody: map[string]interface{}{
				"error": "Invalid input",
			},
		},
		{
			name: "update balance rollback",
			requestBody: map[string]interface{}{
				"user_id": 1,
				"amount":  99999999999999999,
			},
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectBegin()
				mock.ExpectExec("UPDATE users").
					WithArgs(99999999999999999.0, 1).
					WillReturnError(sql.ErrConnDone) // Simulate insert error
				mock.ExpectRollback()
			},
			expectedStatus: http.StatusInternalServerError,
			expectedBody: map[string]interface{}{
				"error": "Failed to update balance",
			},
		},
		{
			name: "transaction rollback",
			requestBody: map[string]interface{}{
				"user_id": 1,
				"amount":  100.0,
			},
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectBegin()
				mock.ExpectExec("UPDATE users").
					WithArgs(100.0, 1).
					WillReturnResult(sqlmock.NewResult(0, 1))
				mock.ExpectExec("INSERT INTO transactions").
					WithArgs(1, "deposit", 100.0, "Deposit to wallet").
					WillReturnError(sql.ErrConnDone) // Simulate insert error
				mock.ExpectRollback()
			},
			expectedStatus: http.StatusInternalServerError,
			expectedBody: map[string]interface{}{
				"error": "Failed to record transaction",
			},
		},
		{
			name: "begin transaction error",
			requestBody: map[string]interface{}{
				"user_id": 1,
				"amount":  100.0,
			},
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectBegin().WillReturnError(sql.ErrConnDone)
			},
			expectedStatus: http.StatusInternalServerError,
			expectedBody: map[string]interface{}{
				"error": "Failed to start transaction",
			},
		},
	}

	// runHandlerTests(t, handler.Transfer, tests, mock)

	// Run tests
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mock expectations
			tt.setupMock(mock)

			// Create request
			jsonBody, _ := json.Marshal(tt.requestBody)
			req := httptest.NewRequest(http.MethodPost, "/deposit", bytes.NewBuffer(jsonBody))
			req.Header.Set("Content-Type", "application/json")

			// Create response recorder
			w := httptest.NewRecorder()

			// Setup Gin context
			gin.SetMode(gin.TestMode)
			c, _ := gin.CreateTestContext(w)
			c.Request = req

			// Call the handler
			handler.Deposit(c)

			// Assert response
			assert.Equal(t, tt.expectedStatus, w.Code)

			var response map[string]interface{}
			err := json.Unmarshal(w.Body.Bytes(), &response)
			assert.NoError(t, err)
			assert.Equal(t, tt.expectedBody, response)

			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestWalletHandler_Withdraw(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to create mock database: %v", err)
	}
	defer db.Close()

	_ = mock

	handler := NewWalletHandler(db)

	tests := []struct {
		name           string
		requestBody    map[string]interface{}
		setupMock      func(sqlmock.Sqlmock)
		expectedStatus int
		expectedBody   map[string]interface{}
	}{
		{
			name: "successful withdraw",
			requestBody: map[string]interface{}{
				"user_id": 1,
				"amount":  50.0,
			},
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectBegin()
				mock.ExpectQuery("SELECT balance FROM users").
					WithArgs(1).
					WillReturnRows(sqlmock.NewRows([]string{"balance"}).AddRow(100.0))
				mock.ExpectExec("UPDATE users").
					WithArgs(50.0, 1).
					WillReturnResult(sqlmock.NewResult(0, 1))
				mock.ExpectExec("INSERT INTO transactions").
					WithArgs(1, "withdraw", 50.0, "Withdraw from wallet").
					WillReturnResult(sqlmock.NewResult(1, 1))
				mock.ExpectCommit()
			},
			expectedStatus: http.StatusOK,
			expectedBody: map[string]interface{}{
				"message": "Withdraw successful",
			},
		},
		{
			name: "invalid amount",
			requestBody: map[string]interface{}{
				"user_id": 1,
				"amount":  -50.0,
			},
			setupMock:      func(_ sqlmock.Sqlmock) {},
			expectedStatus: http.StatusBadRequest,
			expectedBody: map[string]interface{}{
				"error": "Invalid input",
			},
		},
		{
			name: "insufficient balance",
			requestBody: map[string]interface{}{
				"user_id": 1,
				"amount":  150.0,
			},
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectBegin()
				mock.ExpectQuery("SELECT balance FROM users").
					WithArgs(1).
					WillReturnRows(sqlmock.NewRows([]string{"balance"}).AddRow(100.0))
				mock.ExpectRollback()
			},
			expectedStatus: http.StatusBadRequest,
			expectedBody: map[string]interface{}{
				"error": "Insufficient balance",
			},
		},
		{
			name: "update balance rollback",
			requestBody: map[string]interface{}{
				"user_id": 1,
				"amount":  50.0,
			},
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectBegin()
				mock.ExpectQuery("SELECT balance FROM users").
					WithArgs(1).
					WillReturnRows(sqlmock.NewRows([]string{"balance"}).AddRow(100.0))
				mock.ExpectExec("UPDATE users").
					WithArgs(50.0, 1).
					WillReturnError(sql.ErrConnDone) // Simulate insert error
				mock.ExpectRollback()
			},
			expectedStatus: http.StatusInternalServerError,
			expectedBody: map[string]interface{}{
				"error": "Failed to update balance",
			},
		},
		{
			name: "transaction rollback",
			requestBody: map[string]interface{}{
				"user_id": 1,
				"amount":  50.0,
			},
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectBegin()
				mock.ExpectQuery("SELECT balance FROM users").
					WithArgs(1).
					WillReturnRows(sqlmock.NewRows([]string{"balance"}).AddRow(100.0))
				mock.ExpectExec("UPDATE users").
					WithArgs(50.0, 1).
					WillReturnResult(sqlmock.NewResult(0, 1))
				mock.ExpectExec("INSERT INTO transactions").
					WithArgs(1, "withdraw", 50.0, "Withdraw from wallet").
					WillReturnError(sql.ErrConnDone) // Simulate insert error
				mock.ExpectRollback()
			},
			expectedStatus: http.StatusInternalServerError,
			expectedBody: map[string]interface{}{
				"error": "Failed to record transaction",
			},
		},
		{
			name: "begin transaction error",
			requestBody: map[string]interface{}{
				"user_id": 1,
				"amount":  50.0,
			},
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectBegin().WillReturnError(sql.ErrConnDone)
			},
			expectedStatus: http.StatusInternalServerError,
			expectedBody: map[string]interface{}{
				"error": "Failed to start transaction",
			},
		},
	}

	runHandlerTests(t, handler.Withdraw, tests, mock)
}

func TestWalletHandler_Transfer(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to create mock database: %v", err)
	}
	defer db.Close()

	_ = mock

	handler := NewWalletHandler(db)

	tests := []struct {
		name           string
		requestBody    map[string]interface{}
		setupMock      func(sqlmock.Sqlmock)
		expectedStatus int
		expectedBody   map[string]interface{}
	}{
		{
			name: "successful transfer",
			requestBody: map[string]interface{}{
				"from_user_id": 1,
				"to_user_id":   2,
				"amount":       50.0,
			},
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("SELECT id FROM users").
					WithArgs(2).
					WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(2))
				mock.ExpectBegin()
				mock.ExpectQuery("SELECT balance FROM users").
					WithArgs(1).
					WillReturnRows(sqlmock.NewRows([]string{"balance"}).AddRow(100.0))
				mock.ExpectExec("UPDATE users").
					WithArgs(50.0, 1).
					WillReturnResult(sqlmock.NewResult(0, 1))
				mock.ExpectExec("UPDATE users").
					WithArgs(50.0, 2).
					WillReturnResult(sqlmock.NewResult(0, 1))
				mock.ExpectExec("INSERT INTO transactions").
					WithArgs(1, "transfer", 50.0, "Transfer to user 2").
					WillReturnResult(sqlmock.NewResult(1, 1))
				mock.ExpectCommit()
			},
			expectedStatus: http.StatusOK,
			expectedBody: map[string]interface{}{
				"message": "Transfer successful",
			},
		},
		{
			name: "invalid amount",
			requestBody: map[string]interface{}{
				"user_id": 1,
				"amount":  -100.0,
			},
			setupMock:      func(_ sqlmock.Sqlmock) {},
			expectedStatus: http.StatusBadRequest,
			expectedBody: map[string]interface{}{
				"error": "Invalid input",
			},
		},
		{
			name: "recipient user not found",
			requestBody: map[string]interface{}{
				"from_user_id": 1,
				"to_user_id":   999,
				"amount":       50.0,
			},
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("SELECT id FROM users").
					WithArgs(999).
					WillReturnError(sql.ErrNoRows)
			},
			expectedStatus: http.StatusUnauthorized,
			expectedBody: map[string]interface{}{
				"error": "Invalid credentials",
			},
		},
		{
			name: "Insufficient balance",
			requestBody: map[string]interface{}{
				"from_user_id": 1,
				"to_user_id":   2,
				"amount":       150.0,
			},
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("SELECT id FROM users").
					WithArgs(2).
					WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(2))
				mock.ExpectBegin()
				mock.ExpectQuery("SELECT balance FROM users").
					WithArgs(1).
					WillReturnRows(sqlmock.NewRows([]string{"balance"}).AddRow(100.0)).
					WillReturnError(sql.ErrConnDone)
				mock.ExpectRollback()
			},
			expectedStatus: http.StatusBadRequest,
			expectedBody: map[string]interface{}{
				"error": "Insufficient balance",
			},
		},
		{
			name: "begin transaction error",
			requestBody: map[string]interface{}{
				"from_user_id": 1,
				"to_user_id":   2,
				"amount":       50.0,
			},
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("SELECT id FROM users").
					WithArgs(2).
					WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(2))
				mock.ExpectBegin().WillReturnError(sql.ErrConnDone)
			},
			expectedStatus: http.StatusInternalServerError,
			expectedBody: map[string]interface{}{
				"error": "Failed to start transaction",
			},
		},
	}

	runHandlerTests(t, handler.Transfer, tests, mock)
}

func TestWalletHandler_GetBalance(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to create mock database: %v", err)
	}
	defer db.Close()

	handler := NewWalletHandler(db)

	tests := []struct {
		name           string
		userID         string
		setupMock      func(sqlmock.Sqlmock)
		expectedStatus int
		expectedBody   map[string]interface{}
	}{
		{
			name:   "successful balance query",
			userID: "1",
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("SELECT balance FROM users").
					WithArgs("1").
					WillReturnRows(sqlmock.NewRows([]string{"balance"}).AddRow(100.0))
			},
			expectedStatus: http.StatusOK,
			expectedBody: map[string]interface{}{
				"balance": float64(100.0),
			},
		},
		{
			name:   "user not found",
			userID: "999",
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("SELECT balance FROM users").
					WithArgs("999").
					WillReturnError(sql.ErrNoRows)
			},
			expectedStatus: http.StatusNotFound,
			expectedBody: map[string]interface{}{
				"error": "User not found",
			},
		},
		{
			name:   "database error",
			userID: "999",
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("SELECT balance FROM users").
					WithArgs("999").
					WillReturnError(sql.ErrConnDone)
			},
			expectedStatus: http.StatusInternalServerError,
			expectedBody: map[string]interface{}{
				"error": "Database error",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMock(mock)

			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Params = []gin.Param{{Key: "userID", Value: tt.userID}}

			handler.GetBalance(c)

			assert.Equal(t, tt.expectedStatus, w.Code)

			var response map[string]interface{}
			err := json.Unmarshal(w.Body.Bytes(), &response)
			assert.NoError(t, err)
			assert.Equal(t, tt.expectedBody, response)

			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestWalletHandler_GetTransactions(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to create mock database: %v", err)
	}
	defer db.Close()

	_ = mock

	handler := NewWalletHandler(db)

	tests := []struct {
		name           string
		userID         string
		setupMock      func(sqlmock.Sqlmock)
		expectedStatus int
		expectedBody   map[string]interface{}
	}{
		{
			name:   "successful transactions query",
			userID: "1",
			setupMock: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{"id", "type", "amount", "description", "created_at"}).
					AddRow(1, "deposit", 100.0, "Deposit to wallet", "2024-01-01 10:00:00").
					AddRow(2, "withdraw", 50.0, "Withdraw from wallet", "2024-01-02 10:00:00")
				mock.ExpectQuery("SELECT (.+) FROM transactions").
					WithArgs("1").
					WillReturnRows(rows)
			},
			expectedStatus: http.StatusOK,
			expectedBody: map[string]interface{}{
				"transactions": []interface{}{
					map[string]interface{}{
						"id":          float64(1),
						"type":        "deposit",
						"amount":      float64(100.0),
						"description": "Deposit to wallet",
						"created_at":  "2024-01-01 10:00:00",
					},
					map[string]interface{}{
						"id":          float64(2),
						"type":        "withdraw",
						"amount":      float64(50.0),
						"description": "Withdraw from wallet",
						"created_at":  "2024-01-02 10:00:00",
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMock(mock)

			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Params = []gin.Param{{Key: "userID", Value: tt.userID}}

			handler.GetTransactions(c)

			assert.Equal(t, tt.expectedStatus, w.Code)

			var response map[string]interface{}
			err := json.Unmarshal(w.Body.Bytes(), &response)
			assert.NoError(t, err)
			assert.Equal(t, tt.expectedBody, response)

			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

// Helper function to run handler tests
func runHandlerTests(t *testing.T, handlerFunc func(*gin.Context), tests []struct {
	name           string
	requestBody    map[string]interface{}
	setupMock      func(sqlmock.Sqlmock)
	expectedStatus int
	expectedBody   map[string]interface{}
}, mock sqlmock.Sqlmock) {
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMock(mock)

			jsonBody, _ := json.Marshal(tt.requestBody)
			req := httptest.NewRequest(http.MethodPost, "/", bytes.NewBuffer(jsonBody))
			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = req

			handlerFunc(c)

			assert.Equal(t, tt.expectedStatus, w.Code)

			var response map[string]interface{}
			err := json.Unmarshal(w.Body.Bytes(), &response)
			assert.NoError(t, err)
			assert.Equal(t, tt.expectedBody, response)

			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}
