package files

import (
	"bufio"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"strings"
	"sync"

	//"time"

	mrand "math/rand"

	//"github.com/davecgh/go-spew/spew"
	libp2p "github.com/libp2p/go-libp2p"
	crypto "github.com/libp2p/go-libp2p/core/crypto"
	host "github.com/libp2p/go-libp2p/core/host"
	ma "github.com/multiformats/go-multiaddr"
	"github.com/syndtr/goleveldb/leveldb"
)

var mutex = &sync.Mutex{}
var allBlocks []Block

func ReadData(rw *bufio.ReadWriter, db *leveldb.DB, dbAccounts *leveldb.DB, dbCache *leveldb.DB, stopChan <-chan struct{}) {
	//jsonPersona := GetBlock(db)

	for {
		select {
		case <-stopChan: // Si stopChan se cierra, sale de la función
			fmt.Println("Deteniendo ReadData...")
			return

		default:
			// Leer del stream
			str, err := rw.ReadString('\n')
			if err != nil {
				log.Fatal(err)
			}
			// Quita espacios en blanco y saltos de línea
			str = strings.TrimSpace(str)

			// Imprime lo que se leyó, independientemente de su contenido
			fmt.Printf("Received: %s\n", str)

			if str == "" {
				fmt.Printf("Cadena vacía recibida\n")
				continue
			}

			// Intenta deserializar el string en un único Block
			var block Block
			errBlock := json.Unmarshal([]byte(str), &block)

			if errBlock != nil {
				fmt.Printf("String recibido: %s\n", str)

				if str == "1" { // Verifica si el string recibido es "1"
					fmt.Println("Recibido '1', terminando ReadData...")
					return // Termina la ejecución de la función (y por lo tanto de la goroutine)
				}

				if strings.HasPrefix(str, "Nombre: ") && strings.Contains(str, ", Monto: ") {
					// Extraer los datos de la cadena
					parts := strings.Split(str, ", Monto: ")
					nombrePart := parts[0]
					montoPart := parts[1]

					nombre := strings.TrimPrefix(nombrePart, "nombre: ")
					montoStr := strings.TrimSpace(montoPart)
					// amount, err := strconv.ParseFloat(montoStr, 64)
					// if err != nil {
					// 	log.Printf("Error al convertir el monto a número: %v\n", err)
					// 	return
					// }
					// account := Account{
					// 	Name:   nombre,
					// 	Amount: amount,
					// }

					mutex.Lock()
					dbAccounts.Put([]byte(nombre), []byte(montoStr), nil)
					PrintAllAccounts(dbAccounts)

					mutex.Unlock()
					// Convertir el monto a un número, si es necesario
					// monto, err := strconv.ParseFloat(montoStr, 64)
					// if err != nil {
					//     log.Printf("Error al convertir el monto a número: %v\n", err)
					//     return
					// }

					// Aquí puedes manejar los datos extraídos (nombre y monto)
					fmt.Printf("Nombre extraído: %s\n", nombre)
					fmt.Printf("Monto extraído: %s\n", montoStr)

					// Ejemplo: Guardar o procesar la información de nombre y monto
				}
				if strings.HasPrefix(str, "{\"PublicKey\":{\"Curve\":{") {
					var account Account
					err := json.Unmarshal([]byte(str), &account)
					if err != nil {
						log.Println(err)
					}
					data, err := json.Marshal(account)
					if err != nil {
						log.Println(err)
					}
					mutex.Lock()
					dbAccounts.Put([]byte(account.Name), data, nil)
					PrintAllAccounts(dbAccounts)
					mutex.Unlock()

				}

				if strings.HasPrefix(str, "[{\"Index\":1,\"Timestamp\"") {

					var blocks []Block
					err := json.Unmarshal([]byte(str), &blocks)
					if err != nil {
						log.Printf("Error al deserializar el JSON en blocks: %v\n", err)
						return
					}

					// Ahora puedes trabajar con el slice de blocks
					// Por ejemplo, imprimir el primer bloque si existe
					if len(blocks) > 0 {
						fmt.Printf("Primer bloque recibido: %+v\n", blocks[0])
						// Aquí puedes añadir más lógica para manejar los bloques
						UpdateBlockChain(db, dbCache, blocks[0])

					}
				}

			}
			if errBlock == nil {

				// Imprimir detalles del bloque
				fmt.Printf("Block:\n")
				fmt.Printf("Index: %d\n", block.Index)
				fmt.Printf("Timestamp: %d\n", block.Timestamp)
				fmt.Printf("PreviousHash: %s\n", block.PreviousHash)
				fmt.Printf("Hash: %s\n", block.Hash)
				fmt.Printf("Transactions:\n")
				var amountSpended float64
				var recipient string
				var sender string
				for _, tx := range block.Transactions {
					fmt.Printf("\tSender: %s\n", tx.Sender)
					sender = tx.Sender
					fmt.Printf("\tRecipient: %s\n", tx.Recipient)
					recipient = tx.Recipient
					fmt.Printf("\tAmount: %.2f\n", tx.Amount)
					amountSpended = tx.Amount
					fmt.Printf("\tNonce: %d\n", tx.Nonce)
				}
				finalAmountDestiny := AddAmountToAccount(dbAccounts, amountSpended, recipient)
				fmt.Printf("%s quedó con %d unidades\n", recipient, finalAmountDestiny)
				finalAmountOrigin := SetNewAmount(dbAccounts, amountSpended, sender)
				fmt.Printf("%s quedó con %d unidades\n", sender, finalAmountOrigin)
				fmt.Println("---------------------------")

				// Actualizar la cadena de bloques con el nuevo bloque
				err := UpdateBlockChain(db, dbCache, block)
				//err := SaveBlock(db, block)
				if err != nil {
					log.Printf("Error al actualizar la cadena de bloques con el nuevo bloque: %v\n", err)
				}

			}

		}
	}
}

func WriteData(rw *bufio.ReadWriter, db *leveldb.DB, dbAccounts *leveldb.DB, dbCache *leveldb.DB, stopChan chan<- struct{}) {
	stdReader := bufio.NewReader(os.Stdin)

	for {
		fmt.Println("Selecciona una opción:")
		fmt.Println("1. Enviar bloque")
		fmt.Println("2. Enviar mensaje ")
		fmt.Println("3. Mostrar menú")
		fmt.Println("4. Cerrar comunicación")
		fmt.Print("> ")

		option, err := stdReader.ReadString('\n')
		if err != nil {
			log.Fatal(err)
		}

		option = strings.TrimSpace(option) // Elimina espacios y saltos de línea

		switch option {
		case "1":
			jsonPersona := GetBlock(db)

			mutex.Lock()
			bytes, err := json.Marshal(jsonPersona)
			if err != nil {
				log.Println(err)
			}
			rw.WriteString(fmt.Sprintf("%s\n", string(bytes)))
			rw.Flush()
			mutex.Unlock()

		case "2":
			fmt.Print("Ingresa tu mensaje: ")
			message, err := stdReader.ReadString('\n')
			if err != nil {
				log.Fatal(err)
			}

			message = strings.TrimSpace(message) // Elimina espacios y saltos de línea

			mutex.Lock()
			rw.WriteString(fmt.Sprintf("%s\n", message))
			rw.Flush()
			mutex.Unlock()
		case "3":
			Menu(db, dbAccounts, dbCache, rw) // Llama a la función Menu
			// jsonPersona := GetBlock(db)

			//         mutex.Lock()
			//         bytes, err := json.Marshal(jsonPersona)
			//         if err != nil {
			//             log.Println(err)
			//         }
			//         rw.WriteString(fmt.Sprintf("%s\n", string(bytes)))
			//         rw.Flush()
			//         mutex.Unlock()
		case "4":
			fmt.Println("Cerrando comunicación...")
			stopChan <- struct{}{} // Enviar señal para cortar la comunicación
			return

		default:
			fmt.Println("Opción no válida. Por favor, elige 1, 2, 3 o 4.")
		}
	}
}

func MakeBasicHost(listenPort int, secio bool, randseed int64) (host.Host, error) {

	// If the seed is zero, use real cryptographic randomness. Otherwise, use a
	// deterministic randomness source to make generated keys stay the same
	// across multiple runs
	var r io.Reader
	if randseed == 0 {
		r = rand.Reader
	} else {
		r = mrand.New(mrand.NewSource(randseed))
	}

	// Generate a key pair for this host. We will use it
	// to obtain a valid host ID.
	priv, _, err := crypto.GenerateKeyPairWithReader(crypto.RSA, 2048, r)
	if err != nil {
		return nil, err
	}

	opts := []libp2p.Option{
		libp2p.ListenAddrStrings(fmt.Sprintf("/ip4/127.0.0.1/tcp/%d", listenPort)),
		libp2p.Identity(priv),
	}

	basicHost, err := libp2p.New(opts...)
	if err != nil {
		return nil, err
	}

	// Build host multiaddress
	hostAddr, _ := ma.NewMultiaddr(fmt.Sprintf("/ipfs/%s", basicHost.ID().String()))

	// Now we can build a full multiaddress to reach this host
	// by encapsulating both addresses:
	addr := basicHost.Addrs()[0]
	fullAddr := addr.Encapsulate(hostAddr)
	log.Printf("I am %s\n", fullAddr)
	if secio {
		log.Printf("Now run \"go run main.go -l %d -d %s -secio\" on a different terminal\n", listenPort+1, fullAddr)
	} else {
		log.Printf("Now run \"go run main.go -l %d -d %s\" on a different terminal\n", listenPort+1, fullAddr)
	}

	return basicHost, nil
}
