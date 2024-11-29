package handlers

import (
	"database/sql"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

type WalletHandler struct {
	DB *sql.DB
}

func NewWalletHandler(db *sql.DB) *WalletHandler {
	return &WalletHandler{DB: db}
}

func (h *WalletHandler) Deposit(c *gin.Context) {
	var req struct {
		UserID int     `json:"user_id"`
		Amount float64 `json:"amount"`
	}
	if err := c.ShouldBindJSON(&req); err != nil || req.Amount <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
		return
	}

	tx, err := h.DB.Begin()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to start transaction"})
		return
	}

	_, err = tx.Exec("UPDATE users SET balance = balance + $1 WHERE id = $2", req.Amount, req.UserID)
	if err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update balance"})
		return
	}

	_, err = tx.Exec("INSERT INTO transactions (user_id, type, amount, description) VALUES ($1, $2, $3, $4)",
		req.UserID, "deposit", req.Amount, "Deposit to wallet")
	if err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to record transaction"})
		return
	}

	tx.Commit()
	c.JSON(http.StatusOK, gin.H{"message": "Deposit successful"})
}

func (h *WalletHandler) Withdraw(c *gin.Context) {
	var req struct {
		UserID int     `json:"user_id"`
		Amount float64 `json:"amount"`
	}
	if err := c.ShouldBindJSON(&req); err != nil || req.Amount <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
		return
	}

	tx, err := h.DB.Begin()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to start transaction"})
		return
	}

	var balance float64
	err = tx.QueryRow("SELECT balance FROM users WHERE id = $1", req.UserID).Scan(&balance)
	if err != nil || balance < req.Amount {
		tx.Rollback()
		c.JSON(http.StatusBadRequest, gin.H{"error": "Insufficient balance"})
		return
	}

	_, err = tx.Exec("UPDATE users SET balance = balance - $1 WHERE id = $2", req.Amount, req.UserID)
	if err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update balance"})
		return
	}

	_, err = tx.Exec("INSERT INTO transactions (user_id, type, amount, description) VALUES ($1, $2, $3, $4)",
		req.UserID, "withdraw", req.Amount, "Withdraw from wallet")
	if err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to record transaction"})
		return
	}

	tx.Commit()
	c.JSON(http.StatusOK, gin.H{"message": "Withdraw successful"})
}

func (h *WalletHandler) Transfer(c *gin.Context) {
	var req struct {
		FromUserID int     `json:"from_user_id"`
		ToUserID   int     `json:"to_user_id"`
		Amount     float64 `json:"amount"`
	}
	if err := c.ShouldBindJSON(&req); err != nil || req.Amount <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
		return
	}

	var toUser int
	err := h.DB.QueryRow("SELECT id FROM users WHERE id = $1", req.ToUserID).Scan(&toUser)
	if err == sql.ErrNoRows {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	} else if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to query user"})
		return
	}

	tx, err := h.DB.Begin()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to start transaction"})
		return
	}

	var balance float64
	err = tx.QueryRow("SELECT balance FROM users WHERE id = $1", req.FromUserID).Scan(&balance)
	if err != nil || balance < req.Amount {
		tx.Rollback()
		c.JSON(http.StatusBadRequest, gin.H{"error": "Insufficient balance"})
		return
	}

	_, err = tx.Exec("UPDATE users SET balance = balance - $1 WHERE id = $2", req.Amount, req.FromUserID)
	if err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to deduct balance"})
		return
	}

	_, err = tx.Exec("UPDATE users SET balance = balance + $1 WHERE id = $2", req.Amount, req.ToUserID)
	if err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to credit balance"})
		return
	}

	_, err = tx.Exec("INSERT INTO transactions (user_id, type, amount, description) VALUES ($1, $2, $3, $4)",
		req.FromUserID, "transfer", req.Amount, "Transfer to user "+fmt.Sprint(req.ToUserID))
	if err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to record transaction"})
		return
	}

	tx.Commit()
	c.JSON(http.StatusOK, gin.H{"message": "Transfer successful"})
}

func (h *WalletHandler) GetBalance(c *gin.Context) {
	userID := c.Param("userID")
	var balance float64

	err := h.DB.QueryRow("SELECT balance FROM users WHERE id = $1", userID).Scan(&balance)
	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	} else if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"balance": balance})
}

func (h *WalletHandler) GetTransactions(c *gin.Context) {
	userID := c.Param("userID")
	var transactions []struct {
		ID          int     `json:"id"`
		Type        string  `json:"type"`
		Amount      float64 `json:"amount"`
		Description string  `json:"description"`
		CreatedAt   string  `json:"created_at"`
	}

	rows, err := h.DB.Query("SELECT id, type, amount, description, created_at FROM transactions WHERE user_id = $1 ORDER BY created_at DESC", userID)
	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"error": "No transactions found for user"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		}
		return
	}
	defer rows.Close()

	for rows.Next() {
		var tx struct {
			ID          int     `json:"id"`
			Type        string  `json:"type"`
			Amount      float64 `json:"amount"`
			Description string  `json:"description"`
			CreatedAt   string  `json:"created_at"`
		}
		if err := rows.Scan(&tx.ID, &tx.Type, &tx.Amount, &tx.Description, &tx.CreatedAt); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error reading transaction data"})
			return
		}
		transactions = append(transactions, tx)
	}

	c.JSON(http.StatusOK, gin.H{"transactions": transactions})
}
