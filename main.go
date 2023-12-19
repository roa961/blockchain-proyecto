package main

import (
	"blockchain-proyecto/files"
	"bufio"
	"context"

	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"

	net "github.com/libp2p/go-libp2p/core/network"
	peer "github.com/libp2p/go-libp2p/core/peer"
	pstore "github.com/libp2p/go-libp2p/core/peerstore"

	ma "github.com/multiformats/go-multiaddr"
	"github.com/syndtr/goleveldb/leveldb"
	"github.com/tkanos/gonfig"
)

func main() {

	configuration := files.Configuration{}
	err := gonfig.GetConf(files.GetFileName(), &configuration)
	if err != nil {
		fmt.Println(err)
		os.Exit(500)
	}

	dbPath := configuration.DBPath
	dbPath_cache := configuration.DBCachePath
	dbPath_accounts := configuration.DBAccountsPath

	// // Abrir la base de datos (creará una si no existe)
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
	dbAccounts, err := leveldb.OpenFile(dbPath_accounts, nil)

	if err != nil {
		log.Fatal(err)
	}
	defer dbAccounts.Close()

	//Parse options from the command line
	listenF := flag.Int("l", 0, "wait for incoming connections")
	target := flag.String("d", "", "target peer to dial")
	secio := flag.Bool("secio", false, "enable secio")
	seed := flag.Int64("seed", 0, "set random seed for id generation")
	flag.Parse()
	if *listenF == 0 {
		log.Fatal("Please provide a port to bind on with -l")
	}

	ha, err := files.MakeBasicHost(*listenF, *secio, *seed)
	if err != nil {
		log.Fatal(err)
	}
	if *target == "" {
		log.Println("listening for connections")

		// Set a stream handler on host A. /p2p/1.0.0 is
		// a user-defined protocol name.

		ha.SetStreamHandler("/p2p/1.0.0", func(s net.Stream) {
			go HandleStream(s, db, dbAccounts, dbCache)
		})

		select {} // hang forever
		/**** This is where the listener code ends ****/
	} else {
		ipfsaddr, err := ma.NewMultiaddr(*target)
		if err != nil {
			log.Fatalln(err)
		}

		pid, err := ipfsaddr.ValueForProtocol(ma.P_IPFS)
		if err != nil {
			log.Fatalln(err)
		}

		peerid, err := peer.Decode(pid)
		if err != nil {
			log.Fatalln(err)
		}

		// Decapsulate the /ipfs/<peerID> part from the target
		// /ip4/<a.b.c.d>/ipfs/<peer> becomes /ip4/<a.b.c.d>
		targetPeerAddr, _ := ma.NewMultiaddr(
			fmt.Sprintf("/ipfs/%s", peerid.String()))
		targetAddr := ipfsaddr.Decapsulate(targetPeerAddr)

		// We have a peer ID and a targetAddr so we add it to the peerstore
		// so LibP2P knows how to contact it
		ha.Peerstore().AddAddr(peerid, targetAddr, pstore.PermanentAddrTTL)

		log.Println("opening stream")
		// make a new stream from host B to host A
		// it should be handled on host A by the handler we set above because
		// we use the same /p2p/1.0.0 protocol
		s, err := ha.NewStream(context.Background(), peerid, "/p2p/1.0.0")
		if err != nil {
			log.Fatalln(err)
		}
		// Create a buffered stream so that read and writes are non blocking.
		rw := bufio.NewReadWriter(bufio.NewReader(s), bufio.NewWriter(s))

		// Create a thread to read and write data.
		stopChan := make(chan struct{})
		go files.WriteData(rw, db, dbAccounts, dbCache, stopChan)
		go files.ReadData(rw, db, dbAccounts, dbCache, stopChan)

		select {} // hang forever

	}

}
func HandleStream(s net.Stream, db *leveldb.DB, dbAccounts *leveldb.DB, dbCache *leveldb.DB) {

	log.Println("Got a new stream!")

	configuration := files.Configuration{}
	err := gonfig.GetConf(files.GetFileName(), &configuration)
	if err != nil {
		fmt.Println(err)
	}

	transactions := []files.Transaction{
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
	stopChan := make(chan struct{})
	rw := bufio.NewReadWriter(bufio.NewReader(s), bufio.NewWriter(s))
	if empty {
		//Se crea el bloque raíz con índice 1, previous hash "" y una transacción con información contenida en el config file
		block := files.GenerateBlock(1, "", transactions)
		if err := files.UpdateBlockChain(db, dbCache, block); err != nil {
			log.Fatal(err)
		}
		// Create a buffer stream for non blocking read and write.

		jsonPersona := files.GetBlock(db)

		bytes, err := json.Marshal(jsonPersona)
		if err != nil {
			log.Println(err)
		}
		rw.WriteString(fmt.Sprintf("%s\n", string(bytes)))
		rw.Flush()
	}

	go files.ReadData(rw, db, dbAccounts, dbCache, stopChan)
	go files.WriteData(rw, db, dbAccounts, dbCache, stopChan)
	// stream 's' will stay open until you close it (or the other side closes it).
}
