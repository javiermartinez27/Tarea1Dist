package main

import (
	"encoding/csv"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/streadway/amqp"
)

func failOnError(err error, msg string) {
	if err != nil {
		log.Fatalf("%s: %s", msg, err)
	}
}

//Funcion encargada de escribir en finanzas.
func EscribirCsv(aEscribir string) {

	f, err := os.OpenFile("../archivos/finanzas.csv", os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)
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

/*variable donde se acumula ganancias*/
var g int

func recive(segundos int) {
	conn, err := amqp.Dial("amqp://guest:guest@localhost:5672/")
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

	msgs, err := ch.Consume(
		q.Name, // queue
		"",     // consumer
		true,   // auto-ack
		false,  // exclusive
		false,  // no-local
		false,  // no-wait
		nil,    // args
	)
	failOnError(err, "Failed to register a consumer")

	forever := make(chan bool)

	go func() {
		for {
			time.Sleep(time.Duration(segundos) * time.Second)
			for d := range msgs {
				log.Printf("Mensaje desde Logistica: %s", d.Body)
				mensaje := string(d.Body)

				test := strings.Split(mensaje, "+")

				ganancia, err := strconv.Atoi(test[2])
				if err != nil {
					log.Fatalf("Cannot retrieve quantity of %c: %s\n", ganancia, err)
				}
				intentos, err := strconv.Atoi(test[1])
				if err != nil {
					log.Fatalf("Cannot retrieve quantity of %c: %s\n", ganancia, err)
				}
				if intentos >= 2 {
					ganancia = ganancia - intentos*10
				}
				g += ganancia
				EscribirCsv(mensaje)

			}
		}

	}()
	log.Printf(" [*] Esperando mensajes desde logistica.")
	<-forever
}

func main() {
	fmt.Println("Bienvenido al sistema de Finanzas!\n Ingrese cada cuántos segundos quiere recibir información")
	var segundos int
	fmt.Scanf("%d", &segundos)

	go recive(segundos)

	time.Sleep(time.Duration(3) * time.Second)

	fmt.Println("Ingrese cualquier tecla para salir")
	var opcion int
	fmt.Scanln(&opcion)

	if opcion == 0 {
		println("Balance total:", g, " DigniPesos")
	} else {
		println("Balance totales:", g, " DigniPesos")
	}

}
