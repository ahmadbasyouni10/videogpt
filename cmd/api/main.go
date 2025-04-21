package main

import (
	"log"
	"os"
	"github.com/joho/godotenv"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/ahmadbasyouni10/videogpt/internal/handlers"
	"github.com/ahmadbasyouni10/videogpt/pkg/supabase"
)

func main() {
	// Load environment variables from .env file
	err := godotenv.Load()
	if err != nil {
		log.Println("Warning: Error loading .env file, using system environment variables")
	}

	// Initialize Supabase client
	supabaseClient := supabase.NewClient()

	// Initialize handlers
	videoHandler := handlers.NewVideoHandler(supabaseClient)

	// Initialize Echo instance
	e := echo.New()

	// Add middleware
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Use(middleware.CORS())

	// Define routes
	e.GET("/", func(c echo.Context) error {
		return c.String(200, "Welcome to VideoGPT API")
	})
	
	// Video routes
	e.POST("/videos", videoHandler.UploadVideo)

	// Start server
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080" // Default port
	}
	e.Logger.Fatal(e.Start(":" + port))
} 