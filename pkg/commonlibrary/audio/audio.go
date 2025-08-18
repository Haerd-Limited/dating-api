package audio

import (
	"context"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

func GetAudioDuration(file *multipart.File) (time.Duration, error) {
	// Step 1: Write multipart file to a temp file
	tmpFile, err := os.CreateTemp("", "voicenote-*.m4a")
	if err != nil {
		return 0, fmt.Errorf("failed to create temp file: %w", err)
	}
	defer func(name string) {
		err := os.Remove(name)
		if err != nil {
			fmt.Printf("failed to remove temp file: %s\n", name)
		}
	}(tmpFile.Name()) // Clean up

	if _, err := io.Copy(tmpFile, *file); err != nil {
		return 0, fmt.Errorf("failed to write to temp file: %w", err)
	}

	if err := tmpFile.Close(); err != nil {
		return 0, fmt.Errorf("failed to close temp file: %w", err)
	}

	// Step 2: Use ffprobe to get duration
	cmd := exec.CommandContext(context.Background(), "ffprobe",
		"-v", "error",
		"-show_entries", "format=duration",
		"-of", "default=noprint_wrappers=1:nokey=1",
		tmpFile.Name(),
	)

	output, err := cmd.Output()
	if err != nil {
		return 0, fmt.Errorf("failed to run ffprobe: %w", err)
	}

	// Step 3: Parse duration
	secondsStr := strings.TrimSpace(string(output))

	seconds, err := strconv.ParseFloat(secondsStr, 64)
	if err != nil {
		return 0, fmt.Errorf("failed to parse duration: %w", err)
	}

	return time.Duration(seconds * float64(time.Second)), nil
}

// DetectContentType reads the first 512 bytes to guess the MIME type
func DetectContentType(file multipart.File) (string, error) {
	// Read the first 512 bytes
	buf := make([]byte, 512)

	_, err := file.Read(buf)
	if err != nil && err != io.EOF {
		return "", err
	}

	// Reset file pointer back to beginning after read
	if _, err := file.Seek(0, io.SeekStart); err != nil {
		return "", err
	}

	contentType := http.DetectContentType(buf)

	return contentType, nil
}
