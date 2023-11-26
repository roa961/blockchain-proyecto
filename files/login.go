package files

import (
	"crypto/ecdsa"
	"fmt"

	"github.com/syndtr/goleveldb/leveldb"

	// "os"
	// "github.com/tkanos/gonfig"
	"encoding/json"
	"log"
)

type Account struct {
	PublicKey    *ecdsa.PublicKey
	PrivateKey   *ecdsa.PrivateKey
	Mnemonic     string
	Name         string
	Amount       float64
	Transactions float64
}

func Login(db *leveldb.DB) {
	fmt.Println("HORA DE IDENTIFICARSE")
	var name string
	fmt.Println("1. Crear cuenta")
	fmt.Println("2. Ingresar nombre de cuenta para identificarse")
	fmt.Print("Seleccione una opción (1 o 2): ")
	var option int
	fmt.Scanln(&option)

	switch option {
	case 1:
		fmt.Println("CREAR CUENTA")
		fmt.Print("Ingrese su nombre: ")
		fmt.Scanln(&name)
		privKey, pubKey, mnemonic, err := GenerateKeys(name)
		if err != nil {
			fmt.Println("Error:", err)
		} else {
			fmt.Println("Resultado de GenerarClaves para usuario1:")
			fmt.Println("Private Key:", privKey)
			fmt.Println("Public Key:", pubKey)
			fmt.Println("Mnemonic:", mnemonic)
		}
		account := Account{
			PublicKey:    pubKey,
			PrivateKey:   privKey,
			Mnemonic:     mnemonic,
			Name:         name,
			Amount:       0,
			Transactions: 0,
		}

		data, err := json.Marshal(account)
		if err != nil {
			log.Fatal(err)
		}

		err = db.Put([]byte(account.Name), data, nil)
		if err != nil {
			log.Fatal(err)
		}

	case 2:
		fmt.Println("IDENTIFICARSE")
		fmt.Print("Ingrese su nombre de cuenta: ")
		fmt.Scanln(&name)
		fmt.Printf("Se ha identificado como %s", name)
	default:
		fmt.Println("Opción no válida")
	}
}

//	func saveAccount(db *leveldb.DB, account Account) error {
//		data, err := json.Marshal(account)
//		if err != nil {
//			return err
//		}
//		err = db.Put([]byte(username), data, nil)
//		return err
//	}
func ShowAllData(db *leveldb.DB) {
	iter := db.NewIterator(nil, nil)
	for iter.Next() {
		key := iter.Key()
		value := iter.Value()
		fmt.Printf("Clave: %s, Valor: %s\n", key, value)
	}
	iter.Release()
	err := iter.Error()
	if err != nil {
		log.Fatal(err)
	}
}
