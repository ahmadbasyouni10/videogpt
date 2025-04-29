package ffmpeg

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// this processor will process with ffmpeg and the temp directory will be used to store the video and audio files
type Processor struct {
	TempDir string
}

// this is a constructor for the Processor struct
func NewProcessor(tempDir string) (*Processor, error) {
	// get info about temp directory
	// if statement for checking if os.isNotExist then create it if not made
	// os.isnotexist(err) to check if error is a not exist error
	if _, err := os.Stat(tempDir); os.IsNotExist(err) {
		// mkdir expects a path and a mode (0755 default means that everyone can read/execute but owner is only one with write)
		err := os.MkdirAll(tempDir, 0755)
		if err != nil {
			// wraps eror for checking not like %v which is for printing
			// fmt.Sprintf is for formatting strings
			// fmt.Errorf is for formatting errors
			return nil, fmt.Errorf("failed to create temp directory: %w", err)
		}
	}

	// check ffmpeg installed
	_, err := exec.LookPath("ffmpeg")
	if err != nil {
		return nil, fmt.Errorf("ffmpeg not found in path: %w", err)
	}

	return &Processor{
		TempDir: tempDir,
	}, nil
}

// this function will extract audio
// it is a method on the Processor struct
// like making methods for a class using self.
func (p *Processor) ExtractAudio(videoPath string) (string, error) {
	// gets base name of the video without directory extension
	fileName := filepath.Base(videoPath)
	fileNameWithoutExt := fileName[:len(fileName)-len(filepath.Ext(fileName))]
	audioPath := filepath.Join(p.TempDir, fileNameWithoutExt+".mp3")

	// First check if the video has audio streams
	hasAudioCmd := exec.Command(
		"ffprobe",
		"-v", "error",
		"-select_streams", "a",
		"-show_entries", "stream=codec_type",
		"-of", "default=noprint_wrappers=1:nokey=1",
		videoPath,
	)

	output, _ := hasAudioCmd.Output()
	hasAudio := strings.TrimSpace(string(output)) == "audio"

	var cmd *exec.Cmd

	if hasAudio {
		// If the video has audio, extract it normally
		cmd = exec.Command(
			"ffmpeg",
			"-i", videoPath,
			"-vn",            // No video
			"-acodec", "mp3", // Force mp3 codec
			"-y", // Overwrite output files
			audioPath,
		)
	} else {
		// If no audio, create a silent audio track with same duration as the video
		// First get the duration
		durationCmd := exec.Command(
			"ffprobe",
			"-v", "error",
			"-show_entries", "format=duration",
			"-of", "default=noprint_wrappers=1:nokey=1",
			videoPath,
		)

		durationOutput, err := durationCmd.Output()
		if err != nil {
			return "", fmt.Errorf("failed to get video duration: %w", err)
		}

		duration := strings.TrimSpace(string(durationOutput))

		// Create silent audio
		cmd = exec.Command(
			"ffmpeg",
			"-f", "lavfi", // Use libavfilter
			"-i", "anullsrc=r=44100:cl=stereo", // Generate silent audio
			"-t", duration, // Same duration as video
			"-acodec", "mp3", // MP3 codec
			"-y", // Overwrite output
			audioPath,
		)
	}

	// Run the command
	cmdOutput, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("failed to extract audio: %s - %w", string(cmdOutput), err)
	}

	return audioPath, nil
}

func (p *Processor) CreateThumbnail(videoPath string) (string, error) {
	fileName := filepath.Base(videoPath)
	fileNameWithoutExt := fileName[:len(fileName)-len(filepath.Ext(fileName))]
	thumbnailPath := filepath.Join(p.TempDir, fileNameWithoutExt+".jpg")

	cmd := exec.Command(
		"ffmpeg",
		"-i", videoPath,
		"-ss", "00:00:01",
		"-vframes", "1",
		"-y",
		thumbnailPath,
	)

	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("failed to create thumbnail: %s - %w", string(output), err)
	}

	return thumbnailPath, nil
}

func (p *Processor) GetVideoDuration(videoPath string) (float64, error) {
	cmd := exec.Command(
		"ffprobe",
		"-v", "error",
		"-show_entries", "format=duration",
		"-of", "default=noprint_wrappers=1:nokey=1",
		videoPath,
	)
	output, err := cmd.Output()
	if err != nil {
		return 0, fmt.Errorf("failed to get video duration: %w", err)
	}

	var duration float64
	// Sscanf is for scanning the string and returning the number of items successfully parsed
	// %f is for float
	// &duration (address of duration)is for the variable to store the result
	_, err = fmt.Sscanf(string(output), "%f", &duration)

	if err != nil {
		return 0, fmt.Errorf("failed to parse video duration: %w", err)
	}

	return duration, nil
}
