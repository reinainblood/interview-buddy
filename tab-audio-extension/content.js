console.log('Sentiment Sidecar content script loaded');

// Listen for messages from the background script
chrome.runtime.onMessage.addListener((message, sender, sendResponse) => {
    if (message.action === 'getPageInfo') {
        // You can add logic here to extract information from the page if needed
        sendResponse({ title: document.title, url: window.location.href });
    }
});