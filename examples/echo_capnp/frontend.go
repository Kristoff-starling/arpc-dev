package main

import (
	"context"
	"fmt"
	"log"
	"net/http"

	echo "github.com/appnet-org/aprc/examples/echo_capnp/capnp"
	"github.com/appnet-org/aprc/internal/serializer"
	"github.com/appnet-org/aprc/pkg/rpc"
)

var echoClient echo.EchoServiceClient

func handler(w http.ResponseWriter, r *http.Request) {
	content := r.URL.Query().Get("key")
	log.Printf("Received HTTP request with key: %s\n", content)

	req, err := echo.CreateEchoRequest(content)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to create request: %v", err), http.StatusInternalServerError)
		return
	}

	resp, err := echoClient.Echo(context.Background(), req)
	if err != nil {
		http.Error(w, fmt.Sprintf("RPC call failed: %v", err), http.StatusInternalServerError)
		return
	}

	respContent, err := resp.GetContent()
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get response content: %v", err), http.StatusInternalServerError)
		return
	}
	log.Printf("RPC response: %s\n", respContent)
	fmt.Fprintf(w, "Response from RPC: %s\n", respContent)
}

func main() {
	// Create RPC client
	serializer := &serializer.CapnpSerializer{}
	client, err := rpc.NewClient(serializer, "127.0.0.1:9000")
	if err != nil {
		log.Fatal("Failed to create RPC client:", err)
	}

	// Create EchoService client
	echoClient = echo.NewEchoServiceClient(client)

	// Set up HTTP server
	http.HandleFunc("/", handler)
	log.Println("HTTP server listening on :8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal("HTTP server failed:", err)
	}
}
