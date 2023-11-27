package files

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"fmt"
	"log"
	"math/big"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/syndtr/goleveldb/leveldb"
	"github.com/tyler-smith/go-bip39"
)

func GetFileName() string {
	env := os.Getenv("ENV")
	if len(env) == 0 {
		env = "development"
	}
	filename := []string{"config/", "config.", env, ".json"}
	_, dirname, _, _ := runtime.Caller(0)
	filePath := path.Join(filepath.Dir(dirname), strings.Join(filename, ""))

	return filePath
}

func GetHashTransaction(transaction *Transaction) []byte {
	data := fmt.Sprintf("%s%s%f%d", transaction.Sender, transaction.Recipient, transaction.Amount, transaction.Nonce)
	h := sha256.New()
	h.Write([]byte(data))
	return h.Sum(nil)
}

func GenerateKeys(user string) (*ecdsa.PrivateKey, *ecdsa.PublicKey, string, error) {

	db, err := leveldb.OpenFile("./leveldb/keys", nil)
	if err != nil {
		return nil, nil, "", err
	}
	defer db.Close()

	entropy, _ := bip39.NewEntropy(128)
	mnemonic, _ := bip39.NewMnemonic(entropy)
	seed := bip39.NewSeed(mnemonic, "")
	seedReader := bytes.NewReader(seed)
	privKey, _ := ecdsa.GenerateKey(elliptic.P256(), seedReader)
	pubKey := &privKey.PublicKey

	err = db.Put([]byte(user+"_mnemonic"), []byte(mnemonic), nil)
	if err != nil {
		return nil, nil, "", err
	}

	privKeyBytes := privKey.D.Bytes()
	err = db.Put([]byte(user+"_priv"), privKeyBytes, nil)
	if err != nil {
		return nil, nil, "", err
	}
	pubKeyBytes := elliptic.Marshal(pubKey.Curve, pubKey.X, pubKey.Y)
	err = db.Put([]byte(user+"_pub"), pubKeyBytes, nil)
	if err != nil {
		return nil, nil, "", err
	}
	mnemonicBytes, err := db.Get([]byte(user+"_mnemonic"), nil)
	if err != nil {
		return nil, nil, "", err
	}
	mnemonicStr := string(mnemonicBytes)

	privKeyBytes, err = db.Get([]byte(user+"_priv"), nil)
	if err != nil {
		return nil, nil, "", err
	}
	privKey.D.SetBytes(privKeyBytes)

	pubKeyBytes, err = db.Get([]byte(user+"_pub"), nil)
	if err != nil {
		log.Fatal(err)
	}
	pubKey.X, pubKey.Y = elliptic.Unmarshal(pubKey.Curve, pubKeyBytes)
	return privKey, pubKey, mnemonicStr, nil
}
func SignTransaction(privKey *ecdsa.PrivateKey, transaction *Transaction) {
	hash := GetHashTransaction(transaction)
	r, s, err := ecdsa.Sign(rand.Reader, privKey, hash)
	if err != nil {
		log.Fatal(err)
	}
	signature := append(r.Bytes(), s.Bytes()...)
	transaction.Signature = signature

}

func VerifySignature(pubKey *ecdsa.PublicKey, message []byte, sign []byte) bool {
	r := new(big.Int)
	s := new(big.Int)
	r.SetBytes(sign[:len(sign)/2])
	s.SetBytes(sign[len(sign)/2:])
	return ecdsa.Verify(pubKey, message, r, s)
}
