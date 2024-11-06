// background.js
// Service worker setup
const audioContexts = new Map();
let isRecording = false;

// Listen for messages from popup
chrome.runtime.onMessage.addListener((message, sender, sendResponse) => {
    console.log('Received message:', message);

    // Must return true if we want to send a response asynchronously
    if (message.action === 'getState') {
        sendResponse({ isRecording });
        return true;
    }

    if (message.action === 'start') {
        startCapture().catch(console.error);
    } else if (message.action === 'stop') {
        stopCapture();
    }
});

async function startCapture() {
    try {
        const [tab] = await chrome.tabs.query({ active: true, currentWindow: true });
        if (!tab) {
            throw new Error('No active tab found');
        }

        // Capture tab audio
        const stream = await new Promise((resolve, reject) => {
            chrome.tabCapture.capture({
                audio: true,
                video: false,
                audioConstraints: {
                    mandatory: {
                        echoCancellation: false,
                        noiseSuppression: false,
                        autoGainControl: false
                    }
                }
            }, (stream) => {
                if (chrome.runtime.lastError) {
                    reject(chrome.runtime.lastError);
                    return;
                }
                resolve(stream);
            });
        });

        if (!stream) {
            throw new Error('Failed to capture tab audio');
        }

        isRecording = true;

        // Create audio processing setup
        const audioContext = new AudioContext();
        const source = audioContext.createMediaStreamSource(stream);
        const processor = audioContext.createScriptProcessor(4096, 1, 1);

        // Store context for cleanup
        audioContexts.set(tab.id, {
            context: audioContext,
            source,
            processor,
            stream
        });

        // Set up WebSocket
        const ws = new WebSocket('ws://localhost:8080/audio');

        ws.onopen = () => {
            console.log('WebSocket connected');
        };

        ws.onerror = (error) => {
            console.error('WebSocket error:', error);
        };

        // Connect audio nodes
        source.connect(processor);
        processor.connect(audioContext.destination);

        // Process audio
        processor.onaudioprocess = (e) => {
            if (!isRecording) return;

            const inputData = e.inputBuffer.getChannelData(0);
            const pcmData = new Int16Array(inputData.length);

            // Convert to 16-bit PCM
            for (let i = 0; i < inputData.length; i++) {
                const s = Math.max(-1, Math.min(1, inputData[i]));
                pcmData[i] = s < 0 ? s * 0x8000 : s * 0x7FFF;
            }

            if (ws?.readyState === WebSocket.OPEN) {
                ws.send(pcmData.buffer);
            }
        };

        // Store WebSocket for cleanup
        audioContexts.get(tab.id).ws = ws;

    } catch (error) {
        console.error('Error in startCapture:', error);
        isRecording = false;
        throw error;
    }
}

function stopCapture() {
    isRecording = false;

    // Cleanup all tabs
    for (const [tabId, context] of audioContexts.entries()) {
        if (context.stream) {
            context.stream.getTracks().forEach(track => track.stop());
        }
        if (context.ws) {
            context.ws.close();
        }
        if (context.source) {
            context.source.disconnect();
        }
        if (context.processor) {
            context.processor.disconnect();
        }
        if (context.context) {
            context.context.close();
        }
    }

    audioContexts.clear();
}

// Handle service worker lifecycle
self.addEventListener('activate', (event) => {
    console.log('Service Worker activated');
});

self.addEventListener('install', (event) => {
    console.log('Service Worker installed');
});