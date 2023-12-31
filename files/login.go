package files

import (
	"bufio"
	"crypto/ecdsa"
	"fmt"

	"github.com/syndtr/goleveldb/leveldb"

	// "os"
	// "github.com/tkanos/gonfig"
	"encoding/json"

	// "reflect"
	"log"
	"math/big"
)

type Account struct {
	PublicKey  *ecdsa.PublicKey
	PrivateKey *ecdsa.PrivateKey
	Mnemonic   string
	Name       string
	Amount     float64

	Transactions float64
}
type Curve struct {
	P       *big.Int `json:"P"`
	N       *big.Int `json:"N"`
	B       *big.Int `json:"B"`
	Gx      *big.Int `json:"Gx"`
	Gy      *big.Int `json:"Gy"`
	BitSize int      `json:"BitSize"`
	Name    string   `json:"Name"`
}

type PublicKey struct {
	Curve Curve    `json:"Curve"`
	X     *big.Int `json:"X"`
	Y     *big.Int `json:"Y"`
}

type PrivateKey struct {
	Curve Curve    `json:"Curve"`
	X     *big.Int `json:"X"`
	Y     *big.Int `json:"Y"`
	D     *big.Int `json:"D"`
}

type Result struct {
	PublicKey    PublicKey  `json:"PublicKey"`
	PrivateKey   PrivateKey `json:"PrivateKey"`
	Mnemonic     string     `json:"Mnemonic"`
	Name         string     `json:"Name"`
	Amount       int        `json:"Amount"`
	Transactions int        `json:"Transactions"`
}

func printResult(result Result) {
	fmt.Println("Campo Name:", result.Name)
	fmt.Println("Campo Mnemonic:", result.Mnemonic)
	fmt.Println("Campo Amount:", result.Amount)
	fmt.Println("Campo Transactions:", result.Transactions)

	fmt.Println("Clave Pública:")
	fmt.Printf("  X: %d\n", result.PublicKey.X)
	fmt.Printf("  Y: %d\n", result.PublicKey.Y)
	fmt.Printf("  Curve:\n")
	fmt.Printf("    P: %d\n", result.PublicKey.Curve.P)
	fmt.Printf("    N: %d\n", result.PublicKey.Curve.N)
	fmt.Printf("    B: %d\n", result.PublicKey.Curve.B)
	fmt.Printf("    Gx: %d\n", result.PublicKey.Curve.Gx)
	fmt.Printf("    Gy: %d\n", result.PublicKey.Curve.Gy)
	fmt.Printf("    BitSize: %d\n", result.PublicKey.Curve.BitSize)
	fmt.Printf("    Name: %s\n", result.PublicKey.Curve.Name)

	fmt.Println("Clave Privada:")
	fmt.Printf("  X: %d\n", result.PrivateKey.X)
	fmt.Printf("  Y: %d\n", result.PrivateKey.Y)
	fmt.Printf("  D: %d\n", result.PrivateKey.D)
	fmt.Printf("  Curve:\n")
	fmt.Printf("    P: %d\n", result.PrivateKey.Curve.P)
	fmt.Printf("    N: %d\n", result.PrivateKey.Curve.N)
	fmt.Printf("    B: %d\n", result.PrivateKey.Curve.B)
	fmt.Printf("    Gx: %d\n", result.PrivateKey.Curve.Gx)
	fmt.Printf("    Gy: %d\n", result.PrivateKey.Curve.Gy)
	fmt.Printf("    BitSize: %d\n", result.PrivateKey.Curve.BitSize)
	fmt.Printf("    Name: %s\n", result.PrivateKey.Curve.Name)
}
func GetAccount(db *leveldb.DB, name string) (int, string, string, PublicKey, PrivateKey, error) {
	data, err := db.Get([]byte(name), nil)
	jsonString := string(data)
	var result Result
	err2 := json.Unmarshal([]byte(jsonString), &result)

	if err2 != nil {
		fmt.Println("Error al deserializar JSON:", err)
	}

	if err != nil {
		log.Fatal(err)
	}

	return result.Amount, result.Name, result.Mnemonic, result.PublicKey, result.PrivateKey, nil
}

func Login(db *leveldb.DB, rw *bufio.ReadWriter) (int, string, string, PublicKey, PrivateKey, error) {
	fmt.Println("HORA DE IDENTIFICARSE")
	var name string
	fmt.Println("1. Crear cuenta")
	fmt.Println("2. Ingresar nombre de cuenta para identificarse")
	fmt.Print("Seleccione una opción (1 o 2): ")
	var emptyPubK PublicKey
	var emptyPrivK PrivateKey
	for {
		var option int
		fmt.Scan(&option)

		switch option {
		case 1:
			// 	fmt.Println("CREAR CUENTA")
			fmt.Print("Ingrese su nombre: ")
			fmt.Scan(&name)
			nombre, err := db.Get([]byte(name), nil)
			if nombre != nil {
				fmt.Println("Usuario ya existente")
				return GetAccount(db, name)
			}
			privKey, pubKey, mnemonic, err := GenerateKeys(name)

			if err != nil {
				fmt.Println("Error:", err)
			} else {
				fmt.Println("Usuario " + name + " creado con éxito\n")

			}
			account := Account{
				PublicKey:    pubKey,
				PrivateKey:   privKey,
				Mnemonic:     mnemonic,
				Name:         name,
				Amount:       1000,
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

			mutex.Lock()
			defer mutex.Unlock()

			rw.WriteString(fmt.Sprintf("%s\n", string(data)))
			rw.Flush()

			return int(account.Amount), account.Name, account.Mnemonic, emptyPubK, emptyPrivK, nil

		case 2:
			fmt.Println("INDIQUE SU NOMBRE")
			fmt.Print("Ingrese su nombre de cuenta: ")
			fmt.Scan(&name)

			return GetAccount(db, name)

		default:
			fmt.Println("Opción no válida. Por favor, seleccione 1 o 2.")
		}
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
