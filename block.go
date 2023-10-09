package main

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"time"
    "os"
    "path"
    "path/filepath"
	"github.com/syndtr/goleveldb/leveldb"
    "runtime"
	"strings"

	"github.com/tkanos/gonfig"
	//"github.com/syndtr/goleveldb/leveldb/util"
	"encoding/json"
	"log"
	
	//"math/rand"
)

type Configuration struct {
    DBPath          string `json:"DB_PATH"`
    DBCachePath     string `json:"DB_CACHE_PATH"`
    RootSender      string `json:"ROOT_SENDER"`
    RootRecipient   string `json:"ROOT_RECIPIENT"`
    RootAmount      float64 `json:"ROOT_AMOUNT"`
    RootNonce       int `json:"ROOT_NONCE"`
}

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
	fmt.Printf("Hash previo: %s\n", block.PreviousHash)
	fmt.Println("Transactions:")
	for i, tx := range block.Transactions {
		fmt.Printf("  Transacción %d:\n", i+1)
		fmt.Printf("    Sender: %s\n", tx.Sender)
		fmt.Printf("    Recipient: %s\n", tx.Recipient)
		fmt.Printf("    Amount: %.2f\n", tx.Amount)
		fmt.Printf("    Nonce: %d\n", tx.Nonce)

	}
}

func getNewNonce(block Block) int {
	var actNonce int
	for i, tx := range block.Transactions {
		actNonce = tx.Nonce + i
	}

	return actNonce
}

func PrintBlockChain(db *leveldb.DB){
    iter := db.NewIterator(nil, nil)
    for iter.Next() {
        // Remember that the contents of the returned slice should not be modified, and
        // only valid until the next call to Next.
        
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
            fmt.Println() // Línea en blanco para separar las transacciones
        }
            
        
    
        
        fmt.Printf("PreviousHash: %s\n", block.PreviousHash)
        fmt.Printf("Hash: %s\n", block.Hash)
        fmt.Println("------------------------------------------")
        
    }
    
}

func getFileName() string {
	env := os.Getenv("ENV")
	if len(env) == 0 {
		env = "development"
	}
	filename := []string{"config/", "config.", env, ".json"}
	_, dirname, _, _ := runtime.Caller(0)
	filePath := path.Join(filepath.Dir(dirname), strings.Join(filename, ""))

	return filePath
}

func main() {
    configuration := Configuration{}
	err := gonfig.GetConf(getFileName(), &configuration)
	if err != nil {
		fmt.Println(err)
		os.Exit(500)
	}

	fmt.Println(configuration.RootSender)
	

	transactions := []Transaction{
		
		{
            Sender:    configuration.RootSender,
            Recipient: configuration.RootRecipient,
            Amount:    configuration.RootAmount,
            Nonce:     configuration.RootNonce,
        },
	}
	dbPath := configuration.DBPath
	dbPath_cache := configuration.DBCachePath
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
	defer dbCache.Close()
	//boque raiz
	// for i := 1; i <= 20; i++ {
	// 	block := generateBlock(i, "", transactions)
	// 	if err := saveBlock(db, block); err != nil {
	// 		log.Fatal(err)
	// 	}
	// 	fmt.Printf("Bloque %d generado y guardado.\n", i)
	// }

	//Comprueba si la db esta vacía para crear el bloque raiz

	iter_check := db.NewIterator(nil, nil)
	defer iter_check.Release()
	empty := !iter_check.Next()
	if empty {
		//Bloque raiz con índice 1
		block := generateBlock(1, "", transactions)
		if err := saveBlock(db, block); err != nil {
			log.Fatal(err)
		}
		if err := saveBlock(dbCache, block); err != nil {
			log.Fatal(err)
		}

	}

	for {
		// Mostrar el menú
		fmt.Println("----------MENÚ-BLOCKCHAIN----------")
		fmt.Println("Menú:")
		fmt.Println("1. Hacer una transaccion")
		fmt.Println("2. Leer transacción")
        fmt.Println("3. Mostrar cadena de bloques")
		fmt.Println("4. Salir")
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

			// Iterador para buscar el valor de previous hash
			iter_cache := dbCache.NewIterator(nil, nil)
			var prev_hash string
			var key_cache []byte
			var value []byte
			var block Block

			iter_cache.Next()
			value = iter_cache.Value()
			key_cache = iter_cache.Key()
			//fmt.Printf("Valor del iterador con la funcion LAST: %s\n", value)

			if err := json.Unmarshal(value, &block); err != nil {
				log.Fatalf("Error al deserializar el bloque: %v", err)
			}
			fmt.Println("este es el bloque sacado desde leveldb:")

			nonce := block.Transactions[0].Nonce
			//fmt.Print(nonce)
			transaction := []Transaction{
				{
					Sender:    sender,
					Recipient: recipient,
					Amount:    amount,
					Nonce:     nonce + 1,
				},
			}

			prev_hash = block.Hash
			fmt.Print(prev_hash, "\n")
			iter_cache.Release()
			if iter_cache.Error() != nil {
				log.Fatalf("Error al iterar a través de LevelDB: %v", iter_cache.Error())
			}
			nextIndex := block.Index + 1
			block = generateBlock(nextIndex, prev_hash, transaction)
			//fmt.Printf("%s", key_cache)
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
            PrintBlockChain(db)
        case 4:
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
