package main

import (
	"fmt"
	"net/http"
	"time"
)

// RetryConfig configures retry behavior for HTTP requests
type RetryConfig struct {
	MaxRetries     int
	InitialBackoff time.Duration
	MaxBackoff     time.Duration
	Multiplier     float64
}

// DefaultRetryConfig returns sensible defaults for retry configuration
func DefaultRetryConfig() RetryConfig {
	return RetryConfig{
		MaxRetries:     3,
		InitialBackoff: 1 * time.Second,
		MaxBackoff:     30 * time.Second,
		Multiplier:     2.0,
	}
}

// HTTPClientWrapper wraps an HTTP client with retry logic and error handling
type HTTPClientWrapper struct {
	client      *http.Client
	retryConfig RetryConfig
	logger      Logger
}

// NewHTTPClientWrapper creates a new HTTP client wrapper
func NewHTTPClientWrapper(client *http.Client, retryConfig RetryConfig, logger Logger) *HTTPClientWrapper {
	if client == nil {
		client = &http.Client{
			Timeout: 30 * time.Second,
		}
	}
	if logger == nil {
		logger = NewLogger("HTTPClient")
	}
	
	return &HTTPClientWrapper{
		client:      client,
		retryConfig: retryConfig,
		logger:      logger,
	}
}

// Do executes an HTTP request with retry logic
func (w *HTTPClientWrapper) Do(req *http.Request) (*http.Response, error) {
	var lastErr error
	backoff := w.retryConfig.InitialBackoff
	
	for attempt := 0; attempt <= w.retryConfig.MaxRetries; attempt++ {
		if attempt > 0 {
			w.logger.Warn("Retrying request (attempt %d/%d) after %v",
				attempt, w.retryConfig.MaxRetries, backoff)
			time.Sleep(backoff)
			
			// Exponential backoff
			backoff = time.Duration(float64(backoff) * w.retryConfig.Multiplier)
			if backoff > w.retryConfig.MaxBackoff {
				backoff = w.retryConfig.MaxBackoff
			}
		}
		
		resp, err := w.client.Do(req)
		if err != nil {
			lastErr = fmt.Errorf("request failed: %w", err)
			w.logger.Warn("Request attempt %d failed: %v", attempt+1, err)
			continue
		}
		
		// Check if status code indicates we should retry
		if w.shouldRetry(resp.StatusCode) {
			resp.Body.Close()
			lastErr = fmt.Errorf("server returned status %d", resp.StatusCode)
			w.logger.Warn("Request attempt %d got status %d", attempt+1, resp.StatusCode)
			continue
		}
		
		// Success
		return resp, nil
	}
	
	return nil, fmt.Errorf("request failed after %d attempts: %w",
		w.retryConfig.MaxRetries+1, lastErr)
}

// shouldRetry determines if a status code warrants a retry
func (w *HTTPClientWrapper) shouldRetry(statusCode int) bool {
	// Retry on server errors (5xx) and rate limiting (429)
	return statusCode >= 500 || statusCode == 429
}

// Get performs a GET request with retry logic
func (w *HTTPClientWrapper) Get(url string) (*http.Response, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create GET request: %w", err)
	}
	
	return w.Do(req)
}
