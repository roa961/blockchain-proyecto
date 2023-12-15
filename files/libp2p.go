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
	println("AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA")

	jsonPersona := GetBlock(db)
	//fmt.Printf("jsonPersona: %s\n", jsonPersona)

	for {
		str, err := rw.ReadString('\n')
		if err != nil {
			log.Fatal(err)
		}
	
		// Imprime lo que se leyó, independientemente de su contenido
		fmt.Printf("Received: %s\n", str)
	
		if str == "" {
			fmt.Printf("HKJDAHJKFDFAKJHAFDNKJFDSANJAFDKSNKSADF\n")
			// Si quieres terminar el bucle cuando str es vacío, descomenta la siguiente línea
			// return
		}
	
		if str != " " {
			fmt.Print("cualquier wea")
			chain := make([]Block, 0)
			if err := json.Unmarshal([]byte(str), &chain); err != nil {
				log.Fatal(err)
			}
			fmt.Print("cualquier wea 2")
			mutex.Lock()
			fmt.Print("cualquier wea 3")
			//cuidado con este if
			//if len(chain) < len(jsonPersona) {
				fmt.Print("cualquier wea 4")
				jsonPersona = chain
				bytes, err := json.MarshalIndent(jsonPersona, "", "  ")
				if err != nil {
					log.Fatal(err)
				}
	
				// Imprime los bytes en color verde
				fmt.Printf("\x1b[32m%s\x1b[0m> ", string(bytes))
			//}
			mutex.Unlock()
			fmt.Print("cualquier wea 5")
		}
		
	}
}
func WriteData(rw *bufio.ReadWriter, db *leveldb.DB) {
	print("xdd")
	jsonPersona := GetBlock(db)
	// go func() {
	// 	for {
	// 		time.Sleep(5 * time.Second)
	// 		mutex.Lock()
	// 		bytes, err := json.Marshal(jsonPersona)
	// 		if err != nil {
	// 			log.Println(err)
	// 		}
	// 		mutex.Unlock()

	// 		mutex.Lock()
	// 		rw.WriteString(fmt.Sprintf("%s\n", string(bytes)))
	// 		rw.Flush()
	// 		mutex.Unlock()

	// 	}
	// }()

	stdReader := bufio.NewReader(os.Stdin)

	for {
		fmt.Print("exisde exisde exisde")
		fmt.Print("> ")
		sendData, err := stdReader.ReadString('\n')
		if err != nil {
			log.Fatal(err)
		}

		sendData = strings.Replace(sendData, "\n", "", -1)
		// bpm, err := strconv.Atoi(sendData)
		// transaction := Transaction{
		// 	Sender:    "senderAddress",
		// 	Recipient: "recipientAddress",
		// 	Amount:    100.0,
		// 	Nonce:     1,
		// 	Signature: []byte("yourSignature"), // Reemplaza esto con una firma real si es necesario
		// }
		// transactions := []Transaction{transaction}
		// dataToSend := GenerateBlock(2,"9a27d61ded67313a0085903c46d0600e76e6e3b80d60ad97b447cba63efb000f",transactions)
		// PrintBlockData(dataToSend)
		// if isBlockValid(newBlock, jsonPersona[len(jsonPersona)-1]) {
		// 	mutex.Lock()
		// 	Blockchain = append(Blockchain, newBlock)
		// 	mutex.Unlock()
		// }

		// bytes, err := json.Marshal(jsonPersona)
		// if err != nil {
		// 	log.Println(err)
		// }

		//spew.Dump(jsonPersona)
		mutex.Lock()
			bytes, err := json.Marshal(jsonPersona)
			if err != nil {
				log.Println(err)
			}
			mutex.Unlock()

			mutex.Lock()
			rw.WriteString(fmt.Sprintf("%s\n", string(bytes)))
			rw.Flush()
			mutex.Unlock()

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
