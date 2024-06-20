package parser

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/gspad/ethereum-parser/shared"
	"github.com/gspad/ethereum-parser/storage"
)

type HTTPClient interface {
	Post(url, contentType string, body io.Reader) (*http.Response, error)
}

type RealHTTPClient struct{}

func (c *RealHTTPClient) Post(url, contentType string, body io.Reader) (*http.Response, error) {
	return http.Post(url, contentType, body)
}

type NotificationService interface {
	Notify(address string, transaction shared.Transaction)
}

type Parser interface {
	Start(url string)
	Parse(url string)
	GetCurrentBlock() int
	Subscribe(address string) bool
	GetTransactions(address string) []shared.Transaction
}

type ParserImpl struct {
	mu         sync.Mutex
	storage    storage.Storage
	httpClient HTTPClient
}

func NewParser(httpClient HTTPClient, storage storage.Storage) *ParserImpl {
	parser := &ParserImpl{
		storage:    storage,
		httpClient: httpClient,
	}
	return parser
}

func (p *ParserImpl) Start(url string) {
	for {
		p.Parse(url)
		time.Sleep(5 * time.Second)
	}
}

func (p *ParserImpl) Parse(url string) {
	payload := map[string]interface{}{
		"jsonrpc": "2.0",
		"method":  "eth_blockNumber",
		"params":  []interface{}{},
		"id":      1,
	}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		log.Println("Error marshaling JSON payload:", err)
		return
	}

	resp, err := p.httpClient.Post(url, "application/json", bytes.NewBuffer(payloadBytes))
	if err != nil {
		log.Println("Error fetching block number:", err)
		return
	}
	defer resp.Body.Close()

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		log.Println("Error decoding JSON:", err)
		return
	}

	blockHex, ok := result["result"].(string)
	if !ok {
		log.Println("Error: result is not a string")
		return
	}

	blockNum, err := strconv.ParseInt(blockHex[2:], 16, 64)
	if err != nil {
		log.Println("Error parsing block number:", err)
		return
	}

	if int(blockNum) > p.storage.GetCurrentBlock() {
		log.Println("New block:", blockNum)
		p.storage.SetCurrentBlock(int(blockNum))
		p.storeMatchingTransactions(url, blockHex)
	}
}

func (p *ParserImpl) GetCurrentBlock() int {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.storage.GetCurrentBlock()
}

func (p *ParserImpl) Subscribe(address string) bool {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.storage.Subscribe(address)
	log.Println("Subscribed to", address)

	subscribedAddr := p.storage.GetSubscribedAddresses()[address]
	fmt.Println("Subscribed addresses to: ", subscribedAddr)
	return true
}

func (p *ParserImpl) GetTransactions(address string) []shared.Transaction {
	fmt.Println("Getting transactions for", address)
	return p.storage.GetTransactions(address)
}

func (p *ParserImpl) storeMatchingTransactions(url, blockHex string) {
	blockDetails, err := p.fetchBlockDetails(url, blockHex)
	if err != nil {
		log.Println("Error fetching block details:", err)
		return
	}

	txs, ok := blockDetails["transactions"].([]interface{})
	if !ok {
		log.Printf("Error: transactions field is not a slice of interface{}. Got: %T", blockDetails["transactions"])
		return
	}

	for _, tx := range txs {
		txMap, ok := tx.(map[string]interface{})
		if !ok {
			log.Printf("Error: transaction map is not a map[string]interface{}. Got %T", tx)
			continue
		}

		from, fromOk := txMap["from"].(string)
		to, toOk := txMap["to"].(string)

		if !fromOk || !toOk {
			log.Println("Error: from or to field is not a string. Got from:", from, "to:", to)
			continue
		}

		for address := range p.storage.GetSubscribedAddresses() {
			if address == from || address == to {
				transaction := shared.Transaction{
					From:  from,
					To:    to,
					Value: txMap["value"].(string),
				}

				log.Println("Storing transaction")
				p.StoreTransaction(address, transaction)
			}
		}
	}
}

func (p *ParserImpl) fetchBlockDetails(url, blockHex string) (map[string]interface{}, error) {
	payload := map[string]interface{}{
		"jsonrpc": "2.0",
		"method":  "eth_getBlockByNumber",
		"params":  []interface{}{blockHex, true},
		"id":      1,
	}
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	resp, err := p.httpClient.Post(url, "application/json", bytes.NewBuffer(payloadBytes))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return result["result"].(map[string]interface{}), nil
}

func (p *ParserImpl) StoreTransaction(address string, transaction shared.Transaction) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.storage.AddTransaction(address, transaction)
}
