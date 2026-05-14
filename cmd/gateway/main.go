package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

//
// BACKEND SERVER
//

func backendHandler(w http.ResponseWriter, r *http.Request) {

	fmt.Println("Backend received request")

	// Simulate slow backend
	time.Sleep(10 * time.Second)

	fmt.Fprintln(w, "Backend response completed")
}

//
// GATEWAY SERVER
//

func gatewayHandler(w http.ResponseWriter, r *http.Request) {

	fmt.Println("Gateway received request")

	parentCtx := r.Context()

	// Gateway timeout policy
	ctx, cancel := context.WithTimeout(
		parentCtx,
		3*time.Second,
	)

	defer cancel()

	// Outgoing backend request with context propagation
	req, err := http.NewRequestWithContext(
		ctx,
		"GET",
		"http://localhost:8081",
		nil,
	)

	if err != nil {

		http.Error(
			w,
			err.Error(),
			http.StatusInternalServerError,
		)

		return
	}

	client := &http.Client{}

	resp, err := client.Do(req)

	// Backend timed out / cancelled
	if err != nil {

		fmt.Println("Gateway timeout:", err)

		http.Error(
			w,
			"Gateway Timeout",
			http.StatusGatewayTimeout,
		)

		return
	}

	defer resp.Body.Close()

	fmt.Fprintln(w, "Gateway received backend response")
}

//
// MAIN
//

func main() {

	fmt.Println("Starting API Gateway")

	//
	// BACKEND SERVER
	//

	backendMux := http.NewServeMux()

	backendMux.HandleFunc("/", backendHandler)

	backendServer := &http.Server{
		Addr:    ":8081",
		Handler: backendMux,
	}

	go func() {

		fmt.Println("Backend server running on :8081")

		if err := backendServer.ListenAndServe(); err != nil &&
			err != http.ErrServerClosed {

			fmt.Println("Backend server error:", err)
		}
	}()

	//
	// GATEWAY SERVER
	//

	gatewayMux := http.NewServeMux()

	gatewayMux.HandleFunc("/day5", gatewayHandler)

	server := &http.Server{
		Addr:    ":8080",
		Handler: gatewayMux,
	}

	go func() {

		fmt.Println("Gateway server running on :8080")

		if err := server.ListenAndServe(); err != nil &&
			err != http.ErrServerClosed {

			fmt.Println("Gateway server error:", err)
		}
	}()

	//
	// GRACEFUL SHUTDOWN
	//

	quit := make(chan os.Signal, 1)

	signal.Notify(
		quit,
		syscall.SIGTERM,
		os.Interrupt,
	)

	<-quit

	fmt.Println("Shutting down servers...")

	ctx, cancel := context.WithTimeout(
		context.Background(),
		10*time.Second,
	)

	defer cancel()

	server.Shutdown(ctx)

	backendServer.Shutdown(ctx)

	fmt.Println("Servers stopped gracefully")
}
