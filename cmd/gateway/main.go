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

func proxyHandler(w http.ResponseWriter, r *http.Request) {
	ParentCtx := r.Context()

	Childctx, cancel := context.WithTimeout(ParentCtx, 30*time.Second)

	defer cancel()

	err := backendCall(Childctx)

	if err != nil {
		http.Error(
			w,
			"Gateway Timeout",
			http.StatusGatewayTimeout,
		)
		return
	}
	fmt.Fprintln(w, "Request Succesfull")
}

func backendCall(ctx context.Context) error {
	select {
	case <-ctx.Done():
		fmt.Println("Backend Canclled:", ctx.Err())
		return ctx.Err()
	case <-time.After(10 * time.Second):
		fmt.Println("request done")
		return nil
	}

}

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
	http.HandleFunc("/day5", proxyHandler)

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
