package models

import (
	"time"
)

// Video represents a video in the system
type Video struct {
	ID           string    `json:"id"`
	Title        string    `json:"title"`
	Description  string    `json:"description"`
	FilePath     string    `json:"file_path"`
	Transcript   string    `json:"transcript,omitempty"`
	Summary      string    `json:"summary,omitempty"`
	UploadedAt   time.Time `json:"uploaded_at"`
	ProcessedAt  time.Time `json:"processed_at,omitempty"`
	Status       string    `json:"status"` // pending, processing, completed, failed
	ThumbnailURL string    `json:"thumbnail_url,omitempty"`
	Duration     float64   `json:"duration,omitempty"`
	VideoURL     string    `json:"video_url,omitempty"`
}
