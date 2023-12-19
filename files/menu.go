package files

import (
	"bufio"
	"encoding/json"

	"fmt"
	"log"

	"github.com/syndtr/goleveldb/leveldb"
)

func Menu(db *leveldb.DB, dbAccounts *leveldb.DB, dbCache *leveldb.DB, rw *bufio.ReadWriter) {

	Amount, Name, _, _, _, err := Login(dbAccounts, rw)
	if err != nil {
		fmt.Printf("Error durante el login: %s\n", err)
		return // O manejar el error como sea apropiado
	}

	fmt.Printf("Información del usuario:\n")
	fmt.Printf("Usuario: %s\n", Name)
	fmt.Printf("Cantidad de tokens: %d\n", Amount)

	for {
		// Mostrar el menú
		fmt.Println("----------MENÚ-BLOCKCHAIN----------")
		fmt.Println("Menú:")
		fmt.Println("1. Hacer una transaccion")
		fmt.Println("2. Mostrar cadena de bloques")
		fmt.Println("3. Mostrar estado de cuentas")
		fmt.Println("4. Salir")
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
			finalAmountDestiny := AddAmountToAccount(dbAccounts, montoTransferir, recipient)
			fmt.Printf("%v quedó con %v unidades\n", recipient, finalAmountDestiny)
			transaction := []Transaction{
				{
					Sender:    Name,
					Recipient: recipient,
					Amount:    montoTransferir,
					Nonce:     nonce + 1,
				},
			}

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

			if err := UpdateBlockChain(db, dbCache, block); err != nil {
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

		case 2:
			PrintBlockChain(db)
		case 3:
			PrintAllAccounts(dbAccounts)
		case 4:
			fmt.Println("Saliendo del programa.")
			//defer dbCache.Close()
			return
		default:
			fmt.Println("Opción no válida. Inténtalo de nuevo.")
		}

	}
}
