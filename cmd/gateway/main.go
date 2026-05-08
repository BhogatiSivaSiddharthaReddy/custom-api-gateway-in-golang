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

func helloWorld(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Received request:", r.URL.Path)
	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintln(w, `{"message": "Hello World"}`)
}

func pinghandler(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Received request:", r.URL.Path)
	time.Sleep(5 * time.Second)
	fmt.Fprintln(w, "PONG")
}

func main() {
	fmt.Println("Starting API Gateway")

	http.HandleFunc("/", helloWorld)
	http.HandleFunc("/ping", pinghandler)

	//http.ListenAndServe(":8080", nil)

	server := &http.Server{
		Addr:    ":8080",
		Handler: nil,
	}

	go func() {
		server.ListenAndServe()
	}()

	quit := make(chan os.Signal, 1)

	signal.Notify(quit, syscall.SIGTERM, os.Interrupt)

	<-quit

	fmt.Println("Shutting Down Server")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)

	defer cancel()

	server.Shutdown(ctx)
}
