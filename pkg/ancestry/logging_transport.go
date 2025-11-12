package ancestry

import (
	"fmt"
	"net/http"
	"net/http/httputil"
	"os"
	"time"
)

// loggingTransport is an http.RoundTripper that logs requests and responses.
type loggingTransport struct {
	transport http.RoundTripper
	logFile   *os.File
}

// newLoggingTransport creates a new loggingTransport.
func newLoggingTransport(transport http.RoundTripper) (*loggingTransport, error) {
	logFile, err := os.OpenFile("http_log.txt", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return nil, fmt.Errorf("failed to open log file: %w", err)
	}

	return &loggingTransport{
		transport: transport,
		logFile:   logFile,
	}, nil
}

// RoundTrip executes a single HTTP transaction, logging the request and response.
func (t *loggingTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	// Log the request
	reqDump, err := httputil.DumpRequestOut(req, true)
	if err != nil {
		t.log(fmt.Sprintf("--- Failed to dump request: %v ---\n", err))
	} else {
		reqLogEntry := fmt.Sprintf("\n=== REQUEST: %s %s ===\nTime: %s\n%s\n",
			req.Method, req.URL.String(), time.Now().Format(time.RFC3339), string(reqDump))
		t.log(reqLogEntry)
	}

	// Perform the request
	resp, err := t.transport.RoundTrip(req)
	if err != nil {
		t.log(fmt.Sprintf("--- Request failed: %v ---\n\n", err))
		return nil, err
	}

	// Log the response
	respDump, err := httputil.DumpResponse(resp, true)
	if err != nil {
		// Don't fail the request if we can't dump the response
		t.log(fmt.Sprintf("--- Failed to dump response: %v ---\n\n", err))
		return resp, nil
	}

	respLogEntry := fmt.Sprintf("=== RESPONSE: %d ===\nTime: %s\n%s\n",
		resp.StatusCode, time.Now().Format(time.RFC3339), string(respDump))
	t.log(respLogEntry)

	// Add separator for readability
	t.log("========================================\n\n")

	return resp, nil
}

// Close closes the log file.
func (t *loggingTransport) Close() error {
	if t.logFile != nil {
		return t.logFile.Close()
	}
	return nil
}

func (t *loggingTransport) log(msg string) {
	_, err := t.logFile.WriteString(msg)
	if err != nil {
		// Handle the error, e.g., log it to stderr
		fmt.Fprintf(os.Stderr, "--- Failed to write to log file: %v ---\n", err)
	}
}
