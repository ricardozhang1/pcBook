package main

import (
	"context"
	"flag"
	"fmt"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"io"
	"log"
	"pc_book/pd"
	"pc_book/sample"
	"time"
)

func createLaptop(laptopClient pd.LaptopServiceClient) {
	laptop := sample.NewLaptop()
	laptop.Id = ""

	req := &pd.CreateLaptopRequest{
		Laptop: laptop,
	}

	// set timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5 * time.Second)
	defer cancel()

	res, err := laptopClient.CreateLaptop(ctx, req)
	if err != nil {
		st, ok := status.FromError(err)
		if ok && st.Code() == codes.AlreadyExists {
			// not a big deal
			log.Println("laptop already exists")
		} else {
			log.Fatal("can not create a laptop: ", err)
		}
		// 存在其他错误情况
		return
	}

	log.Printf("create laptop with id: %s", res.Id)
}

func searchLaptop(laptopClient pd.LaptopServiceClient, filter *pd.Filter)  {
	log.Printf("search filter: %v", filter)

	ctx, cancel := context.WithTimeout(context.Background(), 5 * time.Second)
	defer cancel()

	req := &pd.SearchLaptopRequest{Filter: filter}
	stream, err := laptopClient.SearchLaptop(ctx, req)
	if err != nil {
		log.Fatal("cannot search laptop: ", err)
	}

	for {
		res, err := stream.Recv()
		if err == io.EOF {
			return
		}
		if err != nil {
			log.Fatal("cannot receive response: ", err)
		}

		laptop := res.GetLaptop()
		log.Println("- found: ", laptop.GetId())
		log.Println("	+ brand: ", laptop.GetBrand())
		log.Println("	+ name: ", laptop.GetName())
		log.Println("	+ cpu cores: ", laptop.GetCpu().GetNumberCores())
		log.Println("	+ cpu min ghz: ", laptop.GetCpu().GetMinGhz())
		log.Println("	+ ram: ", laptop.GetRam())
		log.Println("	+ price: ", laptop.GetPriceUsd())
	}
}

func main() {
	fmt.Println("grpc client")

	serverAddress := flag.String("address", "", "the server address")
	flag.Parse()
	fmt.Printf("dial server %s", serverAddress)

	conn, err := grpc.Dial(*serverAddress, grpc.WithInsecure())
	if err != nil {
		log.Fatal("can not dail server: ", err)
	}

	laptopClient := pd.NewLaptopServiceClient(conn)

	for i := 0; i < 10; i++ {
		createLaptop(laptopClient)
	}

	filter := &pd.Filter{
		MaxPriceUsd: 3000,
		MinCpuCores: 4,
		MinCpuGhz: 2.5,
		MinRam: &pd.Memory{Value: 8, Unit: pd.Memory_GIGABYTE},
	}

	searchLaptop(laptopClient, filter)

}
