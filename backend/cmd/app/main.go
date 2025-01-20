package main

// entry point of application

// DIGITAL OCEAN HTTP WEB TUTORIAL BELOW

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"context"
	"net"
	"encoding/json"
)

const keyServerAddr = "serverAddr"

type HelloResponse struct {
	Message string `json:"message"`
}

func getRoot(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	hasFirst := r.URL.Query().Has("first")
	first := r.URL.Query().Get("first")
	hasSecond := r.URL.Query().Has("second")
	second := r.URL.Query().Get("second")

	fmt.Printf("%s: got / request. first(%t)=%s, second(%t)=%s\n",
		ctx.Value(keyServerAddr),
		hasFirst, first,
		hasSecond, second)
	io.WriteString(w, "This is my website!\n")
}

func getHello(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	fmt.Printf("%s: got /hello request\n", ctx.Value(keyServerAddr))
	
	// myName := r.PostFormValue("myName")
	// if myName == "" {
	// 	w.Header().Set("x-missing-field", "myName")
	// 	w.WriteHeader(http.StatusBadRequest)
	// 	return
	// }
	// io.WriteString(w, fmt.Sprintf("Hello, %s!\n", myName))

	// Set the response header to indicate the content type is JSON.
	w.Header().Set("Content-Type", "application/json")
	// Create a response object.
	response := HelloResponse{Message: "Hello, World!"}
	// Encode the response object to JSON and send it to the client.
	json.NewEncoder(w).Encode(response)
}

// corsMiddleware adds CORS headers to the response.
// This is needed to access different ports in local host
func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Set CORS headers
		w.Header().Set("Access-Control-Allow-Origin", "*") // Replace '*' with specific origins if needed. 
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		// Handle preflight OPTIONS requests
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		// Pass the request to the next handler
		next.ServeHTTP(w, r)
	})
}

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/", getRoot)
	mux.HandleFunc("/hello", getHello)

	// Wrap the mux with the CORS middleware
	handlerWithCors := corsMiddleware(mux)

	ctx, cancelCtx := context.WithCancel(context.Background())

	serverOne := &http.Server{
		Addr:    ":3333",
		Handler: handlerWithCors,
		BaseContext: func(l net.Listener) context.Context {
			ctx = context.WithValue(ctx, keyServerAddr, l.Addr().String())
			return ctx
		},
	}

	go func() {
		err := serverOne.ListenAndServe()
		if errors.Is(err, http.ErrServerClosed) {
			fmt.Printf("server one closed\n")
		} else if err != nil {
			fmt.Printf("error listening for server one: %s\n", err)
		}
		cancelCtx()
	}()

	<-ctx.Done()
}