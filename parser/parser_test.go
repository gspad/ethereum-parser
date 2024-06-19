package parser

import (
	"io"
	"net/http"
	"testing"

	"github.com/gspad/ethereum-parser/shared"
	"github.com/gspad/ethereum-parser/storage"
)

type FakeHTTPClient struct {
	Response *http.Response
	Err      error
}

func (c *FakeHTTPClient) Post(url, contentType string, body io.Reader) (*http.Response, error) {
	return c.Response, c.Err
}

func TestGetCurrentBlock(t *testing.T) {
	parser := NewParser(&FakeHTTPClient{}, storage.NewMemoryStorage())
	parser.storage.SetCurrentBlock(100)

	if block := parser.GetCurrentBlock(); block != 100 {
		t.Errorf("expected block 100, got %d", block)
	}
}

func TestSubscribe(t *testing.T) {
	parser := NewParser(&FakeHTTPClient{}, storage.NewMemoryStorage())
	address := "0x123"

	parser.storage.Subscribe(address)

	if !parser.storage.GetSubscribedAddresses()[address] {
		t.Errorf("expected address %s to be subscribed", address)
	}
}

func TestGetTransactions(t *testing.T) {
	parser := NewParser(&FakeHTTPClient{}, storage.NewMemoryStorage())
	address := "0x123"
	transactions := []shared.Transaction{
		{From: "0xabc", To: address, Value: "100"},
		{From: address, To: "0xdef", Value: "50"},
	}

	parser.storage.AddTransaction(address, transactions[0])
	parser.storage.AddTransaction(address, transactions[1])

	retrievedTransactions := parser.storage.GetTransactions((address))

	if len(retrievedTransactions) != 2 {
		t.Errorf("expected 2 transactions, got %d", len(retrievedTransactions))
	}

	for i, tx := range transactions {
		if tx != retrievedTransactions[i] {
			t.Errorf("expected transaction %v, got %v", tx, retrievedTransactions[i])
		}
	}
}
