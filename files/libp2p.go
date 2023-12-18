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

func ReadData(rw *bufio.ReadWriter, db *leveldb.DB, stopChan <-chan struct{}) {
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
            err = json.Unmarshal([]byte(str), &block)
            if err != nil {
                log.Printf("Error al deserializar el bloque: %v\n", err)
                continue
            }

            if err != nil { // Si hay un error, asume que es un string simple y lo imprime
                fmt.Printf("String recibido: %s\n", str)
            
                if str == "1" { // Verifica si el string recibido es "1"
                    fmt.Println("Recibido '1', terminando ReadData...")
                    return // Termina la ejecución de la función (y por lo tanto de la goroutine)
                }
                
            } else { // Si no hay error, procesa el Block
                mutex.Lock()
            
                // Imprimir detalles del bloque
                fmt.Printf("Block:\n")
                fmt.Printf("Index: %d\n", block.Index)
                fmt.Printf("Timestamp: %d\n", block.Timestamp)
                fmt.Printf("PreviousHash: %s\n", block.PreviousHash)
                fmt.Printf("Hash: %s\n", block.Hash)
                fmt.Printf("Transactions:\n")
                for _, tx := range block.Transactions {
                    fmt.Printf("\tSender: %s\n", tx.Sender)
                    fmt.Printf("\tRecipient: %s\n", tx.Recipient)
                    fmt.Printf("\tAmount: %.2f\n", tx.Amount)
                    fmt.Printf("\tNonce: %d\n", tx.Nonce)
                    // Si deseas imprimir la firma, necesitarás convertirla a un formato legible
                    // Por ejemplo, convertirla a base64 o similar
                }
                fmt.Println("---------------------------")
            
                // Actualizar la cadena de bloques con el nuevo bloque
                err := UpdateBlockChain(db, block)
                if err != nil {
                    log.Printf("Error al actualizar la cadena de bloques con el nuevo bloque: %v\n", err)
                }
            
                mutex.Unlock()
            }
        }
    }
}


func WriteData(rw *bufio.ReadWriter, db *leveldb.DB, stopChan chan<- struct{}) {
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
            Menu(db,rw) // Llama a la función Menu
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
