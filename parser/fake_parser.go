package parser

import (
	"fmt"

	"github.com/gspad/ethereum-parser/shared"
	"github.com/gspad/ethereum-parser/storage"
)

type FakeParser struct {
	Storage storage.Storage
}

func (f *FakeParser) Start(url string) {
	fmt.Printf("Starting fake parser\n")
}

func NewFakeParser() *FakeParser {
	return &FakeParser{
		Storage: storage.NewMemoryStorage(),
	}
}

func (f *FakeParser) Parse(url string) {
	f.Storage.SetCurrentBlock(0)
}

func (f *FakeParser) GetCurrentBlock() int {
	return f.Storage.GetCurrentBlock()
}

func (f *FakeParser) Subscribe(address string) bool {
	f.Storage.Subscribe(address)
	return true
}

func (f *FakeParser) GetTransactions(address string) []shared.Transaction {
	return f.Storage.GetTransactions(address)
}

func (f *FakeParser) AddTransaction(address string, transaction shared.Transaction) {
	f.Storage.AddTransaction(address, transaction)
}
