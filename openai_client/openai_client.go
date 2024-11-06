package openai_client

import (
	"github.com/openai/openai-go"
	"github.com/openai/openai-go/option"
	"log"
	"os"
)

var client *openai.Client

// InitializeClient initializes the OpenAI client with the API key.
// It ensures that only one instance of the client is created.
func InitializeClient() {
	if client != nil {
		return // Client already initialized, no need to create again
	}

	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		log.Fatal("OpenAI API key not set in environment")
	}

	client = openai.NewClient(option.WithAPIKey(apiKey)) // defaults to os.LookupEnv("OPENAI_API_KEY")
	log.Println("OpenAI client initialized successfully")
}

// GetClient returns the initialized OpenAI client.
func GetClient() *openai.Client {
	if client == nil {
		log.Fatal("OpenAI client has not been initialized")
	}
	return client
}
