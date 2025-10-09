package httpserver

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"sync"
	"time"
)

// JSONLogEntry represents a single HTTP request log entry in JSON format
type JSONLogEntry struct {
	Timestamp    string            `json:"timestamp"`
	RemoteAddr   string            `json:"remote_addr"`
	Method       string            `json:"method"`
	URL          string            `json:"url"`
	Proto        string            `json:"proto"`
	StatusCode   int               `json:"status_code"`
	Size         int               `json:"size"`
	UserAgent    string            `json:"user_agent,omitempty"`
	Headers      map[string]string `json:"headers,omitempty"`
	RequestBody  string            `json:"request_body,omitempty"`
	ResponseBody string            `json:"response_body,omitempty"`
}

// JSONLogger handles writing JSON log entries to a file
type JSONLogger struct {
	file   *os.File
	mutex  sync.Mutex
	enable bool
}

// NewJSONLogger creates a new JSON logger instance
func NewJSONLogger(filePath string) (*JSONLogger, error) {
	if filePath == "" {
		return &JSONLogger{enable: false}, nil
	}

	file, err := os.OpenFile(filePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return nil, fmt.Errorf("failed to open JSON log file: %w", err)
	}

	return &JSONLogger{
		file:   file,
		enable: true,
	}, nil
}

// LogRequest logs an HTTP request to the JSON file
func (jl *JSONLogger) LogRequest(r *http.Request, statusCode, size int, userAgent string, headers map[string]string, requestBody, responseBody string) error {
	if !jl.enable {
		return nil
	}

	entry := JSONLogEntry{
		Timestamp:    time.Now().Format(time.RFC3339),
		RemoteAddr:   r.RemoteAddr,
		Method:       r.Method,
		URL:          r.URL.String(),
		Proto:        r.Proto,
		StatusCode:   statusCode,
		Size:         size,
		UserAgent:    userAgent,
		Headers:      headers,
		RequestBody:  requestBody,
		ResponseBody: responseBody,
	}

	jl.mutex.Lock()
	defer jl.mutex.Unlock()

	jsonData, err := json.Marshal(entry)
	if err != nil {
		return fmt.Errorf("failed to marshal JSON log entry: %w", err)
	}

	_, err = jl.file.Write(append(jsonData, '\n'))
	if err != nil {
		return fmt.Errorf("failed to write JSON log entry: %w", err)
	}

	return nil
}

// Close closes the JSON log file
func (jl *JSONLogger) Close() error {
	if jl.file != nil {
		return jl.file.Close()
	}
	return nil
}
