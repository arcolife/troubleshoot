package collect

import (
	"net"
	"net/http"
	"net/http/httptest"
	"testing"

	"time"
)

func TestDoRequestGET(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		rw.WriteHeader(http.StatusOK)
		rw.Write([]byte("Hello, World!"))
	}))
	defer server.Close()

	response, err := doRequest("GET", server.URL, nil, "", false, 5*time.Second)

	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if response.StatusCode != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, response.StatusCode)
	}
}

func TestDoRequestPOST(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		rw.WriteHeader(http.StatusOK)
		rw.Write([]byte("Hello, World!"))
	}))
	defer server.Close()

	response, err := doRequest("POST", server.URL, nil, "", false, 5*time.Second)

	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if response.StatusCode != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, response.StatusCode)
	}
}

func TestDoRequestPUT(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		rw.WriteHeader(http.StatusOK)
		rw.Write([]byte("Hello, World!"))
	}))
	defer server.Close()

	response, err := doRequest("PUT", server.URL, nil, "", false, 5*time.Second)

	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if response.StatusCode != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, response.StatusCode)
	}
}

func TestDoRequestTimeout(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		time.Sleep(3 * time.Second) // Simulate a slow server
	}))
	defer server.Close()

	_, err := doRequest("GET", server.URL, nil, "", false, 1*time.Second)

	if err == nil {
		t.Fatal("Expected timeout error")
	}

	if !isTimeoutError(err) {
		t.Errorf("Expected timeout error, got: %v", err)
	}
}

func isTimeoutError(err error) bool {
	netErr, ok := err.(net.Error)
	return ok && netErr.Timeout()
}
