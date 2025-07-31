//go:build speech
// +build speech

package speech

import (
	"context"
	"fmt"
	"log"
	"math"
	"time"

	"github.com/gordonklaus/portaudio"
	"github.com/pprunty/magikarp/internal/config"
)

const (
	// Audio processing constants
	frameSize = 1024
	channels  = 1
)

// audioBuffer holds a circular buffer for audio data
type audioBuffer struct {
	data     []float32
	size     int
	writePos int
	readPos  int
}

// newAudioBuffer creates a new circular audio buffer
func newAudioBuffer(size int) *audioBuffer {
	return &audioBuffer{
		data: make([]float32, size),
		size: size,
	}
}

// write adds samples to the buffer
func (b *audioBuffer) write(samples []float32) {
	for _, sample := range samples {
		b.data[b.writePos] = sample
		b.writePos = (b.writePos + 1) % b.size
	}
}

// read extracts samples from the buffer
func (b *audioBuffer) read(samples []float32) int {
	count := 0
	for i := range samples {
		if b.readPos == b.writePos {
			break // Buffer empty
		}
		samples[i] = b.data[b.readPos]
		b.readPos = (b.readPos + 1) % b.size
		count++
	}
	return count
}

// available returns the number of samples available to read
func (b *audioBuffer) available() int {
	if b.writePos >= b.readPos {
		return b.writePos - b.readPos
	}
	return b.size - b.readPos + b.writePos
}

// Listen starts listening for speech input and sends transcribed text to the output channel
func Listen(ctx context.Context, out chan<- string, cfg config.Speech) error {
	// Initialize PortAudio
	log.Printf("Initializing PortAudio...")
	if err := portaudio.Initialize(); err != nil {
		return fmt.Errorf("failed to initialize PortAudio: %w", err)
	}
	defer portaudio.Terminate()
	log.Printf("PortAudio initialized successfully")

	// Create recognizer
	recognizer, err := NewRecognizer(cfg.ModelPath)
	if err != nil {
		return fmt.Errorf("failed to create recognizer: %w", err)
	}
	defer recognizer.Close()

	// Set up audio buffer - store 3 seconds of audio for processing
	bufferSize := cfg.SampleRate * 3
	buffer := newAudioBuffer(bufferSize)

	// Audio input buffer
	audioInput := make([]float32, frameSize)

	// Open default input stream
	log.Printf("Opening audio stream with sample rate %d...", cfg.SampleRate)
	stream, err := portaudio.OpenDefaultStream(
		1,                       // input channels
		0,                       // output channels
		float64(cfg.SampleRate), // sample rate
		frameSize,               // frames per buffer
		&audioInput,             // input buffer (pointer to slice)
	)
	if err != nil {
		return fmt.Errorf("failed to open audio stream: %w", err)
	}
	defer stream.Close()
	log.Printf("Audio stream opened successfully")

	// Start the stream
	if err := stream.Start(); err != nil {
		return fmt.Errorf("failed to start audio stream: %w", err)
	}
	defer stream.Stop()

	log.Printf("Speech recognition started. Say your message and end with '%s'", cfg.Keyword)
	log.Printf("Sample rate: %d, Buffer size: %d frames", cfg.SampleRate, frameSize)

	// Read audio continuously to prevent overflow
	go func() {
		ticker := time.NewTicker(10 * time.Millisecond) // Read very frequently to prevent overflow
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				// Read audio data from the stream to prevent overflow
				if err := stream.Read(); err != nil {
					if ctx.Err() != nil {
						return // Context cancelled, normal shutdown
					}
					log.Printf("Error reading audio: %v", err)
					continue
				}

				// Copy input data to avoid concurrent access issues
				inputCopy := make([]float32, len(audioInput))
				copy(inputCopy, audioInput)

				// Add samples to buffer
				buffer.write(inputCopy)
			}
		}
	}()

	// Process audio for speech recognition less frequently
	processTicker := time.NewTicker(2 * time.Second) // Process every 2 seconds to avoid spam
	defer processTicker.Stop()

	var lastTranscriptionTime time.Time
	const transcriptionCooldown = 3 * time.Second // Minimum time between transcriptions

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-processTicker.C:
			// Check cooldown period
			if time.Since(lastTranscriptionTime) < transcriptionCooldown {
				log.Printf("Skipping transcription - in cooldown period")
				continue
			}

			// Check if we have enough data to process (3 seconds worth for better accuracy)
			requiredSamples := cfg.SampleRate * 3
			if buffer.available() >= requiredSamples {
				// Extract samples for processing
				samples := make([]float32, requiredSamples)
				count := buffer.read(samples)
				if count > 0 {
					// Apply voice activity detection with appropriate threshold
					if detectVoiceActivity(samples[:count], 0.015) {
						// Normalize audio for better Whisper processing
						normalizeAudio(samples[:count])

						log.Printf("Processing %d samples for Whisper transcription...", count)
						// Transcribe audio synchronously to prevent multiple concurrent transcriptions
						text, err := recognizer.Transcribe(ctx, samples[:count])
						if err != nil {
							log.Printf("Whisper transcription error: %v", err)
							continue
						}

						// Check if we got meaningful text
						if text != "" && len(text) > 1 {
							log.Printf("Whisper transcribed: %s", text)

							// Check for keyword
							if IsKeywordDetected(text, cfg.Keyword) {
								// Strip keyword and send result
								finalText := StripKeyword(text, cfg.Keyword)
								if finalText != "" {
									select {
									case out <- finalText:
										log.Printf("Sent speech result: %s", finalText)
										lastTranscriptionTime = time.Now() // Update last transcription time
									case <-ctx.Done():
										return ctx.Err()
									}
								}
							} else {
								log.Printf("No keyword '%s' detected in: %s", cfg.Keyword, text)
							}
						}
					} else {
						log.Printf("No voice activity detected (RMS: %f)", calculateRMS(samples[:count]))
					}
				}
			}
		}
	}
}

// normalizeAudio applies basic audio normalization
func normalizeAudio(samples []float32) {
	if len(samples) == 0 {
		return
	}

	// Find the maximum absolute value
	maxVal := float32(0)
	for _, sample := range samples {
		abs := float32(math.Abs(float64(sample)))
		if abs > maxVal {
			maxVal = abs
		}
	}

	// Avoid division by zero
	if maxVal == 0 {
		return
	}

	// Normalize to prevent clipping while maintaining dynamic range
	normalizationFactor := float32(0.8) / maxVal
	if normalizationFactor < 1.0 {
		for i := range samples {
			samples[i] *= normalizationFactor
		}
	}
}

// detectVoiceActivity performs simple voice activity detection
func detectVoiceActivity(samples []float32, threshold float32) bool {
	if len(samples) == 0 {
		return false
	}

	rms := calculateRMS(samples)
	return rms > threshold
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
