package main

import (
	"log"
	"os"
	"ultra-chat-backend/config"
	"ultra-chat-backend/handlers"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"ultra-chat-backend/repositories"
)

func main() {

	db := config.ConnectDB()
	defer config.DisconnectDB()

	userRepo := repositories.NewUserRepository(db)
	summaryRepo, _ := repositories.NewMongoSummaryRepository(db)

	e := echo.New()
	e.Use(middleware.Recover())

	// Auth Routes
	authHandler := handlers.NewAuthHandler(userRepo)
	e.GET("/login", authHandler.Login)
	e.GET("/callback", authHandler.Callback)
	e.GET("/profile", authHandler.Profile)

	summaryHandler := handlers.NewSummaryHandler(summaryRepo)
	e.POST("/create-summary", summaryHandler.CreateSummary)
	e.GET("/summarizer", summaryHandler.GetSummaries)
	e.PUT("/update-summary", summaryHandler.UpdateSummary)
	e.DELETE("/delete-summary", summaryHandler.DeleteSummary)
	e.GET("/is_authenticated", summaryHandler.IsAuthenticated)

	port := os.Getenv("PORT")
	if port == "" {
		port = "5001"
	}

	log.Fatal(e.Start(":" + port))

}
