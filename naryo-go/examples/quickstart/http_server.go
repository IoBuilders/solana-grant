package main

import (
	"bytes"
	"encoding/json"
	"io"
	"log"
	"net/http"
)

func main() {
	mux := http.NewServeMux()

	mux.HandleFunc("POST /contract-events", func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "error reading body", http.StatusBadRequest)
			return
		}
		defer r.Body.Close()

		var pretty bytes.Buffer
		if err := json.Indent(&pretty, body, "", "  "); err != nil {
			log.Printf("Contract event received:\n%s", body)
			return
		}
		log.Printf("Contract event received:\n%s", pretty.String())
	})

	mux.HandleFunc("POST /transactions", func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "error reading body", http.StatusBadRequest)
			return
		}
		defer r.Body.Close()

		var pretty bytes.Buffer
		if err := json.Indent(&pretty, body, "", "  "); err != nil {
			log.Printf("Transaction event received:\n%s", body)
			return
		}
		log.Printf("Transaction event received:\n%s", pretty.String())
	})

	log.Println("server listening on :8081")
	if err := http.ListenAndServe(":8081", mux); err != nil {
		log.Fatal(err)
	}
}
