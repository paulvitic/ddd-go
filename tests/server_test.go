package ddd_tests

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"testing"
	"time"

	"github.com/paulvitic/ddd-go"
	"github.com/paulvitic/ddd-go/tests/test_server/test_context"
)

// TestServer tests the complete server lifecycle and endpoint functionality
func TestServer(t *testing.T) {

	// Create server with the context
	server := ddd.NewServer(ddd.NewServerConfig()).
		WithContexts(test_context.TestContext)

	// Start server in a goroutine
	_, cancelServer := context.WithCancel(context.Background())
	serverErrChan := make(chan error, 1)

	go func() {
		if err := server.Start(); err != nil && err != http.ErrServerClosed {
			serverErrChan <- err
		}
	}()

	// Wait for server to start
	time.Sleep(100 * time.Millisecond)

	// Check if server started successfully
	select {
	case err := <-serverErrChan:
		t.Fatalf("Server failed to start: %v", err)
	default:
		// Server started successfully
	}

	// Run tests
	t.Run("HealthCheck", func(t *testing.T) {
		testHealthCheck(t)
	})

	t.Run("GET_Endpoint", func(t *testing.T) {
		testGetEndpoint(t)
	})

	t.Run("POST_Endpoint", func(t *testing.T) {
		testPostEndpoint(t)
	})

	// t.Run("PUT_Endpoint", func(t *testing.T) {
	// 	testPutEndpoint(t)
	// })

	// t.Run("DELETE_Endpoint", func(t *testing.T) {
	// 	testDeleteEndpoint(t)
	// })

	t.Run("UnsupportedMethod", func(t *testing.T) {
		testUnsupportedMethod(t)
	})

	// t.Run("RequestScoping", func(t *testing.T) {
	// 	testRequestScoping(t)
	// })

	// Stop the server
	cancelServer()

	// Give the server time to shut down gracefully
	time.Sleep(100 * time.Millisecond)

	t.Log("Server test suite completed successfully")
}

// testHealthCheck tests the health check endpoint
func testHealthCheck(t *testing.T) {
	resp, err := http.Get("http://localhost:8081/")
	if err != nil {
		t.Fatalf("Failed to make health check request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("Failed to read response body: %v", err)
	}

	expected := "Status: UP"
	if string(body) != expected {
		t.Errorf("Expected body '%s', got '%s'", expected, string(body))
	}

	t.Log("Health check endpoint working correctly")
}

// testGetEndpoint tests the GET endpoint
func testGetEndpoint(t *testing.T) {
	resp, err := http.Get("http://localhost:8081/test/users/1")
	if err != nil {
		t.Fatalf("Failed to make GET request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}

	// if resp.Header.Get("Content-Type") != "application/json" {
	// 	t.Errorf("Expected Content-Type application/json, got %s", resp.Header.Get("Content-Type"))
	// }

	var response map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to decode JSON response: %v", err)
	}

	expectedName := "User One"
	if response["Name"] != expectedName {
		t.Errorf("Expected name '%s', got '%v'", expectedName, response["Name"])
	}

	t.Log("GET endpoint working correctly")
}

// testPostEndpoint tests the POST endpoint
func testPostEndpoint(t *testing.T) {
	requestBody := map[string]interface{}{
		"userId": "1",
	}
	jsonBody, _ := json.Marshal(requestBody)

	resp, err := http.Post("http://localhost:8081/test/users", "application/json", bytes.NewBuffer(jsonBody))
	if err != nil {
		t.Fatalf("Failed to make POST request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		t.Errorf("Expected status 201, got %d", resp.StatusCode)
	}

	var response map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to decode JSON response: %v", err)
	}

	expectedMessage := "1"
	if response["message"] != expectedMessage {
		t.Errorf("Expected message '%s', got '%v'", expectedMessage, response["message"])
	}

	t.Log("POST endpoint working correctly")
}

// testPutEndpoint tests the PUT endpoint
func testPutEndpoint(t *testing.T) {
	requestBody := map[string]interface{}{
		"update": "data",
	}
	jsonBody, _ := json.Marshal(requestBody)

	req, err := http.NewRequest("PUT", "http://localhost:8081/test/users", bytes.NewBuffer(jsonBody))
	if err != nil {
		t.Fatalf("Failed to create PUT request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("Failed to make PUT request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}

	var response map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to decode JSON response: %v", err)
	}

	if response["method"] != "PUT" {
		t.Errorf("Expected method 'PUT', got '%v'", response["method"])
	}

	t.Log("PUT endpoint working correctly")
}

// testDeleteEndpoint tests the DELETE endpoint
func testDeleteEndpoint(t *testing.T) {
	req, err := http.NewRequest("DELETE", "http://localhost:8081/test/users", nil)
	if err != nil {
		t.Fatalf("Failed to create DELETE request: %v", err)
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("Failed to make DELETE request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent {
		t.Errorf("Expected status 204, got %d", resp.StatusCode)
	}

	// DELETE endpoint returns no content, so we just check the status
	t.Log("DELETE endpoint working correctly")
}

// testUnsupportedMethod tests that unsupported methods return 405
func testUnsupportedMethod(t *testing.T) {
	req, err := http.NewRequest("TRACE", "http://localhost:8081/test/users", nil)
	if err != nil {
		t.Fatalf("Failed to create TRACE request: %v", err)
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("Failed to make TRACE request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusMethodNotAllowed {
		t.Errorf("Expected status 405 for unsupported method, got %d", resp.StatusCode)
	}

	t.Log("Unsupported method correctly returns 405")
}

// Benchmark tests
func BenchmarkGetEndpoint(b *testing.B) {
	server := ddd.NewServer(ddd.NewServerConfig("configs/server_benchmark")).
		WithContexts(test_context.TestContext)

	go func() {
		server.Start()
	}()

	time.Sleep(100 * time.Millisecond) // Wait for server to start

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			resp, err := http.Get("http://localhost:8082/test/users")
			if err != nil {
				b.Fatalf("Request failed: %v", err)
			}
			resp.Body.Close()
		}
	})
}
