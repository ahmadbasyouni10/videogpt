package handlers

import (
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/ahmadbasyouni10/videogpt/internal/models"
	"github.com/ahmadbasyouni10/videogpt/pkg/ffmpeg"
	"github.com/ahmadbasyouni10/videogpt/pkg/supabase"
	"github.com/ahmadbasyouni10/videogpt/pkg/transcription"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

// VideoHandler handles video-related requests
type VideoHandler struct {
	SupabaseClient       *supabase.Client
	FFmpegProcessor      *ffmpeg.Processor
	TranscriptionService transcription.Service
}

// NewVideoHandler creates a new video handler
func NewVideoHandler(supabaseClient *supabase.Client, ffmpegProcessor *ffmpeg.Processor, transcriptionService transcription.Service) *VideoHandler {
	return &VideoHandler{
		SupabaseClient:       supabaseClient,
		FFmpegProcessor:      ffmpegProcessor,
		TranscriptionService: transcriptionService,
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

	// Save uploaded file to temp directory for processing
	tempFilePath := filepath.Join(h.FFmpegProcessor.TempDir, videoID+ext)
	tempFile, err := os.Create(tempFilePath)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to save uploaded file"})
	}

	// Copy uploaded file to temp file
	_, err = io.Copy(tempFile, file)
	tempFile.Close()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to save uploaded file"})
	}

	// Generate thumbnail
	thumbnailPath, err := h.FFmpegProcessor.CreateThumbnail(tempFilePath)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to create thumbnail"})
	}

	// Get video duration
	duration, err := h.FFmpegProcessor.GetVideoDuration(tempFilePath)
	if err != nil {
		// Non-fatal error, continue without duration
		fmt.Printf("Failed to get video duration: %v\n", err)
	}

	// Build path for storage
	videoPath := fmt.Sprintf("%s%s", videoID, ext)
	thumbnailStoragePath := fmt.Sprintf("%s.jpg", videoID)

	// Upload video file to Supabase
	// bucket, path in supabase, and then where to find the file
	videoURL, err := h.SupabaseClient.UploadFileFromPath("videos", videoPath, tempFilePath)
	if err != nil {
		fmt.Printf("Supabase video upload error: %v\n", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to upload video"})
	}
	fmt.Printf("DEBUG - Uploaded video URL: %s\n", videoURL)

	// Upload thumbnail to Supabase
	thumbnailFile, err := os.Open(thumbnailPath)
	if err != nil {
		fmt.Printf("Failed to open thumbnail file: %v\n", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to process thumbnail"})
	}
	defer thumbnailFile.Close()

	thumbnailFileInfo, err := thumbnailFile.Stat()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to get thumbnail file info"})
	}

	thumbnailHeader := &multipart.FileHeader{
		Filename: filepath.Base(thumbnailPath),
		Size:     thumbnailFileInfo.Size(),
		Header:   make(map[string][]string),
	}
	thumbnailHeader.Header.Set("Content-Type", "image/jpeg")

	thumbnailURL, err := h.SupabaseClient.UploadFile("thumbnails", thumbnailStoragePath, thumbnailFile, thumbnailHeader)
	if err != nil {
		fmt.Printf("Supabase thumbnail upload error: %v\n", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to upload thumbnail"})
	}
	fmt.Printf("DEBUG - Uploaded thumbnail URL: %s\n", thumbnailURL)

	// Create video object
	video := models.Video{
		ID:           videoID,
		Title:        title,
		Description:  description,
		FilePath:     videoPath,
		UploadedAt:   time.Now(),
		Status:       "processing",
		ThumbnailURL: thumbnailURL,
		Duration:     duration,
		VideoURL:     videoURL,
	}

	// Clean up temp files
	defer func() {
		os.Remove(tempFilePath)
		os.Remove(thumbnailPath)
	}()

	// Here you would save the video to the database
	// For now, we'll just return the video object

	return c.JSON(http.StatusCreated, video)
}

func (h *VideoHandler) GetThumbnail(c echo.Context) error {
	thumbnailID := c.Param("id")
	if thumbnailID == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Thumbnail ID is required"})
	}

	// Match the exact URL pattern returned by the updated upload function
	thumbnailURL := fmt.Sprintf("%s/storage/v1/object/public/thumbnails/%s.jpg",
		h.SupabaseClient.URL, thumbnailID)

	fmt.Printf("DEBUG - Redirecting to thumbnail URL: %s\n", thumbnailURL)
	return c.Redirect(http.StatusTemporaryRedirect, thumbnailURL)
}

// GetVideo retrieves video details
func (h *VideoHandler) GetVideo(c echo.Context) error {
	videoID := c.Param("id")
	if videoID == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Video ID is required"})
	}

	// Match the exact URL pattern returned by the updated upload function
	videoURL := fmt.Sprintf("%s/storage/v1/object/public/videos/%s.mp4",
		h.SupabaseClient.URL, videoID)

	fmt.Printf("DEBUG - Redirecting to video URL: %s\n", videoURL)
	return c.Redirect(http.StatusTemporaryRedirect, videoURL)
}

// GenerateTranscript generates a transcript for a video
func (h *VideoHandler) GenerateTranscript(c echo.Context) error {
	videoID := c.Param("id")
	if videoID == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Video ID is required"})
	}

	// Get direct URL to the video in Supabase using the same pattern as in GetVideo
	videoURL := fmt.Sprintf("%s/storage/v1/object/public/videos/%s.mp4",
		h.SupabaseClient.URL, videoID)

	// Download the video from the URL
	resp, err := http.Get(videoURL)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error":   "Failed to download video",
			"details": err.Error(),
		})
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return c.JSON(http.StatusNotFound, map[string]string{
			"error":   "Video not found",
			"details": fmt.Sprintf("Status code: %d", resp.StatusCode),
		})
	}

	// Save to temp file for processing
	tempFilePath := filepath.Join(h.FFmpegProcessor.TempDir, videoID+".mp4")
	tempFile, err := os.Create(tempFilePath)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to create temp file",
		})
	}

	// Copy the response body to the temp file
	_, err = io.Copy(tempFile, resp.Body)
	tempFile.Close()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to save video data",
		})
	}
	defer os.Remove(tempFilePath) // Clean up when done

	// Extract audio using your existing FFmpeg processor
	audioPath, err := h.FFmpegProcessor.ExtractAudio(tempFilePath)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error":   "Failed to extract audio",
			"details": err.Error(),
		})
	}
	defer os.Remove(audioPath) // Clean up audio file when done

	// Use the transcription service to convert audio to text
	transcript, err := h.TranscriptionService.TranscribeAudio(audioPath)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error":   "Failed to transcribe audio",
			"details": err.Error(),
		})
	}

	// TODO: Save transcript to database if needed

	return c.JSON(http.StatusOK, map[string]interface{}{
		"status":     "success",
		"message":    "Audio transcribed successfully",
		"video_id":   videoID,
		"transcript": transcript,
	})
}
