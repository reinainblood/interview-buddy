package transcription

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/go-audio/audio"
	"github.com/go-audio/wav"
	"github.com/openai/openai-go"
	"interview-buddy/openai_client"
	"io"
)

// TranscribeAndProcessChunk processes a raw audio chunk, converts it to a WAV format, and transcribes it using the OpenAI Whisper API.
func TranscribeAndProcessChunk(audioChunk []byte) (string, bool, error) {
	// Specify a custom directory to save the .wav files (e.g., "./audio_files/")
	outputDir := "./audio_files/"
	if _, err := os.Stat(outputDir); os.IsNotExist(err) {
		err := os.Mkdir(outputDir, 0755) // Create directory if it doesn't exist
		if err != nil {
			return "", false, fmt.Errorf("failed to create output directory: %w", err)
		}
	}

	// Create a named .wav file in the custom directory
	tempFilePath := filepath.Join(outputDir, fmt.Sprintf("audio_chunk_%d.wav", time.Now().UnixNano()))
	tempFile, err := os.Create(tempFilePath)
	if err != nil {
		return "", false, fmt.Errorf("failed to create temp file: %w", err)
	}

	// Convert the raw audio chunk into a valid WAV format
	err = convertToWav(tempFile, audioChunk)
	if err != nil {
		return "", false, fmt.Errorf("failed to convert to WAV: %w", err)
	}

	// Ensure the file pointer is at the start
	tempFile.Seek(0, io.SeekStart)

	// Transcribe the WAV file using OpenAI Whisper API
	client := openai_client.GetClient()
	ctx := context.Background() // Create a context for the API request

	// Send the request to OpenAI's transcription API
	transcription, err := client.Audio.Transcriptions.New(ctx, openai.AudioTranscriptionNewParams{
		File:     openai.F[io.Reader](tempFile),       // Pass the temp file as io.Reader
		Model:    openai.F(openai.AudioModelWhisper1), // Specify the Whisper model
		Language: openai.F("en"),                      // Specify language (English)
	})
	if err != nil {
		return "", false, fmt.Errorf("transcription error: %w", err)
	}

	// Don't remove the temp file so you can inspect it manually
	fmt.Printf("Saved audio file: %s\n", tempFilePath) // Log the saved file path

	// Check if the transcription contains a question
	isQuestion := containsQuestion(transcription.Text)
	return transcription.Text, isQuestion, nil
}

// convertToWav converts the raw audio data to a valid WAV file format.
func convertToWav(file *os.File, audioData []byte) error {
	// Create a WAV encoder with the target file
	encoder := wav.NewEncoder(file, 16000, 16, 1, 1) // 16kHz, 16-bit, mono

	// Ensure the length of audioData is even to avoid index out of bounds
	if len(audioData)%2 != 0 {
		audioData = audioData[:len(audioData)-1] // Discard the last byte if the length is odd
	}

	// Prepare audio buffer from raw audio data (assuming 16-bit PCM)
	buf := &audio.IntBuffer{
		Data:           make([]int, len(audioData)/2),                    // Each 16-bit sample is 2 bytes
		Format:         &audio.Format{SampleRate: 16000, NumChannels: 1}, // 16kHz, Mono
		SourceBitDepth: 16,
	}

	// Convert byte slice to int slice (16-bit samples)
	for i := 0; i < len(audioData); i += 2 {
		sample := int(audioData[i]) | int(audioData[i+1])<<8 // Combine two bytes into one sample
		buf.Data[i/2] = sample
	}

	// Write the buffer to the WAV file
	if err := encoder.Write(buf); err != nil {
		return fmt.Errorf("failed to write WAV data: %w", err)
	}

	// Close the encoder to finalize the WAV file
	if err := encoder.Close(); err != nil {
		return fmt.Errorf("failed to close WAV encoder: %w", err)
	}

	return nil
}

// containsQuestion checks if the transcription contains a question.
func containsQuestion(text string) bool {
	return strings.Contains(text, "?")
}
