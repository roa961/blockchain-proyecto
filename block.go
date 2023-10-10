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

	"github.com/tkanos/gonfig"
	//"github.com/syndtr/goleveldb/leveldb/util"
	"encoding/json"
	"log"
    "crypto/ecdsa"
	"crypto/x509"
	"encoding/pem"
    
	//"math/rand"
	"bytes"
	
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

type UserData struct {
	Mnemonic  string
	PrivateKey []byte
	PublicKey  []byte
}

type Block struct {
	Index        int
	Timestamp    int64
	Transactions []Transaction
	PreviousHash string
	Hash         string
}

func bytesToECDSAPublicKey(publicKeyBytes []byte) (*ecdsa.PublicKey, error) {
    // Parsear el bloque PEM
    block, _ := pem.Decode(publicKeyBytes)
    if block == nil {
        return nil, fmt.Errorf("No se pudo decodificar el bloque PEM")
    }

    // Convertir los bytes de la clave pública en un *ecdsa.PublicKey
    publicKey, err := x509.ParsePKIXPublicKey(block.Bytes)
    if err != nil {
        return nil, err
    }

    // Verificar que el tipo sea *ecdsa.PublicKey
    ecdsaPublicKey, ok := publicKey.(*ecdsa.PublicKey)
    if !ok {
        return nil, fmt.Errorf("La clave no es un ECDSA PublicKey")
    }

    return ecdsaPublicKey, nil
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
			fmt.Printf("Signature: %d\n", transaction.Signature)
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
func obtenerHashTransaccion(transaccion *Transaction) []byte {
	data := fmt.Sprintf("%s%s%f%d", transaccion.Sender, transaccion.Recipient, transaccion.Amount, transaccion.Nonce)
	h := sha256.New()
	h.Write([]byte(data))
	return h.Sum(nil)
}

func generarClaves(usuario string, db *leveldb.DB) (*ecdsa.PrivateKey, *ecdsa.PublicKey, string, error) {

    

	entropy, _ := bip39.NewEntropy(128)
	mnemonic, _ := bip39.NewMnemonic(entropy)
	seed := bip39.NewSeed(mnemonic, "")
	seedReader := bytes.NewReader(seed)
	privKey, _ := ecdsa.GenerateKey(elliptic.P256(), seedReader)
	pubKey := &privKey.PublicKey
    privKeyBytes := privKey.D.Bytes()
    pubKeyBytes := elliptic.Marshal(pubKey.Curve, pubKey.X, pubKey.Y)
    userData := UserData{
		Mnemonic:  mnemonic,
		PrivateKey: privKeyBytes,
		PublicKey:  pubKeyBytes,
       
	}
    userDataJson, err := json.Marshal(userData)
    db.Put([]byte(usuario),userDataJson,nil)
	// db.Put([]byte(usuario+"_mnemonic"), []byte(mnemonic), nil)
	

	// privKeyBytes := privKey.D.Bytes()
	// db.Put([]byte(usuario+"_priv"), privKeyBytes, nil)
	
	// pubKeyBytes := elliptic.Marshal(pubKey.Curve, pubKey.X, pubKey.Y)
	// db.Put([]byte(usuario+"_pub"), pubKeyBytes, nil)

	
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
    // fmt.Println(privKey)
    // fmt.Println(mnemonicStr)
    // fmt.Println(pubKey)
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


    dbkeys, err := leveldb.OpenFile("./leveldb/keys", nil)
	defer dbkeys.Close()


	transactions := []Transaction{

		{
			Sender:    configuration.RootSender,
			Recipient: configuration.RootRecipient,
			Amount:    configuration.RootAmount,
			Nonce:     configuration.RootNonce,
		},
	}
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

	//HARCODE DE USUARIO
    usuario1 := "Bob"
    usuario2 := "Alice"
    var privKey1 *ecdsa.PrivateKey
    var pubKey1 *ecdsa.PublicKey
    var mnemonic1 string
    var privKey2 *ecdsa.PrivateKey
    var pubKey2 *ecdsa.PublicKey
    var mnemonic2 string
    //var userData1 UserData
    var userData2 UserData


    iter_keys_xd :=dbkeys.NewIterator(nil, nil)
    iter_keys := dbkeys.NewIterator(nil, nil)
    if (iter_keys_xd.Next()){
        
        for iter_keys.Next() {
            // Remember that the contents of the returned slice should not be modified, and
            // only valid until the next call to Next.
    
            key := iter_keys.Key()
            string_value := string(key)
            if(string_value=="Alice"){
                
                value := iter_keys.Value()
                if err := json.Unmarshal([]byte(value), &userData2); err != nil {
                    fmt.Printf("Error al deserializar el bloque: %v\n", err)
                    return
                }

                mnemonic2= userData2.Mnemonic
                fmt.Println(userData2.PublicKey)
                publicKey, err := bytesToECDSAPublicKey(userData2.PublicKey)
                if err != nil {
                    log.Fatal(err)
                }
                fmt.Println(publicKey)

                
                
            }
            

        }
    }
    privKey1, pubKey1, mnemonic1, err = generarClaves(usuario1,dbkeys)
        if err != nil {
            log.Fatal(err)
        }
        if mnemonic1 == "" {
            fmt.Println("No se encontró el mnemónico.")
        }

        
        privKey2, pubKey2, mnemonic2,err = generarClaves(usuario2,dbkeys)
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

		// Procesar la opción seleccionada
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
			fmt.Print(prev_hash, "\n")
			iter_cache.Release()
			if iter_cache.Error() != nil {
				log.Fatalf("Error al iterar a través de LevelDB: %v", iter_cache.Error())
			}
			nextIndex := block.Index + 1
			block = generateBlock(nextIndex, prev_hash, transaction)
			//fmt.Printf("%s", key_cache)
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
			os.Exit(0) // Salir del programa

		default:
			fmt.Println("Opción no válida. Inténtalo de nuevo.")
		}

	}

}

//como buscar el bloque anterior para hashearlo y entregarlo al nuevo bloque? se busca por indice?
//como crear el bloque raiz
