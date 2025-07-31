//go:build speech
// +build speech

package speech

import (
	"context"
	"fmt"
	"log"
	"math"
	"os"
	"path/filepath"
	"strings"

	"github.com/mutablelogic/go-whisper"
)

// Recognizer wraps speech recognition functionality
type Recognizer struct {
	modelPath string
	enabled   bool
	ctx       whisper.Context
}

// NewRecognizer creates a new speech recognizer
func NewRecognizer(modelPath string) (*Recognizer, error) {
	// Expand ~ to home directory
	if strings.HasPrefix(modelPath, "~/") {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return nil, fmt.Errorf("failed to get home directory: %w", err)
		}
		modelPath = filepath.Join(homeDir, modelPath[2:])
	}

	// Check if model file exists
	if _, err := os.Stat(modelPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("whisper model file not found at %s", modelPath)
	}

	log.Printf("Loading Whisper model from: %s", modelPath)

	// Initialize whisper context from model file
	ctx := whisper.Whisper_init_from_file(modelPath)
	if ctx == nil {
		return nil, fmt.Errorf("failed to initialize whisper context from model: %s", modelPath)
	}

	log.Printf("Whisper model loaded successfully")

	return &Recognizer{
		modelPath: modelPath,
		enabled:   true,
		ctx:       ctx,
	}, nil
}

// Close cleans up the recognizer resources
func (r *Recognizer) Close() {
	if r.ctx != nil {
		whisper.Whisper_free(r.ctx)
		r.ctx = nil
	}
	r.enabled = false
	log.Printf("Whisper recognizer closed")
}

// Transcribe processes audio data and returns the transcribed text
func (r *Recognizer) Transcribe(ctx context.Context, audioData []float32) (string, error) {
	if !r.enabled || r.ctx == nil {
		return "", fmt.Errorf("recognizer not initialized")
	}

	// Check context cancellation
	select {
	case <-ctx.Done():
		return "", ctx.Err()
	default:
	}

	// Analyze audio energy to determine if there's likely speech
	if !hasSignificantAudio(audioData) {
		return "", nil // No speech detected
	}

	energy := calculateRMS(audioData)
	log.Printf("Processing audio - Energy: %f, Duration: %d samples", energy, len(audioData))

	// Set up default parameters for transcription
	params := whisper.Whisper_full_default_params(whisper.WHISPER_SAMPLING_GREEDY)
	params.SetLanguage("en")
	params.SetPrintProgress(false)
	params.SetPrintRealtime(false)
	params.SetPrintTimestamps(false)

	// Process audio with Whisper
	if err := whisper.Whisper_full(r.ctx, params, audioData, len(audioData)); err != nil {
		return "", fmt.Errorf("whisper processing failed: %w", err)
	}

	// Get the number of segments
	nSegments := whisper.Whisper_full_n_segments(r.ctx)
	if nSegments == 0 {
		return "", nil // No segments found
	}

	// Collect all segment text
	var results []string
	for i := 0; i < nSegments; i++ {
		segment := whisper.Whisper_full_get_segment_text(r.ctx, i)
		text := strings.TrimSpace(segment)
		if text != "" {
			results = append(results, text)
		}
	}

	// Join all segments
	fullText := strings.Join(results, " ")
	fullText = strings.TrimSpace(fullText)

	if fullText != "" {
		log.Printf("Whisper transcribed: %s", fullText)
	}

	return fullText, nil
}

// hasSignificantAudio checks if the audio data contains significant signal
func hasSignificantAudio(audioData []float32) bool {
	if len(audioData) == 0 {
		return false
	}

	// Calculate RMS energy
	energy := calculateRMS(audioData)

	// Threshold for considering audio as potential speech
	return energy > 0.01 // Adjust threshold as needed
}

// calculateRMS calculates the Root Mean Square energy of audio data
func calculateRMS(audioData []float32) float32 {
	if len(audioData) == 0 {
		return 0
	}

	var sum float64
	for _, sample := range audioData {
		sum += float64(sample * sample)
	}

	return float32(math.Sqrt(sum / float64(len(audioData))))
}

// IsKeywordDetected checks if the transcribed text contains the keyword
func IsKeywordDetected(text, keyword string) bool {
	if keyword == "" {
		return false
	}

	// Convert to lowercase for case-insensitive comparison
	lowerText := strings.ToLower(strings.TrimSpace(text))
	lowerKeyword := strings.ToLower(strings.TrimSpace(keyword))

	return strings.Contains(lowerText, lowerKeyword)
}

// StripKeyword removes the keyword from the transcribed text
func StripKeyword(text, keyword string) string {
	if keyword == "" {
		return text
	}

	lowerText := strings.ToLower(text)
	lowerKeyword := strings.ToLower(keyword)

	// Find the keyword position
	index := strings.Index(lowerText, lowerKeyword)
	if index == -1 {
		return text
	}

	// Remove the keyword and clean up whitespace
	before := strings.TrimSpace(text[:index])
	after := strings.TrimSpace(text[index+len(keyword):])

	if before == "" {
		return after
	}
	if after == "" {
		return before
	}

	return before + " " + after
}
