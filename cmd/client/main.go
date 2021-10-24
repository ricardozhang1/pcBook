package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"io"
	"log"
	"os"
	"path/filepath"
	"pc_book/pd"
	"pc_book/sample"
	"time"
)

func createLaptop(laptopClient pd.LaptopServiceClient, laptop *pd.Laptop) {
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

func uploadImage(laptopClient pd.LaptopServiceClient, laptopID string, imagePath string) {
	file, err := os.Open(imagePath)
	if err != nil {
		log.Fatal("can not open image file: ", err)
	}
	defer file.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	stream, err := laptopClient.UploadImage(ctx)
	if err != nil {
		log.Fatal("cannot upload image: ", err)
	}

	req := &pd.UploadImageRequest{
		Data: &pd.UploadImageRequest_Info{
			Info: &pd.ImageInfo{
				LaptopId: laptopID,
				ImageType: filepath.Ext(imagePath),
			},
		},
	}

	err = stream.Send(req)
	if err != nil {
		log.Fatal("can not send image: ", err, stream.RecvMsg(nil))
	}

	reader := bufio.NewReader(file)
	buffer := make([]byte, 1024)

	for {
		n, err := reader.Read(buffer)
		if err == io.EOF {
			break
		}

		if err != nil {
			log.Fatal("cannot read chunk to buffer: ", err)
		}

		req := &pd.UploadImageRequest{
			Data: &pd.UploadImageRequest_ChunkData{
				ChunkData: buffer[:n],
			},
		}

		err = stream.Send(req)
		if err != nil {
			log.Fatal("can not send chunk to server: ", err)
		}
	}

	res, err := stream.CloseAndRecv()
	if err != nil {
		log.Fatal("cannot receive response: ", err)
	}

	log.Printf("image upload with id: %s, size: %d", res.GetId(), res.GetSize())


}

func testCreateLaptop(laptopClient pd.LaptopServiceClient) {
	createLaptop(laptopClient, sample.NewLaptop())
}

func testSearchLaptop(laptopClient pd.LaptopServiceClient) {
	for i := 0; i < 10; i++ {
		createLaptop(laptopClient, sample.NewLaptop())
	}

	filter := &pd.Filter{
		MaxPriceUsd: 3000,
		MinCpuCores: 4,
		MinCpuGhz: 2.5,
		MinRam: &pd.Memory{Value: 8, Unit: pd.Memory_GIGABYTE},
	}

	searchLaptop(laptopClient, filter)
}

func testUploadImage(laptopClient pd.LaptopServiceClient) {
	laptop := sample.NewLaptop()
	createLaptop(laptopClient, laptop)
	uploadImage(laptopClient, laptop.GetId(), "tmp/laptop.png")
}

func main() {
	fmt.Println("grpc client")

	//serverAddress := flag.String("address", "", "the server address")
	serverAddress := "0.0.0.0:8089"
	flag.Parse()
	fmt.Printf("dial server %s", serverAddress)

	conn, err := grpc.Dial(serverAddress, grpc.WithInsecure())
	if err != nil {
		log.Fatal("can not dail server: ", err)
	}

	laptopClient := pd.NewLaptopServiceClient(conn)

	testUploadImage(laptopClient)
}
