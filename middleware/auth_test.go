package middleware

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"runtime"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
)

func TestAuthMiddleware(t *testing.T) {
	runtime.GOMAXPROCS(1)
	// Setup
	gin.SetMode(gin.TestMode)

	// Helper function to create a valid token
	createToken := func(userID int) string {
		token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
			"userID": userID,
			"exp":    time.Now().Add(time.Hour * 24).Unix(),
		})
		tokenString, _ := token.SignedString(jwtSecret)
		return tokenString
	}

	tests := []struct {
		name           string
		setupHeader    func() string
		expectedStatus int
		expectedUserID interface{}
	}{
		{
			name:           "No Authorization header",
			setupHeader:    func() string { return "" },
			expectedStatus: http.StatusUnauthorized,
			expectedUserID: nil,
		},
		{
			name:           "Invalid Authorization format",
			setupHeader:    func() string { return "InvalidFormat" },
			expectedStatus: http.StatusUnauthorized,
			expectedUserID: nil,
		},
		{
			name:           "Invalid token",
			setupHeader:    func() string { return "Bearer invalid.token.here" },
			expectedStatus: http.StatusUnauthorized,
			expectedUserID: nil,
		},
		{
			name:           "Valid token",
			setupHeader:    func() string { return "Bearer " + createToken(123) },
			expectedStatus: http.StatusOK,
			expectedUserID: 123,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup test router
			w := httptest.NewRecorder()
			_, r := gin.CreateTestContext(w)

			// Add test endpoint with middleware
			r.Use(AuthMiddleware())
			r.GET("/test", func(c *gin.Context) {
				userID, exists := c.Get("userID")
				if exists {
					c.JSON(http.StatusOK, gin.H{"userID": userID})
				}
			})

			// Create test request
			req := httptest.NewRequest("GET", "/test", nil)
			req.Header.Set("Authorization", tt.setupHeader())

			// Serve the request
			r.ServeHTTP(w, req)

			// Assert response status
			assert.Equal(t, tt.expectedStatus, w.Code)

			// If expecting successful auth, verify userID
			if tt.expectedUserID != nil {
				var response map[string]interface{}
				err := json.NewDecoder(w.Body).Decode(&response)
				assert.NoError(t, err)
				assert.Equal(t, float64(tt.expectedUserID.(int)), response["userID"])
			}
		})
	}
}
