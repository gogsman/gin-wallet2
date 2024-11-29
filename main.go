package main

import (
	"gin-wallet2/handlers"
	"gin-wallet2/middleware"
	"log"

	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
)

func main() {
	db := InitDB()

	r := gin.Default()

	//setupRoutes(r)

	auth := handlers.NewAuthHandler(db)

	r.POST("/register", auth.Register)
	r.POST("/login", auth.Login)

	wallet := handlers.NewWalletHandler(db)
	walletGroup := r.Group("/wallet", middleware.AuthMiddleware())
	{
		walletGroup.POST("/deposit", wallet.Deposit)
		walletGroup.POST("/withdraw", wallet.Withdraw)
		walletGroup.POST("/transfer", wallet.Transfer)
		walletGroup.GET("/balance/:userID", wallet.GetBalance)
		walletGroup.GET("/transactions/:userID", wallet.GetTransactions)
	}

	port := ":8080"
	log.Println("Server running on port", port)
	if err := r.Run(port); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}
