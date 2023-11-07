package files

import (
	"fmt"

	"github.com/syndtr/goleveldb/leveldb"

	"encoding/json"
)

type Configuration struct {
	DBPath        string  `json:"DB_PATH"`
	DBCachePath   string  `json:"DB_CACHE_PATH"`
	RootSender    string  `json:"ROOT_SENDER"`
	RootRecipient string  `json:"ROOT_RECIPIENT"`
	RootAmount    float64 `json:"ROOT_AMOUNT"`
	RootNonce     int     `json:"ROOT_NONCE"`
	DBAccountsPath string `json:"DB_ACCOUNTS_PATH"`
}

type Transaction struct {
	Sender    string
	Recipient string
	Amount    float64
	Nonce     int
	Signature []byte
}

type Block struct {
	Index        int
	Timestamp    int64
	Transactions []Transaction
	PreviousHash string
	Hash         string
}


func PrintBlockData(block Block) {
	fmt.Println("Contenido del bloque:")
	fmt.Printf("Index: %d\n", block.Index)
	fmt.Printf("Timestamp: %d\n", block.Timestamp)
	fmt.Printf("Hash: %s\n", block.Hash)
	fmt.Printf("Hash previo: %s\n", block.PreviousHash)
	fmt.Println("Transactions:")
	for i, tx := range block.Transactions {
		fmt.Printf("  Transacción %d:\n", i+1)
		fmt.Printf("    Sender: %s\n", tx.Sender)
		fmt.Printf("    Recipient: %s\n", tx.Recipient)
		fmt.Printf("    Amount: %.2f\n", tx.Amount)
		fmt.Printf("    Nonce: %d\n", tx.Nonce)
		fmt.Printf("    Signature: %d\n", tx.Signature)

	}
}

func PrintBlockChain(db *leveldb.DB) {
	iter := db.NewIterator(nil, nil)
	for iter.Next() {
		value := iter.Value()
		var block Block
		if err := json.Unmarshal([]byte(value), &block); err != nil {
			fmt.Printf("Error al deserializar el bloque: %v\n", err)
			return
		}
		fmt.Printf("Index: %d\n", block.Index)
		fmt.Printf("Timestamp: %d\n", block.Timestamp)
		fmt.Println()

		for _, transaction := range block.Transactions {
			fmt.Printf("Sender: %s\n", transaction.Sender)
			fmt.Printf("Recipient: %s\n", transaction.Recipient)
			fmt.Printf("Amount: %.2f\n", transaction.Amount)
			fmt.Printf("Nonce: %d\n", transaction.Nonce)
			fmt.Printf("Signature: %d\n", transaction.Signature)
			fmt.Println()
		}

		fmt.Printf("PreviousHash: %s\n", block.PreviousHash)
		fmt.Printf("Hash: %s\n", block.Hash)
		fmt.Println("------------------------------------------")

	}

}


