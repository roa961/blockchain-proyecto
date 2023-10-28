package files

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/syndtr/goleveldb/leveldb"

	"encoding/json"
)

func SaveBlock(db *leveldb.DB, block Block) error {
	blockData, err := json.Marshal(block)
	if err != nil {
		return err
	}
	err = db.Put([]byte(fmt.Sprintf("block-%d", block.Index)), blockData, nil)
	if err != nil {
		return err
	}
	return nil
}

func LoadBlock(db *leveldb.DB, index int) (*Block, error) {
	blockData, err := db.Get([]byte(fmt.Sprintf("block-%d", index)), nil)
	if err != nil {
		return nil, err
	}

	var block Block
	err = json.Unmarshal(blockData, &block)
	if err != nil {
		return nil, err

	}
	return &block, nil
}

func CalculateHash(b Block) string {
	data := fmt.Sprintf("%d%d%s", b.Index, b.Timestamp, b.PreviousHash)
	for _, tx := range b.Transactions {
		data += fmt.Sprintf("%s%s%f%d", tx.Sender, tx.Recipient, tx.Amount, tx.Nonce)
	}
	h := sha256.New()
	h.Write([]byte(data))
	return hex.EncodeToString(h.Sum(nil))
}

func GenerateBlock(index int, previousHash string, transactions []Transaction) Block {
	block := Block{
		Index:        index,
		Timestamp:    time.Now().Unix(),
		Transactions: transactions,
		PreviousHash: previousHash,
	}
	block.Hash = CalculateHash(block)

	return block
}