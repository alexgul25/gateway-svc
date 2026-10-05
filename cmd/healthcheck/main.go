package main

import (
	"fmt"
	"net"
	"net/http"
	"os"
	"time"

	"github.com/alexgul25/gateway-svc/internal/http/routing"
)

const requestTimeout = 2 * time.Second

func main() {
	if err := check(os.Getenv("SERVER_ADDR")); err != nil {
		fmt.Fprintln(os.Stderr, "unhealthy:", err)
		os.Exit(1)
	}
}

func check(serverAddr string) error {
	_, port, err := net.SplitHostPort(serverAddr)
	if err != nil {
		return fmt.Errorf("invalid SERVER_ADDR %q: %w", serverAddr, err)
	}

	url := "http://" + net.JoinHostPort("127.0.0.1", port) + routing.PathHealth

	client := http.Client{Timeout: requestTimeout}
	resp, err := client.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status %s", resp.Status)
	}

	return nil
}
