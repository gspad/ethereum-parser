package parser

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"testing"

	"github.com/gspad/ethereum-parser/shared"
	"github.com/gspad/ethereum-parser/storage"
)

type FakeHTTPClient struct {
	Responses map[string]*http.Response
	Err       error
}

func (c *FakeHTTPClient) Post(url, contentType string, body io.Reader) (*http.Response, error) {
	if c.Err != nil {
		return nil, c.Err
	}

	buf := new(bytes.Buffer)
	buf.ReadFrom(body)
	requestBody := buf.String()

	log.Printf("Request body: %s", requestBody)

	if response, exists := c.Responses[requestBody]; exists {
		log.Printf("Matched response: %v", response)
		return response, nil
	}

	log.Printf("No matched response for: %s", requestBody)
	return &http.Response{
		StatusCode: http.StatusNotFound,
		Body:       io.NopCloser(bytes.NewBufferString("Not Found")),
	}, nil
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

	retrievedTransactions := parser.storage.GetTransactions(address)

	if len(retrievedTransactions) != 2 {
		t.Errorf("expected 2 transactions, got %d", len(retrievedTransactions))
	}

	for i, tx := range transactions {
		if tx != retrievedTransactions[i] {
			t.Errorf("expected transaction %v, got %v", tx, retrievedTransactions[i])
		}
	}
}

func TestParse(t *testing.T) {
	storage := storage.NewMemoryStorage()

	blockNumberResponse := `{"jsonrpc":"2.0","result":"0x4b7","id":1}`
	blockDetailsResponse := `{
		"jsonrpc": "2.0",
		"result": {
			"number": "0x4b7",
			"transactions": [
				{
					"from": "0xabc",
					"to": "0x123",
					"value": "100"
				}
			]
		},
		"id": 1
	}`

	blockNumberPayload := map[string]interface{}{
		"jsonrpc": "2.0",
		"method":  "eth_blockNumber",
		"params":  []interface{}{},
		"id":      1,
	}
	blockNumberPayloadBytes, _ := json.Marshal(blockNumberPayload)

	blockDetailsPayload := map[string]interface{}{
		"jsonrpc": "2.0",
		"method":  "eth_getBlockByNumber",
		"params":  []interface{}{"0x4b7", true},
		"id":      1,
	}
	blockDetailsPayloadBytes, _ := json.Marshal(blockDetailsPayload)

	normalizedBlockNumberPayload := string(blockNumberPayloadBytes)
	normalizedBlockDetailsPayload := string(blockDetailsPayloadBytes)

	fakeClient := &FakeHTTPClient{
		Responses: map[string]*http.Response{
			normalizedBlockNumberPayload: {
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(bytes.NewBufferString(blockNumberResponse)),
			},
			normalizedBlockDetailsPayload: {
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(bytes.NewBufferString(blockDetailsResponse)),
			},
		},
	}

	p := NewParser(fakeClient, storage)
	p.storage.SetCurrentBlock(0)
	p.Subscribe("0x123")

	fmt.Println("Calling Parse")
	p.Parse("http://fake-url")

	if currentBlock := storage.GetCurrentBlock(); currentBlock != 1207 {
		t.Errorf("expected current block to be 1207, got %d", currentBlock)
	}

	transactions := p.GetTransactions("0x123")
	if len(transactions) != 1 {
		t.Errorf("expected 1 transaction, got %d", len(transactions))
	}
	if transactions[0].From != "0xabc" || transactions[0].To != "0x123" || transactions[0].Value != "100" {
		t.Errorf("transaction details do not match: %+v", transactions[0])
	}
}
