package main

import (
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"strconv"
	"strings"
)

const (
	defaultPort            = "8080"
	defaultDestinationPort = "8080"
)

func handler(w http.ResponseWriter, r *http.Request) {
	// Retrieve NODE_NAME and POD_NAME from environment variables
	nodeName := os.Getenv("NODE_NAME")
	podName := os.Getenv("POD_NAME")

	// Retrieve the hostname of the system
	hostname, err := os.Hostname()
	if err != nil {
		log.Printf("Error retrieving hostname: %v", err)
		hostname = ""
	}

	// Get the destination port from the environment variable or default to 8080
	destinationPort := os.Getenv("DESTINATION_PORT")
	if destinationPort == "" {
		destinationPort = defaultDestinationPort
	}

	// Check if the destination port is valid
	if _, err := net.LookupPort("tcp", destinationPort); err != nil {
		log.Printf("Invalid destination port: %v", err)
		destinationPort = defaultDestinationPort // Fallback to default port if invalid
	}

	// Get the local address (destination address) using the request's context
	var destinationAddress string
	if conn, ok := r.Context().Value(http.LocalAddrContextKey).(net.Addr); ok {
		destinationAddress = conn.String()
	} else {
		destinationAddress = "unknown"
	}

	// Get the full requested URL
	requestedURL := fmt.Sprintf("http://%s%s", r.Host, r.RequestURI)

	// Parse the URL path to check for segment size request
	var requestedSize int
	path := strings.TrimPrefix(r.URL.Path, "/")
	if strings.HasSuffix(path, "_Response") {
		sizeStr := strings.TrimSuffix(path, "_Response")
		requestedSize, err = strconv.Atoi(sizeStr)
		if err != nil {
			log.Printf("Error parsing requested size: %v", err)
			requestedSize = 0 // Default to 0 if parsing fails
		}
	}

	// Prepare the response
	var response strings.Builder
	if nodeName != "" {
		response.WriteString(fmt.Sprintf("Node Name              : %s\n", nodeName))
	}
	if podName != "" {
		response.WriteString(fmt.Sprintf("Pod Name               : %s\n", podName))
	}
	if hostname != "" {
		response.WriteString(fmt.Sprintf("Hostname               : %s\n", hostname))
	}
	response.WriteString(fmt.Sprintf("Source Address    : %s\n", r.RemoteAddr))
	response.WriteString(fmt.Sprintf("Destination Address    : %s\n", destinationAddress))
	response.WriteString(fmt.Sprintf("Full URL               : %s\n", requestedURL))

	// Add incoming HTTP headers
	response.WriteString("\n--- Incoming HTTP Headers ---\n")
	for name, values := range r.Header {
		response.WriteString(fmt.Sprintf("%-20s: %s\n", name, strings.Join(values, ", ")))
	}

	// If a specific size is requested, fill the response to meet the size
	// Handle filler content or error based on requested size
	if requestedSize > 0 {
		baseResponse := response.String()
		baseSize := len(baseResponse)

		if requestedSize < baseSize {
			response.WriteString(fmt.Sprintf("\nERROR: Requested size (%d bytes) is too small. Base response size is %d bytes.\n", requestedSize, baseSize))
		} else {
			fillerSize := requestedSize - baseSize
			filler := strings.Repeat("#", fillerSize)
			response.WriteString(filler)
			response.WriteString(fmt.Sprintf("\nFiller content added: %d bytes\n", fillerSize))
		}
	}

	// Write the response
	w.Write([]byte(response.String()))
}

func main() {
	http.HandleFunc("/", handler)
	port := os.Getenv("PORT")
	if port == "" {
		port = defaultPort
	}

	ipMode := os.Getenv("IP_MODE")
	var listenAddr string
	switch ipMode {
	case "ipv4":
		listenAddr = "0.0.0.0:" + port
	case "ipv6":
		listenAddr = "[::]:" + port
	default:
		listenAddr = ":" + port
	}
    // Create a TCP listener explicitly for IPv4
    if ipMode == "ipv4" {
        listener, err := net.Listen("tcp4", listenAddr)
        if err != nil {
            log.Fatalf("Failed to start IPv4 listener: %v", err)
        }
        log.Printf("Starting server on %s with IP mode %s", listenAddr, ipMode)
        if err := http.Serve(listener, nil); err != nil {
            log.Fatalf("Server failed: %v", err)
        }
    } else {
        log.Printf("Starting server on %s with IP mode %s", listenAddr, ipMode)
        if err := http.ListenAndServe(listenAddr, nil); err != nil {
            log.Fatalf("Server failed: %v", err)
        }
    }

}
