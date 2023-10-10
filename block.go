package main

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/syndtr/goleveldb/leveldb"

	"encoding/json"
	"log"

	"github.com/tkanos/gonfig"

	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"math/big"

	"github.com/tyler-smith/go-bip39"
)

type Configuration struct {
	DBPath        string  `json:"DB_PATH"`
	DBCachePath   string  `json:"DB_CACHE_PATH"`
	RootSender    string  `json:"ROOT_SENDER"`
	RootRecipient string  `json:"ROOT_RECIPIENT"`
	RootAmount    float64 `json:"ROOT_AMOUNT"`
	RootNonce     int     `json:"ROOT_NONCE"`
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
func obtenerHashTransaccion(transaccion *Transaction) []byte {
	data := fmt.Sprintf("%s%s%f%d", transaccion.Sender, transaccion.Recipient, transaccion.Amount, transaccion.Nonce)
	h := sha256.New()
	h.Write([]byte(data))
	return h.Sum(nil)
}

func generarClaves(usuario string) (*ecdsa.PrivateKey, *ecdsa.PublicKey, string, error) {

	db, err := leveldb.OpenFile("./leveldb/keys", nil)
	if err != nil {
		return nil, nil, "", err
	}
	defer db.Close()

	entropy, _ := bip39.NewEntropy(128)
	mnemonic, _ := bip39.NewMnemonic(entropy)
	seed := bip39.NewSeed(mnemonic, "")
	seedReader := bytes.NewReader(seed)
	privKey, _ := ecdsa.GenerateKey(elliptic.P256(), seedReader)
	pubKey := &privKey.PublicKey

	err = db.Put([]byte(usuario+"_mnemonic"), []byte(mnemonic), nil)
	if err != nil {
		return nil, nil, "", err
	}

	privKeyBytes := privKey.D.Bytes()
	err = db.Put([]byte(usuario+"_priv"), privKeyBytes, nil)
	if err != nil {
		return nil, nil, "", err
	}
	pubKeyBytes := elliptic.Marshal(pubKey.Curve, pubKey.X, pubKey.Y)
	err = db.Put([]byte(usuario+"_pub"), pubKeyBytes, nil)
	if err != nil {
		return nil, nil, "", err
	}
	mnemonicBytes, err := db.Get([]byte(usuario+"_mnemonic"), nil)
	if err != nil {
		return nil, nil, "", err
	}
	mnemonicStr := string(mnemonicBytes)

	privKeyBytes, err = db.Get([]byte(usuario+"_priv"), nil)
	if err != nil {
		return nil, nil, "", err
	}
	privKey.D.SetBytes(privKeyBytes)

	pubKeyBytes, err = db.Get([]byte(usuario+"_pub"), nil)
	if err != nil {
		log.Fatal(err)
	}
	pubKey.X, pubKey.Y = elliptic.Unmarshal(pubKey.Curve, pubKeyBytes)
	return privKey, pubKey, mnemonicStr, nil
}
func firmarTransaccion(privKey *ecdsa.PrivateKey, transaccion *Transaction) {
	hash := obtenerHashTransaccion(transaccion)
	r, s, err := ecdsa.Sign(rand.Reader, privKey, hash)
	if err != nil {
		log.Fatal(err)
	}
	signature := append(r.Bytes(), s.Bytes()...)
	transaccion.Signature = signature

}

func verificarFirma(pubKey *ecdsa.PublicKey, mensaje []byte, firma []byte) bool {
	r := new(big.Int)
	s := new(big.Int)
	r.SetBytes(firma[:len(firma)/2])
	s.SetBytes(firma[len(firma)/2:])
	return ecdsa.Verify(pubKey, mensaje, r, s)
}

func main() {
	configuration := Configuration{}
	err := gonfig.GetConf(getFileName(), &configuration)
	if err != nil {
		fmt.Println(err)
		os.Exit(500)
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
	transactions := []Transaction{

		{
			Sender:    configuration.RootSender,
			Recipient: configuration.RootRecipient,
			Amount:    configuration.RootAmount,
			Nonce:     configuration.RootNonce,
		},
	}

	iter_check := db.NewIterator(nil, nil)
	defer iter_check.Release()
	empty := !iter_check.Next()
	if empty {

		//Se crea el bloque raíz con índice 1, previous hash "" y una transacción con información contenida en el config file
		block := generateBlock(1, "", transactions)
		if err := saveBlock(db, block); err != nil {
			log.Fatal(err)
		}
		if err := saveBlock(dbCache, block); err != nil {
			log.Fatal(err)
		}

	}

	//Los usuarios se "Hardcodean" para mostrar el funcionamiento de las firmas
	usuario1 := "Bob"
	privKey1, pubKey1, mnemonic1, err := generarClaves(usuario1)
	if err != nil {
		log.Fatal(err)
	}
	if mnemonic1 == "" {
		fmt.Println("No se encontró el mnemónico.")
	}

	usuario2 := "Alice"
	privKey2, pubKey2, mnemonic2, err := generarClaves(usuario2)
	if err != nil {
		log.Fatal(err)
	}
	if mnemonic2 == "" {
		fmt.Println("No se encontró el mnemónico.")
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

		// Cases para cada una de las opciones
		switch opcion {
		case 1:
			fmt.Println("---INICIO--TRANSACCION---")
			var remitente, destinatario int
			var amount float64
			var sender, recipient string

			fmt.Println("Ingrese el remitente")
			fmt.Println("1 Bob")
			fmt.Println("2 Alice")
			_, err := fmt.Scan(&remitente)
			if err != nil {
				fmt.Println("Error al leer el remitente:", err)
				continue
			}

			switch remitente {
			case 1:
				sender = usuario1
				fmt.Println("Ingrese el destinatario:")
				fmt.Println("1 Alice")
				fmt.Scan(&destinatario)

				if destinatario == 1 {
					recipient = usuario2
					fmt.Println("Ingrese el monto: ")
					fmt.Scan(&amount)

				} else {
					fmt.Println("Error al leer destinatario")
					continue

				}
			case 2:
				sender = usuario2
				fmt.Println("Ingrese el destinatario:")
				fmt.Println("1 Bob")
				fmt.Scan(&destinatario)
				if destinatario == 1 {
					recipient = usuario1
					fmt.Println("Ingrese el monto: ")
					fmt.Scan(&amount)

				} else {
					fmt.Println("Error al leer destinatario")
					continue
				}
			}

			// Iterador para buscar el valor de previous hash dentro de la base de datos cache
			iter_cache := dbCache.NewIterator(nil, nil)
			var prev_hash string
			var key_cache []byte
			var value []byte
			var block Block

			iter_cache.Next()
			value = iter_cache.Value()
			key_cache = iter_cache.Key()

			if err := json.Unmarshal(value, &block); err != nil {
				log.Fatalf("Error al deserializar el bloque: %v", err)
			}

			nonce := block.Transactions[0].Nonce

			transaction := []Transaction{
				{
					Sender:    sender,
					Recipient: recipient,
					Amount:    amount,
					Nonce:     nonce + 1,
				},
			}
			if destinatario == 1 {
				firmarTransaccion(privKey1, &transaction[0])
				esValida := verificarFirma(pubKey1, obtenerHashTransaccion(&transaction[0]), transaction[0].Signature)
				if esValida {
					fmt.Println("La firma es válida y fue firmado por Bob.")
				} else {
					fmt.Println("La firma es inválida.")
				}

			} else if destinatario == 2 {
				firmarTransaccion(privKey2, &transaction[0])
				esValida := verificarFirma(pubKey2, obtenerHashTransaccion(&transaction[0]), transaction[0].Signature)
				if esValida {
					fmt.Println("La firma es válida y fue firmado por Alice.")
				} else {
					fmt.Println("La firma es inválida.")
				}
			}

			prev_hash = block.Hash

			iter_cache.Release()
			if iter_cache.Error() != nil {
				log.Fatalf("Error al iterar a través de LevelDB: %v", iter_cache.Error())
			}
			nextIndex := block.Index + 1
			block = generateBlock(nextIndex, prev_hash, transaction)

			err = dbCache.Delete(key_cache, nil)
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
				bloque := *block
				PrintBlockData(bloque)
				trans := &bloque.Transactions[0]
				validacion1 := verificarFirma(pubKey1, obtenerHashTransaccion(trans), bloque.Transactions[0].Signature)
				validacion2 := verificarFirma(pubKey2, obtenerHashTransaccion(trans), bloque.Transactions[0].Signature)
				if validacion1 {
					fmt.Println("La firma es válida y fue firmado por Bob.")
				} else if validacion2 {
					fmt.Println("La firma es válida y fue firmado por Alice.")
				} else {
					fmt.Println("La firma es inválida.")
				}
			}
			// Imprime los datos del bloque

		case 3:
			PrintBlockChain(db)
		case 4:
			fmt.Println("Saliendo del programa.")
			defer dbCache.Close()
			os.Exit(0)
		default:
			fmt.Println("Opción no válida. Inténtalo de nuevo.")
		}

	}

}
