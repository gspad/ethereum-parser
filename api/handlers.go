package api

import (
	"encoding/json"
	"net/http"

	"github.com/gspad/ethereum-parser/parser"
)

type API struct {
	Parser parser.Parser
}

func NewAPI(p parser.Parser) *API {
	return &API{Parser: p}
}

func (api *API) HandleSubscribe(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Address string `json:"address"`
	}

	json.NewDecoder(r.Body).Decode(&req)

	if req.Address == "" {
		http.Error(w, "address is required", http.StatusBadRequest)
		return
	}

	success := api.Parser.Subscribe(req.Address)

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]bool{"success": success})
}

func (api *API) HandleGetTransactions(w http.ResponseWriter, r *http.Request) {
	address := r.URL.Query().Get("address")
	if address == "" {
		http.Error(w, "address is required", http.StatusBadRequest)
		return
	}

	transactions := api.Parser.GetTransactions(address)

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(transactions)
}

func (api *API) HandleGetCurrentBlock(w http.ResponseWriter, r *http.Request) {
	currentBlock := api.Parser.GetCurrentBlock()

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]int{"currentBlock": currentBlock})
}
