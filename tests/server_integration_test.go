package ddd_tests

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/paulvitic/ddd-go"
)

// TestServerLifecycle tests the complete server lifecycle with proper setup and teardown
func TestServerLifecycle(t *testing.T) {
	// Test configuration
	const (
		testHost = "localhost"
		testPort = 8083
		baseURL  = "http://localhost:8083"
	)

	// Setup: Create context and server
	testContext := ddd.NewContext("api").
		WithResources(
			ddd.Resource(NewTestEndpoint, ddd.Request),
		)

	server := ddd.NewServer(testHost, testPort).
		WithContexts(testContext)

	// Create context for server lifecycle management
	serverCtx, cancelServer := context.WithCancel(context.Background())
	defer cancelServer()

	// Start server
	serverErrors := make(chan error, 1)
	go func() {
		if err := server.StartWithContext(serverCtx); err != nil {
			serverErrors <- err
		}
	}()

	// Wait for server to be ready
	if !waitForServer(baseURL, 5*time.Second) {
		t.Fatal("Server failed to start within timeout")
	}

	// Check for early server errors
	select {
	case err := <-serverErrors:
		t.Fatalf("Server failed to start: %v", err)
	default:
		// Server is running
	}

	// Run all endpoint tests
	t.Run("ServerEndpoints", func(t *testing.T) {
		testAllEndpoints(t, baseURL)
	})

	// Test concurrent requests
	t.Run("ConcurrentRequests", func(t *testing.T) {
		testConcurrentRequests(t, baseURL)
	})

	// Test server health during load
	t.Run("LoadTest", func(t *testing.T) {
		testServerLoad(t, baseURL)
	})

	// Cleanup: Stop server gracefully
	t.Log("Stopping server...")
	cancelServer()

	// Wait for server to shut down
	select {
	case <-serverErrors:
		t.Log("Server shut down successfully")
	case <-time.After(10 * time.Second):
		t.Error("Server shutdown timed out")
	}

	// Verify server is actually stopped
	if serverStillRunning(baseURL) {
		t.Error("Server is still running after shutdown")
	}
}

// waitForServer waits for the server to be ready
func waitForServer(baseURL string, timeout time.Duration) bool {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if resp, err := http.Get(baseURL + "/"); err == nil {
			resp.Body.Close()
			if resp.StatusCode == http.StatusOK {
				return true
			}
		}
		time.Sleep(50 * time.Millisecond)
	}
	return false
}

// serverStillRunning checks if the server is still responding
func serverStillRunning(baseURL string) bool {
	client := &http.Client{Timeout: 1 * time.Second}
	if resp, err := client.Get(baseURL + "/"); err == nil {
		resp.Body.Close()
		return true
	}
	return false
}

// testAllEndpoints tests all the endpoints
func testAllEndpoints(t *testing.T, baseURL string) {
	tests := []struct {
		name           string
		method         string
		path           string
		expectedStatus int
		expectedMethod string
	}{
		{
			name:           "Health Check",
			method:         "GET",
			path:           "/",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "GET Endpoint",
			method:         "GET",
			path:           "/api/test",
			expectedStatus: http.StatusOK,
			expectedMethod: "GET",
		},
		{
			name:           "POST Endpoint",
			method:         "POST",
			path:           "/api/test",
			expectedStatus: http.StatusCreated,
			expectedMethod: "POST",
		},
		{
			name:           "PUT Endpoint",
			method:         "PUT",
			path:           "/api/test",
			expectedStatus: http.StatusOK,
			expectedMethod: "PUT",
		},
		{
			name:           "DELETE Endpoint",
			method:         "DELETE",
			path:           "/api/test",
			expectedStatus: http.StatusNoContent,
			expectedMethod: "DELETE",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			req, err := http.NewRequest(test.method, baseURL+test.path, nil)
			if err != nil {
				t.Fatalf("Failed to create request: %v", err)
			}

			client := &http.Client{Timeout: 5 * time.Second}
			resp, err := client.Do(req)
			if err != nil {
				t.Fatalf("Request failed: %v", err)
			}
			defer resp.Body.Close()

			if resp.StatusCode != test.expectedStatus {
				t.Errorf("Expected status %d, got %d", test.expectedStatus, resp.StatusCode)
			}

			// For API endpoints, check JSON response
			if test.expectedMethod != "" && resp.StatusCode != http.StatusNoContent {
				var response map[string]interface{}
				if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
					t.Fatalf("Failed to decode JSON: %v", err)
				}

				if method, ok := response["method"]; !ok || method != test.expectedMethod {
					t.Errorf("Expected method %s, got %v", test.expectedMethod, method)
				}
			}
		})
	}
}

// testConcurrentRequests tests handling of concurrent requests
func testConcurrentRequests(t *testing.T, baseURL string) {
	const numGoroutines = 20
	const requestsPerGoroutine = 10

	results := make(chan error, numGoroutines*requestsPerGoroutine)

	for i := 0; i < numGoroutines; i++ {
		go func(goroutineID int) {
			client := &http.Client{Timeout: 5 * time.Second}
			for j := 0; j < requestsPerGoroutine; j++ {
				resp, err := client.Get(baseURL + "/api/test")
				if err != nil {
					results <- err
					continue
				}
				resp.Body.Close()

				if resp.StatusCode != http.StatusOK {
					results <- fmt.Errorf("unexpected status: %d", resp.StatusCode)
					continue
				}

				results <- nil
			}
		}(i)
	}

	// Collect results
	successCount := 0
	errorCount := 0
	totalRequests := numGoroutines * requestsPerGoroutine

	for i := 0; i < totalRequests; i++ {
		select {
		case err := <-results:
			if err != nil {
				errorCount++
				t.Logf("Request error: %v", err)
			} else {
				successCount++
			}
		case <-time.After(10 * time.Second):
			t.Fatal("Concurrent requests test timed out")
		}
	}

	t.Logf("Concurrent requests: %d successful, %d failed", successCount, errorCount)

	if errorCount > totalRequests/10 { // Allow up to 10% failure rate
		t.Errorf("Too many failed requests: %d/%d", errorCount, totalRequests)
	}
}

// testServerLoad performs a simple load test
func testServerLoad(t *testing.T, baseURL string) {
	const duration = 2 * time.Second
	const workers = 10

	ctx, cancel := context.WithTimeout(context.Background(), duration)
	defer cancel()

	requests := make(chan struct{}, 1000)
	results := make(chan bool, 1000)

	// Start workers
	for i := 0; i < workers; i++ {
		go func() {
			client := &http.Client{Timeout: 1 * time.Second}
			for {
				select {
				case <-ctx.Done():
					return
				case <-requests:
					resp, err := client.Get(baseURL + "/api/test")
					if err != nil {
						results <- false
						continue
					}
					resp.Body.Close()
					results <- resp.StatusCode == http.StatusOK
				}
			}
		}()
	}

	// Generate load
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case requests <- struct{}{}:
			}
		}
	}()

	// Collect results
	successCount := 0
	totalCount := 0

	// Wait for test duration
	<-ctx.Done()
	close(requests)

	// Collect remaining results with timeout
	timeout := time.After(2 * time.Second)
	for {
		select {
		case success := <-results:
			totalCount++
			if success {
				successCount++
			}
		case <-timeout:
			goto done
		}
	}

done:
	t.Logf("Load test: %d/%d requests successful (%.1f%%)",
		successCount, totalCount, float64(successCount)/float64(totalCount)*100)

	if totalCount == 0 {
		t.Error("No requests completed during load test")
	} else if float64(successCount)/float64(totalCount) < 0.95 {
		t.Error("Success rate too low during load test")
	}
}

// TestMain can be used to set up and tear down test infrastructure
func TestMain(m *testing.M) {
	// Run tests
	code := m.Run()

	// Cleanup if needed
	os.Exit(code)
}
