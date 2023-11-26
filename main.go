package main

import (
	"blockchain-proyecto/files"
	"encoding/json"
	"fmt"
	"log"
	"os"

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
	dbAccounts, err := leveldb.OpenFile(dbPath_accounts, nil)

	if err != nil {
		log.Fatal(err)
	}
	defer dbAccounts.Close()

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
	if empty {

		//Se crea el bloque raíz con índice 1, previous hash "" y una transacción con información contenida en el config file
		block := files.GenerateBlock(1, "", transactions)
		if err := files.SaveBlock(db, block); err != nil {
			log.Fatal(err)
		}
		if err := files.SaveBlock(dbCache, block); err != nil {
			log.Fatal(err)
		}

	}

	files.Login(dbAccounts)
	files.ShowAllData(dbAccounts)

	//Los usuarios se "Hardcodean" para mostrar el funcionamiento de las firmas
	usuario1 := "Bob"
	privKey1, pubKey1, mnemonic1, err := files.GenerarClaves(usuario1)
	if err != nil {
		log.Fatal(err)
	}
	if mnemonic1 == "" {
		fmt.Println("No se encontró el mnemónico.")
	}

	usuario2 := "Alice"
	privKey2, pubKey2, mnemonic2, err := files.GenerarClaves(usuario2)
	//fmt.println("esta es la llave publica y pribada de alice")
	fmt.Println("Private Key:", privKey2)
	fmt.Println("Public Key:", pubKey2)

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
			var block files.Block

			iter_cache.Next()
			value = iter_cache.Value()
			key_cache = iter_cache.Key()

			if err := json.Unmarshal(value, &block); err != nil {
				log.Fatalf("Error al deserializar el bloque: %v", err)
			}

			nonce := block.Transactions[0].Nonce

			transaction := []files.Transaction{
				{
					Sender:    sender,
					Recipient: recipient,
					Amount:    amount,
					Nonce:     nonce + 1,
				},
			}
			if destinatario == 1 {
				files.FirmarTransaccion(privKey1, &transaction[0])
				esValida := files.VerificarFirma(pubKey1, files.GetHashTransaction(&transaction[0]), transaction[0].Signature)
				if esValida {
					fmt.Println("La firma es válida y fue firmado por Bob.")
				} else {
					fmt.Println("La firma es inválida.")
				}

			} else if destinatario == 2 {
				files.FirmarTransaccion(privKey2, &transaction[0])
				esValida := files.VerificarFirma(pubKey2, files.GetHashTransaction(&transaction[0]), transaction[0].Signature)
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
			block = files.GenerateBlock(nextIndex, prev_hash, transaction)

			err = dbCache.Delete(key_cache, nil)
			if err != nil {
				log.Printf("Error deleting key %s: %v", key_cache, err)
			}

			if err := files.SaveBlock(db, block); err != nil {
				log.Fatal(err)
			}

			if err := files.SaveBlock(dbCache, block); err != nil {
				log.Fatal(err)
			}
			fmt.Printf("Bloque generado y guardado.\n")

			fmt.Println("---FIN--TRANSACCION---")

		case 2:
			var blockNumber int
			fmt.Print("Ingrese el número del bloque que leer: ")
			fmt.Scan(&blockNumber)

			// Carga el bloque desde la base de datos
			block, err := files.LoadBlock(db, blockNumber)
			if err != nil {
				log.Printf("Error al cargar el bloque: %v", err)
			} else {
				fmt.Println("Bloque cargado desde la base de datos.")
				bloque := *block
				files.PrintBlockData(bloque)
				trans := &bloque.Transactions[0]
				validacion1 := files.VerificarFirma(pubKey1, files.GetHashTransaction(trans), bloque.Transactions[0].Signature)
				validacion2 := files.VerificarFirma(pubKey2, files.GetHashTransaction(trans), bloque.Transactions[0].Signature)
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
			files.PrintBlockChain(db)
		case 4:
			fmt.Println("Saliendo del programa.")
			defer dbCache.Close()
			os.Exit(0)
		default:
			fmt.Println("Opción no válida. Inténtalo de nuevo.")
		}

	}

}
