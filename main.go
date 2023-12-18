package main

import (
	"blockchain-proyecto/files"
	"bufio"
	"context"

	//"encoding/json"
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
	//"reflect"
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
	// files.ResetBlockChain(db)
	// files.ResetAccounts(dbAccounts)
	// transactions := []files.Transaction{

	// 	{
	// 		Sender:    configuration.RootSender,
	// 		Recipient: configuration.RootRecipient,
	// 		Amount:    configuration.RootAmount,
	// 		Nonce:     configuration.RootNonce,
	// 	},
	// }

	// iter_check := db.NewIterator(nil, nil)
	// defer iter_check.Release()
	// empty := !iter_check.Next()
	// if empty {

	// 	//Se crea el bloque raíz con índice 1, previous hash "" y una transacción con información contenida en el config file
	// 	block := files.GenerateBlock(1, "", transactions)
	// 	if err := files.SaveBlock(db, block); err != nil {
	// 		log.Fatal(err)
	// 	}
	// 	if err := files.SaveBlock(dbCache, block); err != nil {
	// 		log.Fatal(err)
	// 	}

	// }

	// Amount, Name, Mnemonic, PublicKey, PrivateKey, err := files.Login(dbAccounts)
	// if err != nil {
	// 	log.Fatal(err)
	// 	return
	// }
	//AQUI EMPIEZA LA COMUNICACION CON LIBP2P
	//for {

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
		// ha.SetStreamHandler("/p2p/1.0.0", func(s net.Stream) {
		// 	HandleStream(s, db)
		// })

		// The following code extracts target's peer ID from the
		// given multiaddress
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
	// files.ReadData(db)
	// Obtener el JSON de la persona
	//Imprimir el JSON
	//}
	// fmt.Printf("Resultado desde el main:\n")
	// fmt.Printf("Amount: %d\n", Amount)
	// fmt.Printf("Mnemonic: %s\n", Mnemonic)
	// fmt.Printf("name: %s\n", Name)
	// fmt.Printf("Public: %s\n", PublicKey)
	// fmt.Printf("Private: %s\n", PrivateKey)

	// for {
	// 	// Mostrar el menú
	// 	fmt.Println("----------MENÚ-BLOCKCHAIN----------")
	// 	fmt.Println("Menú:")
	// 	fmt.Println("1. Hacer una transaccion")
	// 	fmt.Println("2. Leer transacción")
	// 	fmt.Println("3. Mostrar cadena de bloques")
	// 	fmt.Println("4. Salir")
	// 	fmt.Println("-----------------------------------")
	// 	// Leer la opción del usuario
	// 	var option int
	// 	fmt.Print("Seleccione una opción: ")
	// 	_, err := fmt.Scan(&option)
	// 	if err != nil {
	// 		fmt.Println("Error al leer la opción:", err)
	// 		continue
	// 	}

	// 	// Cases para cada una de las opciones
	// 	switch option {
	// 	case 1:
	// 		fmt.Println("---INICIO--TRANSACCION---")

	// 		// Iterador para buscar el valor de previous hash dentro de la base de datos cache
	// 		iter_cache := dbCache.NewIterator(nil, nil)
	// 		var prev_hash string
	// 		var key_cache []byte
	// 		var value []byte
	// 		var block files.Block

	// 		iter_cache.Next()
	// 		value = iter_cache.Value()
	// 		key_cache = iter_cache.Key()

	// 		if err := json.Unmarshal(value, &block); err != nil {
	// 			log.Fatalf("Error al deserializar el bloque: %v", err)
	// 		}

	// 		nonce := block.Transactions[0].Nonce
	// 		fmt.Print("Ingrese el destinatario: ")
	// 		var recipient string
	// 		fmt.Scan(&recipient)

	// 		fmt.Print("Ingrese el monto a transferir: ")
	// 		var montoTransferir float64
	// 		_, err := fmt.Scan(&montoTransferir)
	// 		if err != nil {
	// 			fmt.Println("Error al leer el monto:", err)
	// 			return
	// 		}
	// 		Amount_float := float64(Amount)
	// 		// Verificar que el monto sea positivo y menor o igual al monto original
	// 		if montoTransferir <= 0 || montoTransferir > Amount_float {
	// 			fmt.Println("El monto ingresado no es válido.")
	// 			return
	// 		}

	// 		fmt.Printf("Amount: %d\n", Amount)
	// 		dataType := reflect.TypeOf(Amount)
	// 		fmt.Printf("Tipo: %v\n", dataType)

	// 		transaction := []files.Transaction{
	// 			{
	// 				Sender:    Name,
	// 				Recipient: recipient,
	// 				Amount:    montoTransferir,
	// 				Nonce:     nonce + 1,
	// 			},
	// 		}

	// 		fmt.Println(transaction)

	// 		//files.SignTransaction(PrivateKey,&transaction[0])
	// 		//ItIsValid := files.VerifySignature(PublicKey, files.GetHashTransaction(&transaction[0]), transaction[0].Signature)
	// 		//if ItIsValid {
	// 			//fmt.Println("La firma es válida y fue firmado por Bob.")
	// 		//} else {
	// 			//fmt.Println("La firma es inválida.")
	// 		//}

	// 		//if receiver == 1 {
	// 			//files.SignTransaction(privKey1, &transaction[0])
	// 			//ItIsValid := files.VerifySignature(pubKey1, files.GetHashTransaction(&transaction[0]), transaction[0].Signature)
	// 			//if ItIsValid {

	// 				//fmt.Println("La firma es válida y fue firmado por Bob.")
	// 			//} else {
	// 				//fmt.Println("La firma es inválida.")
	// 			//}

	// 		// if receiver == 1 {
	// 		// 	files.SignTransaction(privKey1, &transaction[0])
	// 		// 	ItIsValid := files.VerifySignature(pubKey1, files.GetHashTransaction(&transaction[0]), transaction[0].Signature)
	// 		// 	if ItIsValid {

	// 		// 		fmt.Println("La firma es válida y fue firmado por Bob.")
	// 		// 	} else {
	// 		// 		fmt.Println("La firma es inválida.")
	// 		// 	}

	// 		prev_hash = block.Hash

	// 		iter_cache.Release()
	// 		if iter_cache.Error() != nil {
	// 			log.Fatalf("Error al iterar a través de LevelDB: %v", iter_cache.Error())
	// 		}
	// 		nextIndex := block.Index + 1
	// 		block = files.GenerateBlock(nextIndex, prev_hash, transaction)

	// 		err = dbCache.Delete(key_cache, nil)
	// 		if err != nil {
	// 			log.Printf("Error deleting key %s: %v", key_cache, err)
	// 		}

	// 		if err := files.SaveBlock(db, block); err != nil {
	// 			log.Fatal(err)
	// 		}

	// 		if err := files.SaveBlock(dbCache, block); err != nil {
	// 			log.Fatal(err)
	// 		}
	// 		fmt.Printf("Bloque generado y guardado.\n")

	// 		fmt.Println("---FIN--TRANSACCION---")

	// 	//case 2:
	// 	//var blockNumber int
	// 	//fmt.Print("Ingrese el número del bloque que leer: ")
	// 	//fmt.Scan(&blockNumber)

	// 	//// Carga el bloque desde la base de datos
	// 	//block, err := files.LoadBlock(db, blockNumber)
	// 	//if err != nil {
	// 	//log.Printf("Error al cargar el bloque: %v", err)
	// 	//} else {
	// 	//fmt.Println("Bloque cargado desde la base de datos.")
	// 	//blockaux := *block
	// 	//files.PrintBlockData(blockaux)
	// 	//trans := &blockaux.Transactions[0]
	// 	//verify1 := files.VerifySignature(pubKey1, files.GetHashTransaction(trans), blockaux.Transactions[0].Signature)
	// 	//verify2 := files.VerifySignature(pubKey2, files.GetHashTransaction(trans), blockaux.Transactions[0].Signature)
	// 	//if verify1 {
	// 	//fmt.Println("La firma es válida y fue firmado por Bob.")
	// 	//} else if verify2 {
	// 	//	fmt.Println("La firma es válida y fue firmado por Alice.")
	// 	//} else {
	// 	//	fmt.Println("La firma es inválida.")
	// 	//}
	// 	//}
	// 	//// Imprime los datos del bloque

	// 	case 3:
	// 		files.PrintBlockChain(db)
	// 	case 4:
	// 		fmt.Println("Saliendo del programa.")
	// 		defer dbCache.Close()
	// 		os.Exit(0)
	// 	default:
	// 		fmt.Println("Opción no válida. Inténtalo de nuevo.")
	// 	}

	// }

}
func HandleStream(s net.Stream, db *leveldb.DB, dbAccounts *leveldb.DB, dbCache *leveldb.DB) {

	log.Println("Got a new stream!")

	// Create a buffer stream for non blocking read and write.
	rw := bufio.NewReadWriter(bufio.NewReader(s), bufio.NewWriter(s))
	stopChan := make(chan struct{})

	go files.ReadData(rw, db, dbAccounts, dbCache, stopChan)

	go files.WriteData(rw, db, dbAccounts, dbCache, stopChan)
	// log.Println("otro!")

	// stream 's' will stay open until you close it (or the other side closes it).
}
