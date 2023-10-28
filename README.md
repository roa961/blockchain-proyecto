
# Proyecto Blockchain: Entrega N°1 

En el repositorio es posible encontrar la entrega N°1 del proyecto semestral. Este contiene la implementación de las distintas funcionalidades que fueron solicitadas.





## Integrantes

- Gastón Donoso
- Claudio López
- Rómulo Otárola
- Pablo Sáez
## Funcionalidades

- **Realizar una transacción**: Cada una de estas puede ser realizada por dos usuario, Bob y Alice, los cuales se eligen al momento de hacer una transacción, donde se deberá especificar un monto a transferir.
- **Leer una transacción existente**: En este caso se debe especificar el índice del bloque a consultar, del cual se mostrará la información contenida en este, además de las transacciones que este posea. Además se muestra si la transacción está verificada o no y por quién.
- **Mostrar cadena de bloques**: En esta opción se puede mostrar la opción asociada a cada bloque, de manera secuencial.

## Dependencias
Para poder iniciar el programa, se debe ingresar el siguiente comando en la carpeta raiz del proyecto para crear un nuevo módulo:
```bash
go mod init blockchain-proyecto
```
De esta manera se hace un seguimiento de las dependencias que posea el programa.

Luego se ingresa el siguiente comando para instalar las dependencias:

```bash
go mod tidy
```

## Instrucciones de uso

Para ejecutar el código se debe utilizar el comando:

```bash
go run main.go
```
Se despliega un menú, del cual se debe elegir una de las opciones mostradas.



## Consideración: Firma

Para la presente entrega, se realiza la firma de cada una de las transacciones de manera correcta, pero al momento de realizar la verificación de estas firmas, no fue posible obtener la llave pública en el formato correcto.  
La forma que encontramos para firmar cada una las transacciones no nos permite utilizar las llaves generadas y almacendas en Leveldb para realizar la verificación.   
La verificación de firmas sólo es exitosa al realizarla en una misma ejecución del programa, ya que en siguientes ejecuciones, la llave pública es siempre distinta y la verificacion falla.  
El código no funcional para verificar firmas se encuentra en la rama [intento_firma](https://github.com/roa961/blockchain-proyecto/tree/intento_firma)