type Block struct {
	Index int
	Timestamp int64
	Transactions []Transaction
	PreviousHash string
	Hash string
	Nonce int
}
func newBlock(index int, PreviousHash string, transactions []Transaction, nonce int) *Block {
	block := Block{
		Index: index,
		Timestamp: time.Now().Unix(),
		Transactions: transactions,
		PreviousHash: previousHash,
		Nonce: nonce,

	}
	//block.Hash = calculateHash(block)
	//agregar funcion calculo hash
	return block
}