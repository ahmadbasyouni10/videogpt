package main

import (
	"log"
	"os"
	"path/filepath"

	"github.com/ahmadbasyouni10/videogpt/internal/handlers"
	"github.com/ahmadbasyouni10/videogpt/pkg/ffmpeg"
	"github.com/ahmadbasyouni10/videogpt/pkg/summarization"
	"github.com/ahmadbasyouni10/videogpt/pkg/supabase"
	"github.com/ahmadbasyouni10/videogpt/pkg/transcription"
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

	// Initialize transcription service
	openaiApiKey := os.Getenv("OPENAI_API_KEY")
	if openaiApiKey == "" {
		log.Println("Warning: OPENAI_API_KEY not set, transcription and summarization services will not work")
	}
	transcriptionService := transcription.NewWhisperService(openaiApiKey)

	// Initialize summarization service (using the same OpenAI API key)
	summarizationService := summarization.NewOpenAIService(openaiApiKey)

	// Initialize handlers
	videoHandler := handlers.NewVideoHandler(supabaseClient, ffmpegProcessor, transcriptionService, summarizationService)

	// Initialize Echo instance
	e := echo.New()

	// Add middleware
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Use(middleware.CORS())

	// Serve static files from the static directory
	e.Static("/", "static")

	// Define API routes
	api := e.Group("/api")

	api.GET("/", func(c echo.Context) error {
		return c.String(200, "Welcome to VideoGPT API")
	})

	// Video routes
	api.POST("/videos", videoHandler.UploadVideo)

	// Add route to get video details
	api.GET("/videos/:id", videoHandler.GetVideo)

	// Add route to get thumbnail
	api.GET("/thumbnails/:id", videoHandler.GetThumbnail)

	// Add route to generate transcript
	api.POST("/videos/:id/transcript", videoHandler.GenerateTranscript)

	// Add route to generate summary
	api.POST("/videos/:id/summary", videoHandler.GenerateSummary)

	// Start server
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080" // Default port
	}
	e.Logger.Fatal(e.Start(":" + port))
}
