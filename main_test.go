package main

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"gin-wallet2/handlers"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

// Add at the top of the test file
func setupTestRouter(handler *handlers.WalletHandler) *gin.Engine {
	r := gin.New()
	walletGroup := r.Group("/wallet")
	walletGroup.Use(func(c *gin.Context) {
		// Mock authentication middleware
		c.Set("userID", uint(1)) // Set a default authenticated user
		c.Next()
	})
	{
		walletGroup.POST("/deposit", handler.Deposit)
		walletGroup.POST("/withdraw", handler.Withdraw)
		walletGroup.POST("/transfer", handler.Transfer)
		walletGroup.GET("/balance/:userID", handler.GetBalance)
		walletGroup.GET("/transactions/:userID", handler.GetTransactions)
	}
	return r
}

// Update the test functions to use the router
func TestWalletHandler_Deposit(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to create mock database: %v", err)
	}
	defer db.Close()

	handler := handlers.NewWalletHandler(db)
	router := setupTestRouter(handler)

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
		{
			name: "update balance error",
			requestBody: map[string]interface{}{
				"user_id": 1,
				"amount":  100.0,
			},
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectBegin()
				mock.ExpectExec("UPDATE users").
					WithArgs(100.0, 1).
					WillReturnError(sql.ErrConnDone)
				mock.ExpectRollback()
			},
			expectedStatus: http.StatusInternalServerError,
			expectedBody: map[string]interface{}{
				"error": "Failed to update balance",
			},
		},
		{
			name: "record transaction error",
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
					WillReturnError(sql.ErrConnDone)
				mock.ExpectRollback()
			},
			expectedStatus: http.StatusInternalServerError,
			expectedBody: map[string]interface{}{
				"error": "Failed to record transaction",
			},
		},
	}

	// Update the test execution
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMock(mock)

			jsonBody, _ := json.Marshal(tt.requestBody)
			req := httptest.NewRequest(http.MethodPost, "/wallet/deposit", bytes.NewBuffer(jsonBody))
			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)

			var response map[string]interface{}
			err := json.Unmarshal(w.Body.Bytes(), &response)
			assert.NoError(t, err)
			assert.Equal(t, tt.expectedBody, response)

			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}
