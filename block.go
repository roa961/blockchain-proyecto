package main

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/syndtr/goleveldb/leveldb"

	//"github.com/syndtr/goleveldb/leveldb/util"
	"encoding/json"
	"log"
	"os"
	//"math/rand"
)

type Transaction struct {
	Sender    string
	Recipient string
	Amount    float64
	Nonce     int
}

type Block struct {
	Index        int
	Timestamp    int64
	Transactions []Transaction
	PreviousHash string
	Hash         string
}

func saveBlock(db *leveldb.DB, block Block) error {
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

func loadBlock(db *leveldb.DB, index int) (*Block, error) {
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

func calculateHash(b Block) string {
	data := fmt.Sprintf("%d%d%s", b.Index, b.Timestamp, b.PreviousHash)
	for _, tx := range b.Transactions {
		data += fmt.Sprintf("%s%s%f%d", tx.Sender, tx.Recipient, tx.Amount, tx.Nonce)
	}
	h := sha256.New()
	h.Write([]byte(data))
	return hex.EncodeToString(h.Sum(nil))
}

func generateBlock(index int, previousHash string, transactions []Transaction) Block {
	block := Block{
		Index:        index,
		Timestamp:    time.Now().Unix(),
		Transactions: transactions,
		PreviousHash: previousHash,
	}
	block.Hash = calculateHash(block)

	return block
}
func PrintBlockData(block Block) {
	fmt.Println("Contenido del bloque:")
	fmt.Printf("Index: %d\n", block.Index)
	fmt.Printf("Timestamp: %d\n", block.Timestamp)
	fmt.Printf("Hash: %s\n", block.Hash)
	fmt.Println("Transactions:")
	for i, tx := range block.Transactions {
		fmt.Printf("  Transacción %d:\n", i+1)
		fmt.Printf("    Sender: %s\n", tx.Sender)
		fmt.Printf("    Recipient: %s\n", tx.Recipient)
		fmt.Printf("    Amount: %.2f\n", tx.Amount)
		fmt.Printf("    Nonce: %d\n", tx.Nonce)

	}
}

func main() {
	transactions := []Transaction{
		// Ejemplo 1
		{
			Sender:    "Alice",
			Recipient: "Bob",
			Amount:    25.0,
			Nonce:     1,
		},
	}
	dbPath := "./leveldb/"
	dbPath_cache := "./leveldb/cache"
	// Abrir la base de datos (creará una si no existe)
	db, err := leveldb.OpenFile(dbPath, nil)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()
	dbCache, err := leveldb.OpenFile(dbPath_cache, nil)
	if err != nil {
		log.Fatal(err)
	}

	//boque raiz
	// for i := 1; i <= 20; i++ {
	// 	block := generateBlock(i, "", transactions)
	// 	if err := saveBlock(db, block); err != nil {
	// 		log.Fatal(err)
	// 	}
	// 	fmt.Printf("Bloque %d generado y guardado.\n", i)
	// }

	block := generateBlock(1, "", transactions)
	if err := saveBlock(db, block); err != nil {
		log.Fatal(err)
	}
	if err := saveBlock(dbCache, block); err != nil {
		log.Fatal(err)
	}

	nonce := 1

	for {
		// Mostrar el menú
		fmt.Println("----------MENÚ-BLOCKCHAIN----------")
		fmt.Println("Menú:")
		fmt.Println("1. Hacer una transaccion")
		fmt.Println("2. Leer transacción")
		fmt.Println("3. Salir")
		fmt.Println("-----------------------------------")
		// Leer la opción del usuario
		var opcion int
		fmt.Print("Seleccione una opción: ")
		_, err := fmt.Scan(&opcion)
		if err != nil {
			fmt.Println("Error al leer la opción:", err)
			continue
		}

		// Procesar la opción seleccionada
		switch opcion {
		case 1:
			fmt.Println("---INICIO--TRANSACCION---")
			var sender, recipient string
			var amount float64

			fmt.Print("Ingrese el remitente: ")
			fmt.Scan(&sender)

			fmt.Print("Ingrese el destinatario: ")
			fmt.Scan(&recipient)

			fmt.Print("Ingrese el monto: ")
			fmt.Scan(&amount)

			// Aquí puedes establecer el valor del nonce como 1, ya que no se solicita al usuario.

			transaction := []Transaction{
				{
					Sender:    sender,
					Recipient: recipient,
					Amount:    amount,
					Nonce:     nonce,
				},
			}
			// Iterador para buscar el valor de nonce
			iter_cache := dbCache.NewIterator(nil, nil)
			var prev_hash string
			var key_cache []byte
			var value []byte
			var block Block
			if iter_cache.Last() {
				value = iter_cache.Value()
				key_cache = iter_cache.Key()
				fmt.Printf("Valor del iterador con la funcion LAST: %s\n", value)

			}
			if err := json.Unmarshal(value, &block); err != nil {
				log.Fatalf("Error al deserializar el bloque: %v", err)
			}
			fmt.Println("este es el bloque sacado desde leveldb:")

			prev_hash = block.Hash
			fmt.Print(prev_hash, "\n")

			iter_cache.Release()
			if iter_cache.Error() != nil {
				log.Fatalf("Error al iterar a través de LevelDB: %v", iter_cache.Error())
			}
			nextIndex := block.Index + 1
			block = generateBlock(nextIndex, prev_hash, transaction)
			fmt.Printf("%s", key_cache)
			err := dbCache.Delete(key_cache, nil)
			if err != nil {
				log.Printf("Error deleting key %s: %v", key_cache, err)
			}

			if err := saveBlock(db, block); err != nil {
				log.Fatal(err)
			}

			if err := saveBlock(dbCache, block); err != nil {
				log.Fatal(err)
			}
			fmt.Printf("Bloque generado y guardado.\n")

			fmt.Println("---FIN--TRANSACCION---")

		case 2:
			var blockNumber int
			fmt.Print("Ingrese el número del bloque que leer: ")
			fmt.Scan(&blockNumber)

			// Carga el bloque desde la base de datos
			block, err := loadBlock(db, blockNumber)
			if err != nil {
				log.Printf("Error al cargar el bloque: %v", err)
			} else {
				fmt.Println("Bloque cargado desde la base de datos.")
				PrintBlockData(*block)
			}
			// Imprime los datos del bloque

		case 3:
			fmt.Println("Saliendo del programa.")
			defer dbCache.Close()
			os.Exit(0) // Salir del programa
		default:
			fmt.Println("Opción no válida. Inténtalo de nuevo.")
		}
	}

}

//como buscar el bloque anterior para hashearlo y entregarlo al nuevo bloque? se busca por indice?
//como crear el bloque raiz
