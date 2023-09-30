package main
import (
    "time"
    "crypto/sha256"
    "encoding/hex"
    "fmt"
    "github.com/syndtr/goleveldb/leveldb"
    "log"
    "os"
    "encoding/json"
)
type Transaction struct {
    Sender string
    Recipient string
    Amount float64
}

type Block struct{
    Index int
    Timestamp int64
    Transactions []Transaction
    PreviousHash string
    Hash string
    Nonce int
    
}

func saveBlock(db *leveldb.DB, block Block) error{
    blockData, err:=json.Marshal(block)
    if err != nil {
        return err
    }
    err = db.Put([]byte(fmt.Sprintf("block-%d",block.Index)),blockData, nil)
    if err != nil {
        return err
    }
    return nil
}

func loadBlock(db *leveldb.DB, index int) (*Block, error) {
    blockData, err := db.Get([]byte(fmt.Sprintf("block-%d", index)), nil) 
    if err != nil {
        return nil, err
    }

    var block Block
    err = json.Unmarshal(blockData, &block)
    if err != nil {
        return nil, err

    }
    return &block, nil
}
//nonce tiene que ser incremental
func calculateHash(b Block) string {
    data:= fmt.Sprintf("%d%d%d", b.Index, b.Timestamp, b.PreviousHash, b.Nonce)
    for _, tx := range b.Transactions{
        data += fmt.Sprintf("%s%s%f", tx.Sender , tx.Recipient, tx.Amount)
    }
    h := sha256.New()
    h.Write([]byte(data))
    return hex.EncodeToString(h.Sum(nil))
}

func generateBlock(index int, previousHash string, transactions []Transaction, nonce int) Block {
    block := Block{
        Index: index,
        Timestamp: time.Now().Unix(),
        Transactions: transactions,
        PreviousHash: previousHash,
        Nonce: nonce,
    }
    block.Hash = calculateHash(block)


    return block
}
func PrintBlockData(block *Block){
    fmt.Println("Contenido del bloque:")
    fmt.Printf("Index: %d\n", block.Index)
    fmt.Printf("Timestamp: %d\n", block.Timestamp)
    fmt.Println("Transactions:")
    for i, tx := range block.Transactions {
        fmt.Printf("  Transacción %d:\n", i+1)
        fmt.Printf("    Sender: %s\n", tx.Sender)
        fmt.Printf("    Recipient: %s\n", tx.Recipient)
        fmt.Printf("    Amount: %.2f\n", tx.Amount)
    }
}
func main(){
    transactions := []Transaction{
        // Ejemplo 1
        {
            Sender:    "Alice",
            Recipient: "Bob",
            Amount:    25.0,
        },
        // Ejemplo 2
        {
            Sender:    "Charlie",
            Recipient: "David",
            Amount:    50.5,
        },
        // Ejemplo 3
        {
            Sender:    "Eve",
            Recipient: "Frank",
            Amount:    100.75,
        },
        // Agrega más transacciones aquí si lo deseas
    }

    previousHash := "hash_previo" // Reemplaza con el hash anterior adecuado
    nonce := 1                 // Reemplaza con el valor adecuado



    block := generateBlock(9999, previousHash, transactions, nonce)
    //PrintBlockData(block)

    dbPath := "./leveldb"
    // Abrir la base de datos (creará una si no existe)
    db, err := leveldb.OpenFile(dbPath, nil)
    if err != nil {
        log.Fatal(err)
    }
    defer db.Close()


    if err := saveBlock(db, block); err != nil {
        log.Fatal(err)
    }
    Index := 1
    blockAux, err := loadBlock(db,  Index)
    if err != nil {
        log.Fatal(err)
    }

    PrintBlockData(blockAux)

    for {
        // Mostrar el menú
        fmt.Println("Menú:")
        fmt.Println("1. Opción 1")
        fmt.Println("2. Opción 2")
        fmt.Println("3. Salir")

        // Leer la opción del usuario
        var opcion int
        fmt.Print("Seleccione una opción: ")
        _, err := fmt.Scan(&opcion)
        if err != nil {
            fmt.Println("Error al leer la opción:", err)
            continue
        }

        // Procesar la opción seleccionada
        switch opcion {
        case 1:
            fmt.Println("Has seleccionado la Opción 1.")
            // Agrega aquí el código para la Opción 1
        case 2:
            fmt.Println("Has seleccionado la Opción 2.")
            // Agrega aquí el código para la Opción 2
        case 3:
            fmt.Println("Saliendo del programa.")
            os.Exit(0) // Salir del programa
        default:
            fmt.Println("Opción no válida. Inténtalo de nuevo.")
        }
    }


}

