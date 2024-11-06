const startButton = document.getElementById('start');
const stopButton = document.getElementById('stop');
const questionsList = document.getElementById('questions-list');
const answersList = document.getElementById('answers-list');

let mediaRecorder;
let socket;

const supportedMimeTypes = [
    'audio/webm;codecs=opus',  // Preferred for browsers that support it
    'audio/webm',              // General webm audio
    'audio/ogg',               // Alternative format
];

// Function to get the first supported mimeType
// function getSupportedMimeType() {
//     for (const mimeType of supportedMimeTypes) {
//         if (MediaRecorder.isTypeSupported(mimeType)) {
//             return mimeType;
//         }
//     }
//     throw new Error('No supported media types found for MediaRecorder.');
// }
const constraints = {
    video: true,
    audio: true
};
startButton.onclick = async () => {
    try {
        // Request to share a tab (with audio)
        const stream = await navigator.mediaDevices.getDisplayMedia(constraints)

        // Get a supported MIME type for MediaRecorder
       // const mimeType = getSupportedMimeType();
      //  console.log("Supported Mime Type is " + mimeType)

        // Open a WebSocket connection to the backend
        socket = new WebSocket("wss://localhost:8080/audio-stream");
        socket.onopen = () => console.log('WebSocket connected');
        socket.onerror = (error) => console.error('WebSocket error:', error);

        // Initialize MediaRecorder with the supported MIME type
        mediaRecorder = new MediaRecorder(stream);

        // Send audio data in chunks via WebSocket
        mediaRecorder.ondataavailable = function(event) {
            if (event.data.size > 0 && socket.readyState === 1) {
                socket.send(event.data);
            }
        };

        // Start recording audio in 1-second chunks
        mediaRecorder.start(1000);

        // Listen for responses from the backend (questions and answers)
        socket.onmessage = function(event) {
            const responseData = JSON.parse(event.data);
            const { question, answer } = responseData;

            if (question) {
                const questionElement = document.createElement('li');
                questionElement.textContent = question;
                questionsList.appendChild(questionElement);
            }

            if (answer) {
                const answerElement = document.createElement('li');
                answerElement.textContent = answer;
                answersList.appendChild(answerElement);
            }
        };

        startButton.disabled = true; // Disable the start button once recording starts
        stopButton.disabled = false; // Enable the stop button
    } catch (error) {
        console.error('Error accessing media:', error);
    }
};

stopButton.onclick = () => {
    if (mediaRecorder) {
        mediaRecorder.stop();
    }
    if (socket) {
        socket.close();
    }
    startButton.disabled = false; // Enable the start button again
    stopButton.disabled = true;   // Disable the stop button
};

// Initially disable the stop button
stopButton.disabled = true;