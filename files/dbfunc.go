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
	//	"reflect
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
func findAvailableIndex(db *leveldb.DB, startIndex int) (int, error) {
	for {
		indexKey := []byte(fmt.Sprintf("%d", startIndex))
		_, err := db.Get(indexKey, nil)
		if err == leveldb.ErrNotFound {
			return startIndex, nil
		} else if err != nil {
			return 0, err
		}
		// Si el índice ya está en uso, incrementamos y probamos de nuevo
		startIndex++
	}
}

func UpdateBlockChain(db, dbCache *leveldb.DB, block Block) error {
	// Serializar el bloque a JSON

	availableIndex, err := findAvailableIndex(db, block.Index)
	if err != nil {
		return err
	}
	block.Index = availableIndex
	block.Transactions[0].Nonce = availableIndex

	// Usar el índice del bloque como clave para almacenarlo en la base de datos principal (db)
	blockKey := fmt.Sprintf("%d", block.Index)

	blockData, err := json.Marshal(block)
	if err != nil {
		return fmt.Errorf("error al serializar el bloque %d: %v", block.Index, err)
	}
	err = db.Put([]byte(blockKey), blockData, nil)
	if err != nil {
		return fmt.Errorf("error al guardar el bloque %d en LevelDB principal: %v", block.Index, err)
	}

	// Usar una clave única con prefijo para almacenar el bloque en la base de datos cache (dbCache)
	cacheKey := fmt.Sprintf("block-%d", block.Index)
	err = dbCache.Put([]byte(cacheKey), blockData, nil)
	if err != nil {
		return fmt.Errorf("error al guardar el bloque %d en dbCache: %v", block.Index, err)
	}

	return nil
}

func PrintAllAccounts(dbAccounts *leveldb.DB) {
	iter := dbAccounts.NewIterator(nil, nil)
	defer iter.Release()

	for iter.Next() {
		key := iter.Key()
		value := iter.Value()

		var result Result
		err := json.Unmarshal(value, &result)
		if err != nil {
			log.Printf("Error al deserializar la cuenta para la clave %v: %v\n", key, err)
			continue
		}

		// Usa %d para enteros en lugar de %f
		fmt.Printf("Nombre: %v, Monto: %v\n", result.Name, result.Amount)
	}

	if err := iter.Error(); err != nil {
		log.Printf("Error durante la iteración: %v\n", err)
	}
}
func ResetAccounts(dbAccounts *leveldb.DB) {
	iter := dbAccounts.NewIterator(nil, nil)
	for iter.Next() {
		key := iter.Key()
		value := iter.Value()
		fmt.Printf("Clave: %s, Valor: %s\n", key, value)
		dbAccounts.Delete([]byte(key), nil)
	}
}
func ResetBlockChain(db *leveldb.DB) {
	iter := db.NewIterator(nil, nil)
	for iter.Next() {
		key := iter.Key()
		value := iter.Value()
		fmt.Printf("Clave: %s, Valor: %s\n", key, value)
		db.Delete([]byte(key), nil)
	}
}

func AddAmountToAccount(dbAccounts *leveldb.DB, amountToAdd float64, accountName string) float64 {
	// Obtener la cuenta desde la base de datos
	accountData, err := dbAccounts.Get([]byte(accountName), nil)
	if err != nil {
		log.Printf("Error al obtener la cuenta: %v\n", err)
		return -1
	}

	// Deserializar la cuenta
	var result Result
	err = json.Unmarshal(accountData, &result)
	if err != nil {
		log.Printf("Error al deserializar JSON: %v\n", err)
		return -1
	}
	// Sumar el monto a la cuenta
	result.Amount += int(amountToAdd)

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
