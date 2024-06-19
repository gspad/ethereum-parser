package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/gspad/ethereum-parser/api"
	"github.com/gspad/ethereum-parser/parser"
	"github.com/gspad/ethereum-parser/storage"
)

func main() {
	memoryStorage := storage.NewMemoryStorage()

	httpClient := &parser.RealHTTPClient{}

	p := parser.NewParser(httpClient, memoryStorage)

	apiHandler := api.NewAPI(p)

	http.HandleFunc("/subscribe", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
			return
		}
		apiHandler.HandleSubscribe(w, r)
	})
	http.HandleFunc("/transactions", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
			return
		}
		apiHandler.HandleGetTransactions(w, r)
	})
	http.HandleFunc("/block", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
			return
		}
		apiHandler.HandleGetCurrentBlock(w, r)
	})

	fmt.Println("Starting server on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
