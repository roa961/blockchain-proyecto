package files
//in this file we do some verifications with the database also read/write operations

import (
	
	// "bufio"
	// "context"
	"encoding/json"
	//"flag"
	"fmt"
	"log"
	//"os"
	//"github.com/tkanos/gonfig"
	// net "github.com/libp2p/go-libp2p/core/network"
	// peer "github.com/libp2p/go-libp2p/core/peer"
	// pstore "github.com/libp2p/go-libp2p/core/peerstore"

	// ma "github.com/multiformats/go-multiaddr"
	"github.com/syndtr/goleveldb/leveldb"
    
	//"github.com/syndtr/goleveldb/leveldb/storage"
	"errors"
	//"github.com/tkanos/gonfig"

//	"reflect"
)

func ExistAccount(dbAccounts *leveldb.DB, recipient string) bool {
    _, err := dbAccounts.Get([]byte(recipient), nil)

    if err != nil {
        // Si el error es leveldb.ErrNotFound, la cuenta no existe
        if errors.Is(err, leveldb.ErrNotFound) {
            return false
        }
        // Manejo de otros tipos de errores si es necesario
        // Por ejemplo, podrías querer registrar estos errores o tratarlos de manera dferente
        return false
    }

    // Si no hay error, la cuenta existe
    return true
}
func SetNewAmount(dbAccounts *leveldb.DB, amount float64, accountName string) float64 {
    // Obtener la cuenta desde la base de datos
    accountData, err := dbAccounts.Get([]byte(accountName), nil)
    if err != nil {
        log.Printf("Error al obtener la cuenta: %v\n", err)
        return -1
    }
	
	//data, err := db.Get([]byte(name), nil)
	jsonString := string(accountData)
	
	var result Result
	err2 := json.Unmarshal([]byte(jsonString), &result)
	if err2 != nil {
		fmt.Println("Error al deserializar JSON:", err)
	}
    fmt.Printf("Saldo Original: %d\n", result.Amount)
		if err != nil {
			log.Fatal(err)
		}
    // Realizar la operación en el monto
    result.Amount -= int(amount)

    
    // Verificar si el nuevo monto es válido
    if result.Amount < 0 {
        log.Println("El monto resultante no puede ser negativo")
        return -1
    }

    // Serializar la cuenta actualizada
    updatedAccountData, err := json.Marshal(result)
    if err != nil {
        log.Printf("Error al serializar la cuenta actualizada: %v\n", err)
        return -1
    }

    // Guardar la cuenta actualizada en la base de datos
    if err := dbAccounts.Put([]byte(accountName), updatedAccountData, nil); err != nil {
        log.Printf("Error al guardar la cuenta actualizada: %v\n", err)
        return -1
    }

    return float64(result.Amount)
}
func UpdateBlockChain(db *leveldb.DB, blocks []Block) error {
    // Crear un iterador para la base de datos
    iter := db.NewIterator(nil, nil)
    for iter.Next() {
        // Eliminar cada clave encontrada
        err := db.Delete(iter.Key(), nil)
        if err != nil {
            iter.Release()
            return fmt.Errorf("error al eliminar la clave existente: %v", err)
        }
    }
    iter.Release()

    // Verificar si hubo errores durante la iteración
    if err := iter.Error(); err != nil {
        return fmt.Errorf("error durante la iteración: %v", err)
    }

    // Ahora escribe los nuevos bloques
    for _, block := range blocks {
        // Serializar el bloque a JSON
        blockData, err := json.Marshal(block)
        if err != nil {
            return fmt.Errorf("error al serializar el bloque %d: %v", block.Index, err)
        }

        // Usar el índice del bloque como clave para almacenarlo en LevelDB
        blockKey := fmt.Sprintf("%d", block.Index)

        // Guardar el bloque serializado en LevelDB
        err = db.Put([]byte(blockKey), blockData, nil)
        if err != nil {
            return fmt.Errorf("error al guardar el bloque %d en LevelDB: %v", block.Index, err)
        }
    }

    return nil
}