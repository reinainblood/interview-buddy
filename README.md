###  ⚠️ **WIP** this is a work in progress, not production code  ⚠️ 


## Interview Buddy
#### an OpenAI enabled voice transcription Go application that will eventually run as a browser extension (or in a popup window, haven't totally decided)

#### To run:
    1. Create a ```.env``` file and add ```OPENAI_API_KEY=<your openai key>```
    2. This has only been run locally as it's a very new project ```go build interview-buddy``` with the current configuration will stand up a dev Gin API server
    3. You can try using curl to send audio to the endpoints in ```main.go``` or use a browser at localhost:8080