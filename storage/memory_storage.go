package storage

import "github.com/gspad/ethereum-parser/shared"

type MemoryStorage struct {
	currentBlock    int
	subscribedAddrs map[string]bool
	transactions    map[string][]shared.Transaction
}

func NewMemoryStorage() *MemoryStorage {
	return &MemoryStorage{
		currentBlock:    0,
		subscribedAddrs: make(map[string]bool),
		transactions:    make(map[string][]shared.Transaction),
	}
}

func (s *MemoryStorage) GetCurrentBlock() int {
	return s.currentBlock
}

func (s *MemoryStorage) SetCurrentBlock(block int) {
	s.currentBlock = block
}

func (s *MemoryStorage) Subscribe(address string) {
	s.subscribedAddrs[address] = true
}

func (s *MemoryStorage) GetSubscribedAddresses() map[string]bool {
	return s.subscribedAddrs
}

func (s *MemoryStorage) GetTransactions(address string) []shared.Transaction {
	return s.transactions[address]
}

func (s *MemoryStorage) AddTransaction(address string, transaction shared.Transaction) {
	s.transactions[address] = append(s.transactions[address], transaction)
}
