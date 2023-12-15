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

func ReadData(rw *bufio.ReadWriter, db *leveldb.DB) {
    jsonPersona := GetBlock(db)

    for {
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

        chain := make([]Block, 0)
        err = json.Unmarshal([]byte(str), &chain)

        if err != nil { // Si hay un error, asume que es un string simple y lo imprime
            fmt.Printf("String recibido: %s\n", str)
        } else { // Si no hay error, asume que es un tipo Block y procede como antes
            mutex.Lock()
           // if len(chain) > len(jsonPersona) {
                jsonPersona = chain
                bytes, err := json.MarshalIndent(jsonPersona, "", "  ")
                if err != nil {
                    log.Fatal(err)
                }
                fmt.Printf("\x1b[32m%s\x1b[0m> ", string(bytes))
          //  }
            mutex.Unlock()
        }
    }
}

func WriteData(rw *bufio.ReadWriter, db *leveldb.DB) {
    stdReader := bufio.NewReader(os.Stdin)

    for {
        fmt.Println("Selecciona una opción:")
        fmt.Println("1. Enviar bloque")
        fmt.Println("2. Enviar mensaje personalizado")
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

        default:
            fmt.Println("Opción no válida. Por favor, elige 1 o 2.")
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
