package main

import (
	"log"

	"golang.org/x/net/context"
	"google.golang.org/grpc"

	"encoding/csv"
	"fmt"
	"os"
	"time"

	"github.com/tutorialedge/go-grpc-tutorial/chat"
)

func readCsvFile(filePath string) [][]string {
	f, err := os.Open(filePath)
	if err != nil {
		log.Fatal("Unable to read input file "+filePath, err)
	}
	defer f.Close()

	csvReader := csv.NewReader(f)
	records, err := csvReader.ReadAll()
	if err != nil {
		log.Fatal("Unable to parse file as CSV for "+filePath, err)
	}

	return records
}

func enviar(records [][]string, tipo string, segundos int) {
	var aux string
	var tipoAux string
	tipoAux = tipo
	for i, producto := range records {
		if i != 0 {
			if tipo == "pyme" && producto[len(producto)-1] == "1" {
				tipo = "prioritario"
			} else if tipo == "pyme" && producto[len(producto)-1] == "0" {
				tipo = "normal"
			}
			for j, carac := range producto {
				if j == 0 {
					aux = carac + "+" + tipo + "+"
				} else if j < 4 {
					aux = aux + carac + "+"
				} else if j == 4 {
					aux = aux + carac
				} else {
					continue
				}
			}
			time.Sleep(time.Duration(segundos) * time.Second)
			Send(aux, "orden")
			tipo = tipoAux
		}
	}
}

func Send(aEnviar string, funcion string) {

	var conn *grpc.ClientConn
	conn, err := grpc.Dial("10.10.28.154:50051", grpc.WithInsecure())
	if err != nil {
		log.Fatalf("No se pudo conectar: %s", err)
	}
	defer conn.Close()

	c := chat.NewChatServiceClient(conn)

	response, err := c.SayHello(context.Background(), &chat.Message{Body: aEnviar + "_" + funcion, Otro: "Otro"})
	if err != nil {
		log.Fatalf("Error al llamar a SayHello: %s", err)
	}
	log.Printf("Respuesta del Servidor: %s", response.Body)

}

func main() {
	fmt.Println("Bienvenido al sistema de Clientes!\n Seleccione una opción:")

	var opcion string
	opcion = "0"

	for opcion != "4" {

		fmt.Println("	1) Ingresar órdenes como Cliente retail\n	2) Ingresar órdenes como Cliente pyme\n	3) Ver estado de un pedido\n	4) Salir")
		fmt.Scanf("%s", &opcion)
		var segundos int
		var codigo string
		if opcion == "1" {
			fmt.Println("Elegiste la opción ingresar ordenes como cliente retail")
			fmt.Println("Indique cada cuántos segundos quiere enviar órdenes:")
			fmt.Scanf("%d", &segundos)
			records := readCsvFile("../archivos/retail.csv")
			enviar(records, "retail", segundos)
		} else if opcion == "2" {
			fmt.Println("Elegiste la opción ingresar ordenes como cliente pyme")
			fmt.Println("Indique cada cuántos segundos quiere enviar órdenes:")
			fmt.Scanf("%d", &segundos)
			records := readCsvFile("../archivos/pymes.csv")
			enviar(records, "pyme", segundos)
		} else if opcion == "3" {
			fmt.Println("Elegiste la opción ver estado de una orden")
			fmt.Println("Ingrese el código de seguimiento:")
			fmt.Scanf("%s", &codigo)
			Send(codigo, "codigo")
		} else if opcion == "4" {
			fmt.Print("Adios\n")
		} else {
			fmt.Println("Por favor, selecciona una opción válida")
		}
	}
}
