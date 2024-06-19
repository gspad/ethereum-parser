package parser

import "github.com/gspad/ethereum-parser/shared"

type FakeParser struct {
	CurrentBlock int
	Subscribed   map[string]bool
	Transactions map[string][]shared.Transaction
}

func NewFakeParser() *FakeParser {
	return &FakeParser{
		Subscribed:   make(map[string]bool),
		Transactions: make(map[string][]shared.Transaction),
	}
}

func (f *FakeParser) GetCurrentBlock() int {
	return f.CurrentBlock
}

func (f *FakeParser) Subscribe(address string) bool {
	f.Subscribed[address] = true
	return true
}

func (f *FakeParser) GetTransactions(address string) []shared.Transaction {
	return f.Transactions[address]
}
