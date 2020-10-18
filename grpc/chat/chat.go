package chat

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	"encoding/csv"
	"io/ioutil"
	"strings"

	"github.com/streadway/amqp"
	"golang.org/x/net/context"
)

var colaRetail []paquete
var colaPrioritario []paquete
var colaNormal []paquete

func CheckError(message string, err error) {
	if err != nil {
		log.Fatal(message, err)
	}
}

func EscribirCSV(aEscribir string) string {

	nroOrden, err := ioutil.ReadFile("../archivos/indexAct.data")
	if err != nil {
		fmt.Println(err)
	}

	f, err := os.OpenFile("../archivos/registro.csv", os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		fmt.Println(err)
		return "0"
	}
	w := csv.NewWriter(f)
	nroOrdenStr := string(nroOrden[:])

	Hora := time.Now()
	HoraStr := Hora.Format("2006-01-02 15:04:05")

	aEscribir = HoraStr + "+" + aEscribir + "+" + nroOrdenStr

	nueva := strings.Split(aEscribir, "+")
	retorno := "El código de seguimiento de " + nueva[1] + " es " + nroOrdenStr
	w.Write(nueva)

	var aux paquete //anadiendo a las colas en memoria
	ValorInt, _ := strconv.Atoi(nueva[4])
	aux = paquete{nueva[1], nueva[7], nueva[2], ValorInt, 0, "En bodega", nueva[5], nueva[6]}
	if nueva[2] == "retail" {
		colaRetail = append(colaRetail, aux)
	} else if nueva[2] == "prioritario" {
		colaPrioritario = append(colaPrioritario, aux)
	} else {
		colaNormal = append(colaNormal, aux)
	}

	nroOrdenInt, err := strconv.Atoi(nroOrdenStr)
	bs := []byte(strconv.Itoa(nroOrdenInt + 1))

	error := ioutil.WriteFile("../archivos/indexAct.data", bs, 0777)
	if error != nil {
		fmt.Println(err)
	}

	w.Flush()
	return retorno
}

type paquete struct {
	IDPaquete   string
	seguimiento string
	tipo        string
	valor       int
	intentos    int
	estado      string
	origen      string
	destino     string
}

type Server struct {
}

func failOnError(err error, msg string) {
	if err != nil {
		log.Fatalf("%s: %s", msg, err)
	}
}

func sendToFinanzas(aEnviar string) {
	conn, err := amqp.Dial("amqp://guest:guest@10.10.28.157:5672/")
	failOnError(err, "Failed to connect to RabbitMQ")
	defer conn.Close()

	ch, err := conn.Channel()
	failOnError(err, "Failed to open a channel")
	defer ch.Close()

	q, err := ch.QueueDeclare(
		"hello", // name
		false,   // durable
		false,   // delete when unused
		false,   // exclusive
		false,   // no-wait
		nil,     // arguments
	)
	failOnError(err, "Failed to declare a queue")

	body := aEnviar
	err = ch.Publish(
		"",     // exchange
		q.Name, // routing key
		false,  // mandatory
		false,  // immediate
		amqp.Publishing{
			ContentType: "text/plain",
			Body:        []byte(body),
		})
	log.Printf(" [x] Sent %s", body)
	failOnError(err, "Failed to publish a message")
}

func (s *Server) SayHello(ctx context.Context, message *Message) (*Message, error) {
	mensaje := strings.Split(message.Body, "_")
	log.Printf("Mensaje recibido desde el cliente: %s", mensaje[0])
	if mensaje[1] == "orden" {
		mensaje := EscribirCSV(mensaje[0])
		return &Message{Body: mensaje}, nil
	} else { //Si se mandó un codigo
		if colaRetail == nil && colaNormal == nil && colaPrioritario == nil {
			fmt.Println("Codigo no encontrado")
			return &Message{Body: "No hay ordenes ingresadas"}, nil
		} else {
			for _, orden := range colaRetail {
				if orden.seguimiento == mensaje[0] {
					return &Message{Body: orden.estado}, nil
				}
			}
			for _, orden := range colaNormal {
				if orden.seguimiento == mensaje[0] {
					return &Message{Body: orden.estado}, nil
				}
			}
			for _, orden := range colaPrioritario {
				if orden.seguimiento == mensaje[0] {
					return &Message{Body: orden.estado}, nil
				}
			}
			return &Message{Body: "No se encontró el código"}, nil
		}
	}
}

func (s *Server) SayHelloAgain(ctx context.Context, in *Message) (*Message, error) {
	log.Printf("Recibido el mensaje del Camion: %s", in.Body)
	if in.Otro == "ENTREGA" { //si el mensaje corresponde a un camion avisando una entrega completada
		fmt.Println("Recibido " + in.Body)
		sendToFinanzas(in.Body)
		entrega := strings.Split(in.Body, "+")
		if entrega[1] == "retail" {
			for i, _ := range colaRetail {
				if colaRetail[i].IDPaquete == entrega[0] {
					if entrega[6] != "0" {
						colaRetail[i].estado = "Recibido"
					} else {
						colaRetail[i].estado = "No recibido"
					}
					intento, _ := strconv.Atoi(entrega[5])
					colaRetail[i].intentos = intento
				}
			}
		} else if entrega[1] == "prioritario" {
			for i, _ := range colaPrioritario {
				if colaPrioritario[i].IDPaquete == entrega[0] {
					if entrega[6] != "0" {
						colaPrioritario[i].estado = "Recibido"
					} else {
						colaPrioritario[i].estado = "No recibido"
					}
					intento, _ := strconv.Atoi(entrega[5])
					colaPrioritario[i].intentos = intento
				}
			}
		} else {
			for i, _ := range colaNormal {
				if colaNormal[i].IDPaquete == entrega[0] {
					if entrega[6] != "0" {
						colaNormal[i].estado = "Recibido"
					} else {
						colaNormal[i].estado = "No recibido"
					}
					intento, _ := strconv.Atoi(entrega[5])
					colaNormal[i].intentos = intento
				}
			}
		}
		return &Message{Body: "Orden actualizada"}, nil

	} else { //si el mensaje es un camion pidiendo un paquete
		if in.Body == "Retail1" || in.Body == "Retail2" { //Si es camion retail
			if colaRetail == nil && colaPrioritario == nil {
				return &Message{Body: "SINPAQUETES"}, nil
			} else if colaRetail != nil { //Si hay paquetes
				for i, _ := range colaRetail { //busca paquetes retail
					if colaRetail[i].estado == "En bodega" { //solo paquetes retail que esten en bodega
						aux := colaRetail[i]
						mensaje := aux.IDPaquete + "+" + aux.tipo + "+" + strconv.Itoa(aux.valor) + "+" + aux.origen + "+" + aux.destino + "+" + strconv.Itoa(aux.intentos)
						colaRetail[i].estado = "En camino"
						return &Message{Body: mensaje}, nil
					}
				}
				for i, _ := range colaPrioritario { //si no hay paquetes retail sin despachar, busca en los prioritarios
					if colaPrioritario[i].estado == "En bodega" {
						aux := colaPrioritario[i]
						mensaje := aux.IDPaquete + "+" + aux.tipo + "+" + strconv.Itoa(aux.valor) + "+" + aux.origen + "+" + aux.destino + "+" + strconv.Itoa(aux.intentos)
						colaPrioritario[i].estado = "En camino"
						return &Message{Body: mensaje}, nil
					}
				}
			}

		} else { //Si es camion normal
			if colaPrioritario == nil && colaNormal == nil {
				return &Message{Body: "SINPAQUETES"}, nil
			} else { //Si hay paquetes
				for i, _ := range colaPrioritario { //si hay prioritarios
					if colaPrioritario[i].estado == "En bodega" { //si hay prioritarios sin despachar
						aux := colaPrioritario[i]
						mensaje := aux.IDPaquete + "+" + aux.tipo + "+" + strconv.Itoa(aux.valor) + "+" + aux.origen + "+" + aux.destino + "+" + strconv.Itoa(aux.intentos)
						colaPrioritario[i].estado = "En camino"
						return &Message{Body: mensaje}, nil
					}
				}
				for i, _ := range colaNormal { //si no quedan prioritarios sin despachar, busca en paquetes normales
					if colaNormal[i].estado == "En bodega" {
						aux := colaNormal[i]
						mensaje := aux.IDPaquete + "+" + aux.tipo + "+" + strconv.Itoa(aux.valor) + "+" + aux.origen + "+" + aux.destino + "+" + strconv.Itoa(aux.intentos)
						colaNormal[i].estado = "En camino"
						return &Message{Body: mensaje}, nil
					}
				}
			}
		}
	}
	return &Message{Body: "SINPAQUETES"}, nil
}
