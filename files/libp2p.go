package files

import (
	"bufio"
	"context"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"sync"

	"github.com/libp2p/go-libp2p-core/crypto"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/syndtr/goleveldb/leveldb"
)

var mutex = &sync.Mutex{}

func ReadData(db *leveldb.DB) {
	println("AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA")

	jsonPersona := GetBlock(db)
	fmt.Printf("jsonPersona: %s\n", jsonPersona)

	for {
		str, err := rw.ReadString('\n')
		if err != nil {
			log.Fatal(err)
		}

		if str == "" {
			return
		}
		if str != "\n" {

			chain := make([]Block, 0)
			if err := json.Unmarshal([]byte(str), &chain); err != nil {
				log.Fatal(err)
			}

			mutex.Lock()
			if len(chain) > len(jsonPersona) {
				jsonPersona = chain
				bytes, err := json.MarshalIndent(jsonPersona, "", "  ")
				if err != nil {

					log.Fatal(err)
				}
				// Green console color: 	\x1b[32m
				// Reset console color: 	\x1b[0m
				fmt.Printf("\x1b[32m%s\x1b[0m> ", string(bytes))
			}
			mutex.Unlock()
		}
	}
}
func writeData(rw *bufio.ReadWriter) {
	print("xdd")
	// go func() {
	// 	for {
	// 		time.Sleep(5 * time.Second)
	// 		mutex.Lock()
	// 		bytes, err := json.Marshal(Blockchain)
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

	// stdReader := bufio.NewReader(os.Stdin)

	// for {
	// 	fmt.Print("> ")
	// 	sendData, err := stdReader.ReadString('\n')
	// 	if err != nil {
	// 		log.Fatal(err)
	// 	}

	// 	sendData = strings.Replace(sendData, "\n", "", -1)
	// 	bpm, err := strconv.Atoi(sendData)
	// 	if err != nil {
	// 		log.Fatal(err)
	// 	}
	// 	newBlock := GenerateBlock(Blockchain[len(Blockchain)-1], bpm)

	// 	if isBlockValid(newBlock, Blockchain[len(Blockchain)-1]) {
	// 		mutex.Lock()
	// 		Blockchain = append(Blockchain, newBlock)
	// 		mutex.Unlock()
	// 	}

	// 	bytes, err := json.Marshal(Blockchain)
	// 	if err != nil {
	// 		log.Println(err)
	// 	}

	// 	spew.Dump(Blockchain)

	// 	mutex.Lock()
	// 	rw.WriteString(fmt.Sprintf("%s\n", string(bytes)))
	// 	rw.Flush()
	// 	mutex.Unlock()
	// }

}
func handleStream(s net.Stream) {

	log.Println("Got a new stream!")

	// Create a buffer stream for non blocking read and write.
	rw := bufio.NewReadWriter(bufio.NewReader(s), bufio.NewWriter(s))

	go readData(rw)
	go writeData(rw)

	// stream 's' will stay open until you close it (or the other side closes it).
}
func makeBasicHost(listenPort int, secio bool, randseed int64) (host.Host, error) {

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

	if !secio {
		opts = append(opts, libp2p.NoEncryption())
	}

	basicHost, err := libp2p.New(context.Background(), opts...)
	if err != nil {
		return nil, err
	}

	// Build host multiaddress
	hostAddr, _ := ma.NewMultiaddr(fmt.Sprintf("/ipfs/%s", basicHost.ID().Pretty()))

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
