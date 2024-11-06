package api

import (
	"bytes"
	"context"
	"encoding/json"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/openai/openai-go"
	"github.com/openai/openai-go/shared"
	"interview-buddy/openai_client"
	"interview-buddy/transcription"
	"log"
	"net/http"
	"time"
)

// Create the WebSocket upgrader
var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true // In production, you should check the request origin for security
	},
}

// WebSocket handler function with pause detection
func AudioStreamHandler(c *gin.Context) {
	// Upgrade the HTTP connection to a WebSocket connection
	wsConn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Println("Error upgrading HTTP connection to WebSocket:", err)
		return
	}
	defer wsConn.Close()

	var audioBuffer bytes.Buffer
	silenceDuration := 2 * time.Second // Consider a 2-second silence as a pause
	chunkSize := 1024 * 512            // 512 KB per chunk

	lastAudioReceived := time.Now()
	ticker := time.NewTicker(500 * time.Millisecond) // Check for silence every 500 ms
	defer ticker.Stop()

	go func() {
		for range ticker.C {
			if time.Since(lastAudioReceived) >= silenceDuration && audioBuffer.Len() > 0 {
				// Process the chunk since a pause was detected
				processAudioChunk(&audioBuffer, wsConn)
				audioBuffer.Reset() // Reset the buffer after processing
			}
		}
	}()

	for {
		_, audioData, err := wsConn.ReadMessage()
		if err != nil {
			log.Println("Error reading WebSocket message:", err)
			break
		}

		// Write the received audio data into the buffer
		audioBuffer.Write(audioData)
		lastAudioReceived = time.Now()

		// Process the chunk if the buffer exceeds the chunk size
		if audioBuffer.Len() >= chunkSize {
			processAudioChunk(&audioBuffer, wsConn)
			audioBuffer.Reset() // Reset the buffer after processing
		}
	}
}

func processAudioChunk(audioBuffer *bytes.Buffer, wsConn *websocket.Conn) {
	chunk := audioBuffer.Bytes()

	// Transcribe the audio chunk and check if it contains a question
	transcribedText, isQuestion, err := transcription.TranscribeAndProcessChunk(chunk)
	if err != nil {
		log.Println("Error transcribing audio chunk:", err)
		return
	}

	responseData := make(map[string]string)
	responseData["question"] = transcribedText

	if isQuestion {
		// Send the question to ChatGPT and get the response
		chatResponse, err := getChatGPTResponse(transcribedText)
		if err != nil {
			log.Println("Error getting ChatGPT response:", err)
			return
		}
		responseData["answer"] = chatResponse
	}

	// Send the question and answer back to the frontend via WebSocket
	responseJSON, err := json.Marshal(responseData)
	if err != nil {
		log.Println("Error encoding JSON response:", err)
		return
	}

	if err = wsConn.WriteMessage(websocket.TextMessage, responseJSON); err != nil {
		log.Println("Error writing message to WebSocket:", err)
	}
}
func getChatGPTResponse(prompt string) (string, error) {
	// Example function to send a question to ChatGPT and get the response
	client := openai_client.GetClient()
	ctx := context.Background()
	response, err := client.Completions.New(ctx, openai.CompletionNewParams{
		Model:     openai.F(openai.CompletionNewParamsModelGPT3_5TurboInstruct),
		Prompt:    openai.F[openai.CompletionNewParamsPromptUnion](shared.UnionString(prompt)),
		MaxTokens: openai.F(int64(150)),
	})

	if err != nil {
		return "", err
	}

	return response.Choices[0].Text, nil
}
