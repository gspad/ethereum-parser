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

	if !fakeParser.Subscribed["0x123"] {
		t.Errorf("address not subscribed")
	}
}

func TestHandleGetTransactions(t *testing.T) {
	fakeParser := parser.NewFakeParser()
	apiHandler := NewAPI(fakeParser)

	address := "0x123"
	transactions := []shared.Transaction{
		{From: "0xabc", To: address, Value: "100"},
		{From: address, To: "0xdef", Value: "50"},
	}
	fakeParser.Transactions[address] = transactions

	req, err := http.NewRequest("GET", "/transactions?address=0x123", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(apiHandler.HandleGetTransactions)
	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	var got []shared.Transaction
	if err := json.NewDecoder(rr.Body).Decode(&got); err != nil {
		t.Errorf("failed to decode response body: %v", err)
	}

	if len(got) != len(transactions) {
		t.Errorf("expected %d transactions, got %d", len(transactions), len(got))
	}

	for i, tx := range transactions {
		if tx != got[i] {
			t.Errorf("expected transaction %v, got %v", tx, got[i])
		}
	}
}

func TestHandleGetCurrentBlock(t *testing.T) {
	fakeParser := parser.NewFakeParser()
	apiHandler := NewAPI(fakeParser)

	fakeParser.CurrentBlock = 1207

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
