package main

import (
	"fmt"
	"net/http"
	"os"
	"time"
)

func main() {
	url := "http://localhost:8080/healthz"
	if len(os.Args) > 1 {
		url = os.Args[1]
	}

	client := &http.Client{
		Timeout: 5 * time.Second,
	}

	resp, err := client.Get(url)
	if err != nil {
		fmt.Fprintf(os.Stderr, "healthcheck failed: %v\n", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusServiceUnavailable {
		fmt.Fprintf(os.Stderr, "healthcheck returned status: %d\n", resp.StatusCode)
		os.Exit(1)
	}

	// Status 200 or 503 are acceptable (503 means DB is down but server is up)
	os.Exit(0)
}
