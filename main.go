package main

import (
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"interview-buddy/api"
	"interview-buddy/openai_client"
	"log"
)

func main() {
	// Load the .env file
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}
	r := gin.Default()
	// Initialize the OpenAI client once during application startup
	openai_client.InitializeClient()

	// Serve static files from the /static directory
	r.Static("/static", "./static")

	// Serve the index.html from the /frontend directory
	r.StaticFile("/", "./frontend/index.html")

	// Handle WebSocket for audio stream
	r.GET("/audio-stream", api.AudioStreamHandler)
	// Start HTTPS server using self-signed certificate
	log.Println("Starting server on https://localhost:8080")
	connect_err := r.RunTLS(":8080", "certs/cert.pem", "certs/key.pem")
	if connect_err != nil {
		log.Fatal("Failed to start HTTPS server:", err)
	}

}
