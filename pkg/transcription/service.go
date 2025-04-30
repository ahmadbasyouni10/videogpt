package transcription

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"time"
)

// Service defines the interface for transcription services
type Service interface {
	TranscribeAudio(audioPath string) (string, error)
}

// WhisperService implements the Service interface using OpenAI's Whisper API
type WhisperService struct {
	APIKey  string
	Timeout time.Duration
}

// NewWhisperService creates a new Whisper transcription service
func NewWhisperService(apiKey string) *WhisperService {
	return &WhisperService{
		APIKey:  apiKey,
		Timeout: 5 * time.Minute, // Default timeout of 5 minutes
	}
}

// WhisperResponse represents the response from OpenAI's Whisper API
type WhisperResponse struct {
	Text string `json:"text"`
}

// TranscribeAudio sends audio to OpenAI's Whisper API for transcription
func (s *WhisperService) TranscribeAudio(audioPath string) (string, error) {
	if s.APIKey == "" {
		return "", errors.New("OpenAI API key is required")
	}

	file, err := os.Open(audioPath)
	if err != nil {
		return "", fmt.Errorf("failed to open audio file: %w", err)
	}
	defer file.Close()

	// Create a new HTTP request
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	// Add the file to the request
	part, err := writer.CreateFormFile("file", "audio.mp3")
	if err != nil {
		return "", fmt.Errorf("failed to create form file: %w", err)
	}

	if _, err = io.Copy(part, file); err != nil {
		return "", fmt.Errorf("failed to copy file to request: %w", err)
	}

	// Add the model parameter
	if err = writer.WriteField("model", "whisper-1"); err != nil {
		return "", fmt.Errorf("failed to add model field: %w", err)
	}

	// Close the writer
	if err = writer.Close(); err != nil {
		return "", fmt.Errorf("failed to close writer: %w", err)
	}

	// Create the HTTP request
	req, err := http.NewRequest("POST", "https://api.openai.com/v1/audio/transcriptions", body)
	if err != nil {
		return "", fmt.Errorf("failed to create HTTP request: %w", err)
	}

	// Set headers
	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", s.APIKey))

	// Create HTTP client with timeout
	client := &http.Client{
		Timeout: s.Timeout,
	}

	// Send the request
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	// Check the response status
	if resp.StatusCode != http.StatusOK {
		responseBody, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(responseBody))
	}

	// Parse the response
	var result WhisperResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("failed to decode response: %w", err)
	}

	return result.Text, nil
}
