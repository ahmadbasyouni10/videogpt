package summarization

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"
)

// Service defines the interface for summarization services
type Service interface {
	SummarizeText(text string) (string, error)
}

// OpenAIService implements the Service interface using OpenAI's API
type OpenAIService struct {
	APIKey  string
	Timeout time.Duration
}

// NewOpenAIService creates a new OpenAI summarization service
func NewOpenAIService(apiKey string) *OpenAIService {
	return &OpenAIService{
		APIKey:  apiKey,
		Timeout: 60 * time.Second, // Default timeout of 60 seconds
	}
}

// OpenAIRequest represents a request to OpenAI's API
type OpenAIRequest struct {
	Model       string        `json:"model"`
	Messages    []ChatMessage `json:"messages"`
	Temperature float64       `json:"temperature"`
}

// ChatMessage represents a message in a chat conversation
type ChatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// OpenAIResponse represents the response from OpenAI's API
type OpenAIResponse struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
	Error *struct {
		Message string `json:"message"`
	} `json:"error,omitempty"`
}

// SummarizeText sends text to OpenAI's API for summarization
func (s *OpenAIService) SummarizeText(text string) (string, error) {
	if s.APIKey == "" {
		return "", errors.New("OpenAI API key is required")
	}

	// Prepare the request
	prompt := fmt.Sprintf("Please provide a concise summary of the following transcript. Focus on the main topics, key points, and important details:\n\n%s", text)

	request := OpenAIRequest{
		Model: "gpt-3.5-turbo", // Using a cheaper model for cost-effectiveness
		Messages: []ChatMessage{
			{
				Role:    "system",
				Content: "You are an assistant that summarizes video transcripts clearly and concisely.",
			},
			{
				Role:    "user",
				Content: prompt,
			},
		},
		Temperature: 0.3, // Lower temperature for more focused and consistent outputs
	}

	// Convert request to JSON
	jsonData, err := json.Marshal(request)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	// Create HTTP request
	req, err := http.NewRequest("POST", "https://api.openai.com/v1/chat/completions", bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("failed to create HTTP request: %w", err)
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")
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
		var errorResponse OpenAIResponse
		if err := json.NewDecoder(resp.Body).Decode(&errorResponse); err == nil && errorResponse.Error != nil {
			return "", fmt.Errorf("API request failed: %s", errorResponse.Error.Message)
		}
		return "", fmt.Errorf("API request failed with status %d", resp.StatusCode)
	}

	// Parse the response
	var result OpenAIResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("failed to decode response: %w", err)
	}

	// Check if we got any choices back
	if len(result.Choices) == 0 {
		return "", errors.New("no summary was generated")
	}

	return result.Choices[0].Message.Content, nil
}
