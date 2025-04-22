package supabase

import (
	"bytes"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
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
// tested good
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

// UploadFileFromPath uploads a file to Supabase storage from a local file path
// TESTED and is good
func (c *Client) UploadFileFromPath(bucket string, path string, filePath string) (string, error) {
	// Open the file
	file, err := os.Open(filePath)
	if err != nil {
		return "", fmt.Errorf("error opening file: %w", err)
	}
	defer file.Close()

	// Get file info
	fileInfo, err := file.Stat()
	if err != nil {
		return "", fmt.Errorf("error getting file info: %w", err)
	}

	// Create file header
	fileHeader := &multipart.FileHeader{
		Filename: filepath.Base(filePath),
		Size:     fileInfo.Size(),
		Header:   make(map[string][]string),
	}

	// Set content type based on file extension
	contentType := "application/octet-stream"
	ext := filepath.Ext(filePath)
	if ext == ".mp4" {
		contentType = "video/mp4"
	} else if ext == ".mov" {
		contentType = "video/quicktime"
	} else if ext == ".avi" {
		contentType = "video/x-msvideo"
	}
	fileHeader.Header.Set("Content-Type", contentType)

	// Use the existing UploadFile method
	return c.UploadFile(bucket, path, file, fileHeader)
}

// GetFile retrieves a file from Supabase storage

func (c *Client) GetFile(bucket string, path string) ([]byte, error) {
	// Create request - matching the format used in UploadFile
	url := fmt.Sprintf("%s/storage/v1/object/%s/%s", c.URL, bucket, path)
	fmt.Printf("DEBUG - Fetching from URL: %s\n", url) // Add debug output
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
