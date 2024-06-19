package storage

import (
	"github.com/gspad/ethereum-parser/shared"
)

type Storage interface {
	GetCurrentBlock() int
	SetCurrentBlock(block int)
	Subscribe(address string)
	GetSubscribedAddresses() map[string]bool
	GetTransactions(address string) []shared.Transaction
	AddTransaction(address string, transaction shared.Transaction)
}
