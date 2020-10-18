package main

import (
	"encoding/csv"
	"fmt"
	"log"
	"math/rand"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/tutorialedge/go-grpc-tutorial/chat"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
)

func EscribirCsv(aEscribir string, camion string) {

	f, err := os.OpenFile("../archivos/registro"+camion+".csv", os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		fmt.Println(err)
		return
	}
	w := csv.NewWriter(f)
	nueva := strings.Split(aEscribir, "+")

	w.Write(nueva)
	w.Flush()
	return
}

func SendCamion(aEnviar string, tipoMsj string) string {

	var conn *grpc.ClientConn
	conn, err := grpc.Dial("10.10.28.154:50052", grpc.WithInsecure())
	if err != nil {
		log.Fatalf("No se pudo conectar: %s", err)
	}
	defer conn.Close()

	c := chat.NewChatServiceClient(conn)

	response, err := c.SayHelloAgain(context.Background(), &chat.Message{Body: aEnviar, Otro: tipoMsj})
	if err != nil {
		log.Fatalf("Error al llamar a SayHelloAgain: %s", err)
	}
	log.Printf("Respuesta del Servidor: %s", response.Body)
	return response.Body

}

func LlevarPaquetes(paquete1 string, paquete2 string, demora int, camion string) {
	fmt.Println("Entregando paquetes")
	entrega1 := strings.Split(paquete1, "+")
	entregado1 := 0
	intentos1 := 0
	fechaEntrega1 := ""
	entrega2 := strings.Split(paquete2, "+")
	entregado2 := 0
	intentos2 := 0
	fechaEntrega2 := ""

	if entrega1[1] == "retail" && entrega2[1] == "retail" { //Si ambas entregas son retail, las dos se intentan 3 veces
		for i := 0; i < 3; i++ {
			if entregado1 == 0 {
				time.Sleep(time.Duration(demora) * time.Second)
				rand.Seed(time.Now().UnixNano())
				if rand.Intn(100) < 80 {
					entregado1 = 1
					Hora := time.Now()
					fechaEntrega1 = Hora.Format("2006-01-02 15:04:05")
					fmt.Println("Entregado " + entrega1[0])
				}
				intentos1++
			}
			if entregado2 == 0 {
				time.Sleep(time.Duration(demora) * time.Second)
				rand.Seed(time.Now().UnixNano())
				if rand.Intn(100) < 80 {
					entregado2 = 1
					Hora := time.Now()
					fechaEntrega2 = Hora.Format("2006-01-02 15:04:05")
					fmt.Println("Entregado " + entrega2[0])
				}
				intentos2++
			}
		}
	} else if entrega1[1] == "retail" && entrega2[1] == "prioritario" { //Si una es retail y la otra prioritaria, se intenta 3 veces la primera y la segunda 2 (a menos que la penalizacion supere el valor del producto, que sería la ganancia también)
		for i := 0; i < 3; i++ {
			if entregado1 == 0 {
				time.Sleep(time.Duration(demora) * time.Second)
				rand.Seed(time.Now().UnixNano())
				if rand.Intn(100) < 80 {
					entregado1 = 1
					Hora := time.Now()
					fechaEntrega1 = Hora.Format("2006-01-02 15:04:05")
					fmt.Println("Entregado " + entrega1[0])
				}
				intentos1++
			}
		}
		Valor, _ := strconv.Atoi(entrega2[2])
		penalizacion := 0
		for j := 0; j < 2 || penalizacion < Valor; j++ {
			if entregado2 == 0 {
				time.Sleep(time.Duration(demora) * time.Second)
				rand.Seed(time.Now().UnixNano())
				if rand.Intn(100) < 80 {
					entregado2 = 1
					Hora := time.Now()
					fechaEntrega2 = Hora.Format("2006-01-02 15:04:05")
					fmt.Println("Entregado " + entrega2[0])
				}
				intentos2++
				penalizacion = penalizacion + 10
				if penalizacion > Valor {
					break
				}
			}
		}
	} else if entrega1[1] == "retail" && paquete2 == "SINPAQUETE" { //Caso en que el camion retail va con una sola entrega
		for i := 0; i < 3; i++ {
			if entregado1 == 0 {
				time.Sleep(time.Duration(demora) * time.Second)
				rand.Seed(time.Now().UnixNano())
				if rand.Intn(100) < 80 {
					entregado1 = 1
					Hora := time.Now()
					fechaEntrega1 = Hora.Format("2006-01-02 15:04:05")
					fmt.Println("Entregado " + entrega1[0])
				}
				intentos1++
			}
		}
		if entregado1 == 0 {
			fechaEntrega1 = "0"
		}
		entrega1[5] = strconv.Itoa(intentos1)
		paquete1 = strings.Join(entrega1, "+") + "+" + fechaEntrega1
		EscribirCsv(paquete1, camion)
		SendCamion(paquete1, "ENTREGA")
		return
	} else { //si ambos son prioritario o normal, se intentan 2 veces a menos que la penalización supere el valor
		Valor, _ := strconv.Atoi(entrega1[2])
		penalizacion := 0
		for i := 0; i < 2; i++ {
			if entregado1 == 0 {
				time.Sleep(time.Duration(demora) * time.Second)
				rand.Seed(time.Now().UnixNano())
				if rand.Intn(100) < 80 {
					entregado1 = 1
					Hora := time.Now()
					fechaEntrega1 = Hora.Format("2006-01-02 15:04:05")
					fmt.Println("Entregado " + entrega1[0])
				}
				intentos1++
				penalizacion = penalizacion + 10
				if penalizacion > Valor {
					break
				}

			}
		}
		if paquete2 != "SINPAQUETES" {
			Valor, _ = strconv.Atoi(entrega2[2])
			penalizacion = 0
			for j := 0; j < 2; j++ {
				if entregado2 == 0 {
					time.Sleep(time.Duration(demora) * time.Second)
					rand.Seed(time.Now().UnixNano())
					if rand.Intn(100) < 80 {
						entregado2 = 1
						Hora := time.Now()
						fechaEntrega2 = Hora.Format("2006-01-02 15:04:05")
						fmt.Println("Entregado " + entrega2[0])
					}
					intentos2++
					penalizacion = penalizacion + 10
					if penalizacion > Valor {
						break
					}
				}
			}
		} else { //Caso especial en que el camion va con un solo paquete normal o prioritario
			if entregado1 == 0 {
				fechaEntrega1 = "0"
			}
			entrega1[5] = strconv.Itoa(intentos1)
			paquete1 = strings.Join(entrega1, "+") + "+" + fechaEntrega1
			EscribirCsv(paquete1, camion)
			SendCamion(paquete1, "ENTREGA")
			return
		}
	}
	if entregado1 == 0 {
		fechaEntrega1 = "0"
	}
	if entregado2 == 0 {
		fechaEntrega2 = "0"
	}
	entrega1[5] = strconv.Itoa(intentos1)
	entrega2[5] = strconv.Itoa(intentos2)
	paquete1 = strings.Join(entrega1, "+") + "+" + fechaEntrega1
	paquete2 = strings.Join(entrega2, "+") + "+" + fechaEntrega2
	SendCamion(paquete1, "ENTREGA")
	SendCamion(paquete2, "ENTREGA")
	EscribirCsv(paquete1, camion)
	EscribirCsv(paquete2, camion)
}

func main() {
	fmt.Println("Seleccione el tipo de camión:\n	1) Retail 1\n	2) Retail 2\n	3) Normal")
	var tipo string
	var segundos int
	var demora int
	fmt.Scanf("%s", &tipo)
	if tipo == "1" {
		tipo = "Retail1"
	} else if tipo == "2" {
		tipo = "Retail2"
	} else {
		tipo = "Normal"
	}
	fmt.Println("Ingrese cuántos segundos esperará a un segundo paquete: ")
	fmt.Scanf("%d", &segundos)
	fmt.Println("Ingrese cuántos segundos demorará en entregar un paquete: ")
	fmt.Scanf("%d", &demora)
	var paquete1 string
	var paquete2 string
	for { //iterar por siempre
		response := SendCamion(tipo, "Paquete")
		if response == "SINPAQUETES" { //no hay paquetes
			time.Sleep(5 * time.Second)
		} else { // Sí hay paquetes
			paquete1 = response
			response := SendCamion(tipo, "Paquete") //pide un segundo paquete
			if response == "SINPAQUETES" {          //si no hay un segundo paquete
				time.Sleep(time.Duration(segundos) * time.Second) //espera los segundos indicados
				response := SendCamion(tipo, "Paquete")           //vuelve a pedir el paquete
				paquete2 = response
			}
			paquete2 = response
			LlevarPaquetes(paquete1, paquete2, demora, tipo)
		}
	}

}
