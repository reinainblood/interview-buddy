// popup.js
document.addEventListener('DOMContentLoaded', function() {
    const startBtn = document.getElementById('startBtn');
    const stopBtn = document.getElementById('stopBtn');

    if (startBtn) {
        startBtn.addEventListener('click', function() {
            console.log('Start button clicked');
            chrome.runtime.sendMessage({action: 'start'});
        });
    }

    if (stopBtn) {
        stopBtn.addEventListener('click', function() {
            console.log('Stop button clicked');
            chrome.runtime.sendMessage({action: 'stop'});
        });
    }
});