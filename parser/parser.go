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
	p := &ParserImpl{
		storage:    storage,
		httpClient: httpClient,
	}
	go p.updateCurrentBlock("https://cloudflare-eth.com")
	return p
}

func (p *ParserImpl) updateCurrentBlock(url string) {
	for {
		resp, err := p.httpClient.Post(url, "application/json", bytes.NewBuffer([]byte(`{"jsonrpc":"2.0","method":"eth_blockNumber","params":[],"id":1}`)))
		if err != nil {
			log.Println("Error fetching block number:", err)
			time.Sleep(time.Second)
			continue
		}

		var result map[string]interface{}
		if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
			log.Println("Error decoding JSON:", err)
			resp.Body.Close()
			time.Sleep(time.Second)
			continue
		}
		resp.Body.Close()

		blockHex, ok := result["result"].(string)
		if !ok {
			log.Println("Error: result is not a string")
			time.Sleep(time.Second)
			continue
		}

		blockNum, err := strconv.ParseInt(blockHex[2:], 16, 64)
		if err != nil {
			log.Println("Error parsing block number:", err)
			time.Sleep(time.Second)
			continue
		}

		p.mu.Lock()
		if int(blockNum) > p.storage.GetCurrentBlock() {
			log.Println("New block:", blockNum)
			p.storage.SetCurrentBlock(int(blockNum))

			//check for transactions involving subscribed addresses in the new block
			p.checkTransactionsInBlock(url, blockHex)
		}
		p.mu.Unlock()

		time.Sleep(10 * time.Second)
	}
}

func (p *ParserImpl) GetCurrentBlock() int {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.storage.GetCurrentBlock()
}

func (p *ParserImpl) Subscribe(address string) bool {
	log.Println("Subscribing to", address)
	p.mu.Lock()
	defer p.mu.Unlock()
	p.storage.Subscribe(address)
	log.Println("Subscribed to", address)
	return true
}

func (p *ParserImpl) GetTransactions(address string) []shared.Transaction {
	p.mu.Lock()
	currentBlock := p.storage.GetCurrentBlock()
	p.mu.Unlock()

	url := "https://cloudflare-eth.com"
	transactions := []shared.Transaction{}

	blockHex := fmt.Sprintf("0x%x", currentBlock)
	blockDetails, err := p.fetchBlockDetails(url, blockHex)
	if err != nil {
		log.Println("Error fetching block details:", err)
		return transactions
	}

	txs, ok := blockDetails["transactions"].([]interface{})
	if !ok {
		log.Println("Error: transactions field is not a slice")
		return transactions
	}

	for _, tx := range txs {
		txMap, ok := tx.(map[string]interface{})
		if !ok {
			log.Println("Error: transaction is not a map")
			continue
		}

		from, fromOk := txMap["from"].(string)
		to, toOk := txMap["to"].(string)

		if !fromOk || !toOk {
			log.Println("Error: from or to field is not a string")
			continue
		}

		if address == from || address == to {
			transaction := shared.Transaction{
				From:  from,
				To:    to,
				Value: txMap["value"].(string),
			}
			transactions = append(transactions, transaction)
		}
	}

	return transactions
}

func (p *ParserImpl) checkTransactionsInBlock(url, blockHex string) {
	blockDetails, err := p.fetchBlockDetails(url, blockHex)
	if err != nil {
		log.Println("Error fetching block details:", err)
		return
	}

	txs, ok := blockDetails["transactions"].([]interface{})
	if !ok {
		log.Println("Error: transactions field is not a slice")
		return
	}

	for _, tx := range txs {
		txMap, ok := tx.(map[string]interface{})
		if !ok {
			log.Println("Error: transaction is not a map")
			continue
		}

		from, fromOk := txMap["from"].(string)
		to, toOk := txMap["to"].(string)

		if !fromOk || !toOk {
			log.Println("Error: from or to field is not a string")
			continue
		}

		for address := range p.storage.GetSubscribedAddresses() {
			if address == from || address == to {
				transaction := shared.Transaction{
					From:  from,
					To:    to,
					Value: txMap["value"].(string),
				}
				//notify about the transaction (notification service not implemented)
				log.Println("Transaction involving subscribed address:", transaction)
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
