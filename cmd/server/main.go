package main

import (
	"flag"
	"fmt"
	"google.golang.org/grpc"
	"log"
	"net"
	"pc_book/pd"
	"pc_book/service"
)

func main() {
	fmt.Println("grpc server")

	port := flag.Int("port", 0, "the server port")
	flag.Parse()
	log.Printf("start server on port: %d", *port)

	laptopStore := service.NewInMemoryLaptopStore()
	imageStore := service.NewDiskImageStore("img")

	laptopServer := service.NewLaptopService(laptopStore, imageStore)
	grpcServer := grpc.NewServer()
	pd.RegisterLaptopServiceServer(grpcServer, laptopServer)

	address := fmt.Sprintf("0.0.0.0:%d", *port)
	listener, err := net.Listen("tcp", address)
	if err != nil {
		log.Fatal("can not start listener: ", err)
	}

	err = grpcServer.Serve(listener)
	if err != nil {
		log.Fatal("can not start grpcServer: ", err)
	}


}


