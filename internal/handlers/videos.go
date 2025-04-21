package handlers

import (
	"fmt"
	"net/http"
	"path/filepath"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/ahmadbasyouni10/videogpt/internal/models"
	"github.com/ahmadbasyouni10/videogpt/pkg/supabase"
)

// VideoHandler handles video-related requests
type VideoHandler struct {
	SupabaseClient *supabase.Client
}

// NewVideoHandler creates a new video handler
func NewVideoHandler(supabaseClient *supabase.Client) *VideoHandler {
	return &VideoHandler{
		SupabaseClient: supabaseClient,
	}
}

// UploadVideo handles video upload
func (h *VideoHandler) UploadVideo(c echo.Context) error {
	// Get form values
	title := c.FormValue("title")
	description := c.FormValue("description")
	
	// Get file from form
	file, fileHeader, err := c.Request().FormFile("video")
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "No video file provided"})
	}
	defer file.Close()
	
	// Validate file type
	ext := filepath.Ext(fileHeader.Filename)
	if ext != ".mp4" && ext != ".mov" && ext != ".avi" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Unsupported file format"})
	}
	
	// Generate unique ID for video
	videoID := uuid.New().String()
	
	// Build path for storage
	path := fmt.Sprintf("videos/%s%s", videoID, ext)
	
	// Upload file to Supabase
	fileURL, err := h.SupabaseClient.UploadFile("videos", path, file, fileHeader)
	if err != nil {
		// Print detailed error for debugging
		fmt.Printf("Supabase error: %v\n", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to upload video"})
	}
	
	// Create video object
	video := models.Video{
		ID:          videoID,
		Title:       title,
		Description: description,
		FilePath:    path,
		UploadedAt:  time.Now(),
		Status:      "pending",
		ThumbnailURL: fileURL,
	}
	
	// Here you would normally save the video to the database
	// For now, we'll just return the video object
	
	return c.JSON(http.StatusCreated, video)
} 