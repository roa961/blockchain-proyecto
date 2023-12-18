package files

import (
	"bufio"
	// "context"
	"encoding/json"
	//"flag"
	"fmt"
	"log"

	"github.com/tkanos/gonfig"
	// net "github.com/libp2p/go-libp2p/core/network"
	// peer "github.com/libp2p/go-libp2p/core/peer"
	// pstore "github.com/libp2p/go-libp2p/core/peerstore"

	// ma "github.com/multiformats/go-multiaddr"
	"github.com/syndtr/goleveldb/leveldb"
	//"github.com/tkanos/gonfig"
	//"reflect"
)

func Menu(db *leveldb.DB, dbAccounts *leveldb.DB, dbCache *leveldb.DB, rw *bufio.ReadWriter) {
	configuration := Configuration{}
	err := gonfig.GetConf(GetFileName(), &configuration)
	if err != nil {
		fmt.Println(err)
		// os.Exit(500)
	}

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
		fmt.Printf("EMPTYYYYYYYYYY LA CONCHA DE TU HERMANAS")
		//Se crea el bloque raíz con índice 1, previous hash "" y una transacción con información contenida en el config file
		block := GenerateBlock(1, "", transactions)
		if err := UpdateBlockChain(db,dbCache, block); err != nil {
			log.Fatal(err)
		}
		// if err := SaveBlock(dbCache, block); err != nil {
		// 	log.Fatal(err)
		// }

	}else{
		fmt.Printf("EMPTYYYYYYYYYYNTTTTTTTTTTT ")
	}

	Amount, Name, Mnemonic, PublicKey, PrivateKey, err := Login(dbAccounts, rw)
	if err != nil {
		fmt.Printf("Error durante el login: %s\n", err)
		return // O manejar el error como sea apropiado
	}

	fmt.Printf("Resultado desde el main:\n")
	fmt.Printf("Amount: %d\n", Amount)
	fmt.Printf("Mnemonic: %s\n", Mnemonic)
	fmt.Printf("name: %s\n", Name)
	fmt.Printf("Public: %s\n", PublicKey)
	fmt.Printf("Private: %s\n", PrivateKey)

	for {
		// Mostrar el menú
		fmt.Println("----------MENÚ-BLOCKCHAIN----------")
		fmt.Println("Menú:")
		fmt.Println("1. Hacer una transaccion")
		fmt.Println("2. Leer transacción")
		fmt.Println("3. Mostrar cadena de bloques")
		fmt.Println("4. Mostra Datos de la cuenta")
		fmt.Println("5. Salir")
		fmt.Println("-----------------------------------")
		// Leer la opción del usuario
		var option int
		fmt.Print("Seleccione una opción: ")
		_, err := fmt.Scan(&option)
		if err != nil {
			fmt.Println("Error al leer la opción:", err)
			continue
		}

		// Cases para cada una de las opciones
		switch option {
		case 1:
			fmt.Println("---INICIO--TRANSACCION---")

			iter_cache := dbCache.NewIterator(nil, nil)
			var prev_hash string
			var key_cache []byte
			var value []byte
			var block Block
			
			PrintBlockChain(dbCache)
			iter_cache.Next()
			value = iter_cache.Value()
			key_cache = iter_cache.Key()

			if err := json.Unmarshal(value, &block); err != nil {
				log.Fatalf("Error al deserializar el bloque: %v", err)
			}

			fmt.Println("Monto disponible:", Amount)

			nonce := block.Transactions[0].Nonce
			fmt.Print("Ingrese el destinatario: ")
			var recipient string
			fmt.Scan(&recipient)
			//FUNCION VERIFICAR DESTINATARIO
			existAccount := ExistAccount(dbAccounts, recipient)

			if existAccount {
				fmt.Printf("La cuenta '%s' existe.\n", recipient)
			} else {
				fmt.Printf("La cuenta '%s' no existe.\n", recipient)
				continue
			}

			fmt.Print("Ingrese el monto a transferir: ")
			var montoTransferir float64
			_, err := fmt.Scan(&montoTransferir)
			if err != nil {
				fmt.Println("Error al leer el monto:", err)
				return
			}
			Amount_float := float64(Amount)
			// Verificar que el monto sea positivo y menor o igual al monto original
			if montoTransferir <= 0 || montoTransferir > Amount_float {
				fmt.Println("El monto ingresado no es válido.")
				return
			}
			finalAmount := SetNewAmount(dbAccounts, montoTransferir, Name)
			if finalAmount == -1 {
				fmt.Println("hubo un error en el asginamiento del saldo")
				return
			}

			fmt.Printf("Te quedaste con: %.2f\n", finalAmount) // Utiliza %.2f para dos decimales
			Amount = int(finalAmount)
			finalAmountDestiny := AddAmountToAccount(dbAccounts,montoTransferir,recipient)
			fmt.Printf("%s quedó con %d unidades\n", recipient, finalAmountDestiny)
			transaction := []Transaction{
				{
					Sender:    Name,
					Recipient: recipient,
					Amount:    montoTransferir,
					Nonce:     nonce + 1,
				},
			}

			//files.SignTransaction(PrivateKey,&transaction[0])
			//ItIsValid := files.VerifySignature(PublicKey, files.GetHashTransaction(&transaction[0]), transaction[0].Signature)
			//if ItIsValid {
			//fmt.Println("La firma es válida y fue firmado por Bob.")
			//} else {
			//fmt.Println("La firma es inválida.")
			//}

			//if receiver == 1 {
			//files.SignTransaction(privKey1, &transaction[0])
			//ItIsValid := files.VerifySignature(pubKey1, files.GetHashTransaction(&transaction[0]), transaction[0].Signature)
			//if ItIsValid {

			//fmt.Println("La firma es válida y fue firmado por Bob.")
			//} else {
			//fmt.Println("La firma es inválida.")
			//}

			// if receiver == 1 {
			// 	files.SignTransaction(privKey1, &transaction[0])
			// 	ItIsValid := files.VerifySignature(pubKey1, files.GetHashTransaction(&transaction[0]), transaction[0].Signature)
			// 	if ItIsValid {

			// 		fmt.Println("La firma es válida y fue firmado por Bob.")
			// 	} else {
			// 		fmt.Println("La firma es inválida.")
			// 	}

			prev_hash = block.Hash

			iter_cache.Release()
			if iter_cache.Error() != nil {
				log.Fatalf("Error al iterar a través de LevelDB: %v", iter_cache.Error())
			}
			nextIndex := block.Index + 1
			block = GenerateBlock(nextIndex, prev_hash, transaction)

			err = dbCache.Delete(key_cache, nil)
			if err != nil {
				log.Printf("Error deleting key %s: %v", key_cache, err)
			}

			if err := UpdateBlockChain(db, dbCache,block); err != nil {
				log.Fatal(err)
			}

			// if err := SaveBlock(dbCache, block); err != nil {
			// 	log.Fatal(err)
			// }
			fmt.Printf("Bloque generado y guardado.\n")

			mutex.Lock()
			bytes, err := json.Marshal(block)
			if err != nil {
				log.Println(err)
			}
			rw.WriteString(fmt.Sprintf("%s\n", string(bytes)))
			rw.Flush()
			mutex.Unlock()

			fmt.Println("---FIN--TRANSACCION---")

		//case 2:
		//var blockNumber int
		//fmt.Print("Ingrese el número del bloque que leer: ")
		//fmt.Scan(&blockNumber)

		//// Carga el bloque desde la base de datos
		//block, err := files.LoadBlock(db, blockNumber)
		//if err != nil {
		//log.Printf("Error al cargar el bloque: %v", err)
		//} else {
		//fmt.Println("Bloque cargado desde la base de datos.")
		//blockaux := *block
		//files.PrintBlockData(blockaux)
		//trans := &blockaux.Transactions[0]
		//verify1 := files.VerifySignature(pubKey1, files.GetHashTransaction(trans), blockaux.Transactions[0].Signature)
		//verify2 := files.VerifySignature(pubKey2, files.GetHashTransaction(trans), blockaux.Transactions[0].Signature)
		//if verify1 {
		//fmt.Println("La firma es válida y fue firmado por Bob.")
		//} else if verify2 {
		//	fmt.Println("La firma es válida y fue firmado por Alice.")
		//} else {
		//	fmt.Println("La firma es inválida.")
		//}
		//}
		//// Imprime los datos del bloque

		case 3:
			PrintBlockChain(db)
		case 4:
			
		case 5:
			fmt.Println("Saliendo del programa.")
			defer dbCache.Close() 
			return
		default:
			fmt.Println("Opción no válida. Inténtalo de nuevo.")
		}

	}
}
