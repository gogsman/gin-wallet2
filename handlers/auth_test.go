package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"golang.org/x/crypto/bcrypt"
)

func TestRegister(t *testing.T) {
	gin.SetMode(gin.TestMode)

	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	authHandler := NewAuthHandler(db)

	router := gin.Default()
	router.POST("/register", authHandler.Register)

	mock.ExpectExec("INSERT INTO users").WithArgs("testuser", sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(1, 1))

	body := map[string]string{
		"name":     "testuser",
		"password": "password123",
	}
	jsonBody, _ := json.Marshal(body)

	req, _ := http.NewRequest("POST", "/register", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)
	assert.JSONEq(t, `{"message": "User created successfully"}`, w.Body.String())

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestLogin(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	authHandler := NewAuthHandler(db)

	router := gin.Default()
	router.POST("/login", authHandler.Login)

	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
	mock.ExpectQuery("SELECT id, password_hash FROM users WHERE name = \\$1").
		WithArgs("testuser").
		WillReturnRows(sqlmock.NewRows([]string{"id", "password_hash"}).AddRow(1, string(hashedPassword)))

	body := map[string]string{
		"name":     "testuser",
		"password": "password123",
	}
	jsonBody, _ := json.Marshal(body)

	req, _ := http.NewRequest("POST", "/login", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var respBody map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &respBody)
	assert.NoError(t, err)
	assert.Contains(t, respBody, "token")

	assert.NoError(t, mock.ExpectationsWereMet())
}
