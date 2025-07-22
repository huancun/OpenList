package middlewares

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestFilteredLogger(t *testing.T) {
	// Set Gin to test mode
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		method         string
		path           string
		shouldLog      bool
		description    string
	}{
		{
			name:        "Health check ping should be filtered",
			method:      "GET",
			path:        "/ping",
			shouldLog:   false,
			description: "Health check requests should not be logged",
		},
		{
			name:        "HEAD request should be filtered",
			method:      "HEAD",
			path:        "/any/path",
			shouldLog:   false,
			description: "HEAD requests should not be logged",
		},
		{
			name:        "PROPFIND request should be filtered",
			method:      "PROPFIND",
			path:        "/dav/some/path",
			shouldLog:   false,
			description: "PROPFIND requests should not be logged",
		},
		{
			name:        "WebDAV path should be filtered",
			method:      "GET",
			path:        "/dav/some/webdav/path",
			shouldLog:   false,
			description: "WebDAV paths should not be logged",
		},
		{
			name:        "Normal GET request should be logged",
			method:      "GET",
			path:        "/api/files",
			shouldLog:   true,
			description: "Normal API requests should be logged",
		},
		{
			name:        "POST request should be logged",
			method:      "POST",
			path:        "/api/upload",
			shouldLog:   true,
			description: "POST requests should be logged",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a buffer to capture log output
			var buf bytes.Buffer

			// Create a new Gin engine with our filtered logger
			r := gin.New()
			config := FilteredLoggerConfig{
				SkipPaths: []string{"/ping"},
				SkipMethods: []string{"HEAD"},
				SkipPathPrefixes: []string{"/dav/"},
				Output: &buf,
			}
			r.Use(FilteredLoggerWithConfig(config))

			// Add a simple handler
			r.Any("/*path", func(c *gin.Context) {
				c.String(http.StatusOK, "OK")
			})

			// Create a test request
			req := httptest.NewRequest(tt.method, tt.path, nil)
			w := httptest.NewRecorder()

			// Perform the request
			r.ServeHTTP(w, req)

			// Check if logging occurred as expected
			logOutput := buf.String()
			if tt.shouldLog {
				assert.NotEmpty(t, logOutput, tt.description)
				assert.Contains(t, logOutput, tt.method, "Log should contain the HTTP method")
				assert.Contains(t, logOutput, tt.path, "Log should contain the request path")
			} else {
				assert.Empty(t, logOutput, tt.description)
			}
		})
	}
}

func TestShouldSkipLogging(t *testing.T) {
	config := FilteredLoggerConfig{
		SkipPaths:        []string{"/ping", "/health"},
		SkipMethods:      []string{"HEAD", "OPTIONS"},
		SkipPathPrefixes: []string{"/dav/", "/static/"},
	}

	tests := []struct {
		path     string
		method   string
		expected bool
		reason   string
	}{
		{"/ping", "GET", true, "ping path should be skipped"},
		{"/health", "POST", true, "health path should be skipped"},
		{"/api/test", "HEAD", true, "HEAD method should be skipped"},
		{"/api/test", "OPTIONS", true, "OPTIONS method should be skipped"},
		{"/dav/some/path", "GET", true, "dav prefix should be skipped"},
		{"/static/css/style.css", "GET", true, "static prefix should be skipped"},
		{"/any/path", "PROPFIND", true, "PROPFIND method should be skipped"},
		{"/api/files", "GET", false, "normal API request should not be skipped"},
		{"/upload", "POST", false, "normal POST request should not be skipped"},
	}

	for _, tt := range tests {
		t.Run(tt.reason, func(t *testing.T) {
			result := shouldSkipLogging(tt.path, tt.method, config)
			assert.Equal(t, tt.expected, result, tt.reason)
		})
	}
}