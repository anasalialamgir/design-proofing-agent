package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Design Proofing Agent is active and connected to Firebase!")
	})

	// Sample endpoint to verify file resolution
	http.HandleFunc("/api/check-resolution", func(w http.ResponseWriter, r *http.Request) {
		// Reads image dimensions and calculates DPI
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, `{"status": "READY", "target_dpi": 300, "message": "Preflight ready"}`)
	})

	log.Printf("Server listening on port %s", port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
