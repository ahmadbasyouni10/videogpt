package supabase

import (
	"bytes"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
)

// Client represents a Supabase client
type Client struct {
	URL string
	Key string
}

// NewClient creates a new Supabase client
func NewClient() *Client {
	return &Client{
		URL: os.Getenv("SUPABASE_URL"),
		Key: os.Getenv("SUPABASE_KEY"),
	}
}

// UploadFile uploads a file to Supabase storage
func (c *Client) UploadFile(bucket string, path string, file multipart.File, fileHeader *multipart.FileHeader) (string, error) {
	// Read file content
	fileBytes, err := io.ReadAll(file)
	if err != nil {
		return "", fmt.Errorf("error reading file: %w", err)
	}

	// Create request
	url := fmt.Sprintf("%s/storage/v1/object/%s/%s", c.URL, bucket, path)
	req, err := http.NewRequest("POST", url, bytes.NewReader(fileBytes))
	if err != nil {
		return "", fmt.Errorf("error creating request: %w", err)
	}

	// Set headers
	req.Header.Set("Authorization", "Bearer "+c.Key)
	req.Header.Set("Content-Type", fileHeader.Header.Get("Content-Type"))

	// Send request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("error uploading file: %w", err)
	}
	defer resp.Body.Close()

	// Check response
	if resp.StatusCode != http.StatusOK {
		// Read response body for more details
		responseBody, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("error uploading file: status code %d, response: %s", resp.StatusCode, string(responseBody))
	}

	// Build and return the file URL
	fileURL := fmt.Sprintf("%s/storage/v1/object/public/%s/%s", c.URL, bucket, path)
	return fileURL, nil
}

// GetFile retrieves a file from Supabase storage
func (c *Client) GetFile(bucket string, path string) ([]byte, error) {
	// Create request
	url := fmt.Sprintf("%s/storage/v1/object/public/%s/%s", c.URL, bucket, path)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("error creating request: %w", err)
	}

	// Set headers
	req.Header.Set("Authorization", "Bearer "+c.Key)

	// Send request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error retrieving file: %w", err)
	}
	defer resp.Body.Close()

	// Check response
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("error retrieving file: status code %d", resp.StatusCode)
	}

	// Read and return the file content
	return io.ReadAll(resp.Body)
} 