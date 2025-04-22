package main

import (
	"log"
	"os"
	"path/filepath"

	"github.com/ahmadbasyouni10/videogpt/internal/handlers"
	"github.com/ahmadbasyouni10/videogpt/pkg/ffmpeg"
	"github.com/ahmadbasyouni10/videogpt/pkg/supabase"
	"github.com/joho/godotenv"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

func main() {
	// Load environment variables from .env file
	err := godotenv.Load()
	if err != nil {
		log.Println("Warning: Error loading .env file, using system environment variables")
	}

	// Initialize Supabase client
	supabaseClient := supabase.NewClient()

	// Create temp directory for video processing
	tempDir := os.Getenv("TEMP_DIR")
	if tempDir == "" {
		tempDir = filepath.Join(os.TempDir(), "videogpt")
	}

	// Initialize FFmpeg processor
	ffmpegProcessor, err := ffmpeg.NewProcessor(tempDir)
	if err != nil {
		log.Fatalf("Failed to initialize FFmpeg processor: %v", err)
	}

	// Initialize handlers
	videoHandler := handlers.NewVideoHandler(supabaseClient, ffmpegProcessor)

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

	// Add route to get video details
	e.GET("/videos/:id", videoHandler.GetVideo)

	// Add route to get thumbnail
	e.GET("/thumbnails/:id", videoHandler.GetThumbnail)

	// Start server
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080" // Default port
	}
	e.Logger.Fatal(e.Start(":" + port))
}
