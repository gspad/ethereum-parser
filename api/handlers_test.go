package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gspad/ethereum-parser/parser"
	"github.com/gspad/ethereum-parser/shared"
)

func TestHandleSubscribe(t *testing.T) {
	fakeParser := parser.NewFakeParser()
	apiHandler := NewAPI(fakeParser)

	req, err := http.NewRequest("POST", "/subscribe", bytes.NewBuffer([]byte(`{"address":"0x123"}`)))
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(apiHandler.HandleSubscribe)
	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	var got map[string]bool
	if err := json.NewDecoder(rr.Body).Decode(&got); err != nil {
		t.Fatalf("failed to decode response body: %v", err)
	}

	expected := map[string]bool{"success": true}
	if got["success"] != expected["success"] {
		t.Errorf("expected %v, got %v", expected, got)
	}
}

func TestParser_GetTransactions(t *testing.T) {
	p := parser.NewFakeParser()

	apiHandler := NewAPI(p)

	tx := shared.Transaction{
		From:  "0x100",
		To:    "0x200",
		Value: "100",
	}
	p.AddTransaction("0xToAddress", tx)

	req, err := http.NewRequest("GET", "/transactions?address=0xToAddress", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(apiHandler.HandleGetTransactions)
	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	var transactions []shared.Transaction
	err = json.NewDecoder(rr.Body).Decode(&transactions)
	if err != nil {
		t.Fatal(err)
	}

	if len(transactions) != 1 {
		t.Fatalf("expected 1 transaction, got %d", len(transactions))
	}

	if transactions[0] != tx {
		t.Errorf("expected transaction %v, got %v", tx, transactions[0])
	}
}

func TestHandleGetCurrentBlock(t *testing.T) {
	p := parser.NewFakeParser()
	apiHandler := NewAPI(p)

	p.Storage.SetCurrentBlock(1207)

	req, err := http.NewRequest("GET", "/block", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(apiHandler.HandleGetCurrentBlock)
	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	var got map[string]int
	if err := json.NewDecoder(rr.Body).Decode(&got); err != nil {
		t.Errorf("failed to decode response body: %v", err)
	}

	expected := map[string]int{"currentBlock": 1207}
	if got["currentBlock"] != expected["currentBlock"] {
		t.Errorf("expected currentBlock to be %d, got %d", expected["currentBlock"], got["currentBlock"])
	}
}
