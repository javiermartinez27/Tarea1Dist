package main

import (
	"fmt"
	"log"
	"net"

	"github.com/tutorialedge/go-grpc-tutorial/chat"
	"google.golang.org/grpc"
)

func Abrir(port string, usuario string) {

	fmt.Println("Escuchando al " + usuario + " en el puerto " + port)

	lis, err := net.Listen("tcp", port)
	if err != nil {
		log.Fatalf("Hubo un fallo al abrir el servidor: %v", err)
	}

	s := chat.Server{}

	grpcServer := grpc.NewServer()

	chat.RegisterChatServiceServer(grpcServer, &s)

	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("Hubo un fallo al abrir el servidor gRPC: %s", err)
	}
}

func main() {
	fmt.Println("Abriendo servido de logística\n Bienvenido al sistema de Logistica!")

	go Abrir(":50052", "Camiones")
	Abrir(":50051", "Clientes")

}
