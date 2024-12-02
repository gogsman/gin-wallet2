package main

import (
	"gin-wallet2/handlers"
	"gin-wallet2/middleware"
	"log"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"github.com/rs/zerolog"
)

var zlog = zerolog.New(os.Stdout).With().Timestamp().Logger()

func init() {
	if err := godotenv.Load(); err != nil {
		log.Printf("Error loading .env file: %v", err)
	}

	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	zlog = zerolog.New(os.Stdout).With().Timestamp().Logger()
	zerolog.DefaultContextLogger = &zlog
}

func main() {
	db := InitDB()
	defer func() {
		if err := db.Close(); err != nil {
			log.Printf("failed to close database: %v", err)
		}
	}()

	r := gin.Default()

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
	zlog.Info().
		Str("port", port).
		Msg("Server starting")

	if err := r.Run(port); err != nil {
		zlog.Fatal().
			Err(err).
			Str("port", port).
			Msg("Server failed to start")
	}
}
