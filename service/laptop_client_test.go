package service_test

import (
	"context"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"io"
	"net"
	"pc_book/pd"
	"pc_book/sample"
	"pc_book/serializer"
	"pc_book/service"
	"testing"
)

func TestClientCreateLaptop(t *testing.T) {
	t.Parallel()

	laptopStore := service.NewInMemoryLaptopStore()
	serverAddress := startTestLaptopServer(t, laptopStore, nil)
	laptopClient := newTestLaptopClient(t, serverAddress)

	laptop := sample.NewLaptop()
	exceptId := laptop.Id

	req := &pd.CreateLaptopRequest{
		Laptop: laptop,
	}

	res, err := laptopClient.CreateLaptop(context.Background(), req)
	require.NoError(t, err)
	require.NotNil(t, res)
	require.Equal(t, res.Id, exceptId)

	// check that the laptop is saved to store
	other, err := laptopStore.Find(res.Id)
	require.NoError(t, err)
	require.NotNil(t, other)

	// check that the saved laptop is the same as the one we send
	// 将protobuf文件转换成json格式再进行比较
	requireSameLaptop(t, laptop, other)
}

// TestClientSearchLaptop 查找laptop的测试
func TestClientSearchLaptop(t *testing.T) {
	t.Parallel()

	filter := &pd.Filter{
		MaxPriceUsd: 2000,
		MinCpuCores: 4,
		MinCpuGhz: 2.2,
		MinRam: &pd.Memory{Value: 8, Unit: pd.Memory_GIGABYTE},
	}

	// 准备数据，确定存储方式
	laptopStore := service.NewInMemoryLaptopStore()
	expectedIDs := make(map[string]bool)

	for i:=0; i<6; i++ {
		laptop := sample.NewLaptop()

		switch i {
		case 0:
			laptop.PriceUsd = 2500
		case 1:
			laptop.Cpu.NumberCores = 2
		case 2:
			laptop.Cpu.MinGhz = 2.0
		case 3:
			laptop.Ram = &pd.Memory{Value: 4096, Unit: pd.Memory_MEGABYTE}
		case 4:
			laptop.PriceUsd = 1999
			laptop.Cpu.NumberCores = 4
			laptop.Cpu.MinGhz = 2.5
			laptop.Cpu.MaxGhz = 4.5
			laptop.Ram = &pd.Memory{Value: 16, Unit: pd.Memory_GIGABYTE}
			expectedIDs[laptop.Id] = true
		case 5:
			laptop.PriceUsd = 2000
			laptop.Cpu.NumberCores = 6
			laptop.Cpu.MinGhz = 2.5
			laptop.Cpu.MaxGhz = 5.0
			laptop.Ram = &pd.Memory{Value: 64, Unit: pd.Memory_GIGABYTE}
			expectedIDs[laptop.Id] = true
		}

		err := laptopStore.Save(laptop)
		require.NoError(t, err)
	}

	// 生成server端，返回地址
	serverAddress := startTestLaptopServer(t, laptopStore, nil)
	// 生成client端
	laptopClient := newTestLaptopClient(t, serverAddress)

	req := &pd.SearchLaptopRequest{Filter: filter}
	// client端调用 服务端的 SearchLaptop 发的方法，返回的结果传递给stream对象
	stream, err := laptopClient.SearchLaptop(context.Background(), req)
	require.NoError(t, err)

	found := 0
	for {
		res, err := stream.Recv()
		if err == io.EOF {
			break
		}
		require.NoError(t, err)
		require.Contains(t, expectedIDs, res.GetLaptop().Id)

		found += 1
	}
	require.Equal(t, len(expectedIDs), found)
}


func startTestLaptopServer(t *testing.T, laptopStore service.LaptopStore, imageStore service.ImageStore) string {
	laptopServer := service.NewLaptopService(laptopStore, imageStore)

	grpcServer := grpc.NewServer()
	pd.RegisterLaptopServiceServer(grpcServer, laptopServer)

	listener, err := net.Listen("tcp", ":0")  // random available port
	require.NoError(t, err)
	go grpcServer.Serve(listener)
	return listener.Addr().String()
}

func newTestLaptopClient(t *testing.T, serverAddress string) pd.LaptopServiceClient {
	conn, err := grpc.Dial(serverAddress, grpc.WithInsecure())
	require.NoError(t, err)

	return pd.NewLaptopServiceClient(conn)
}

func requireSameLaptop(t *testing.T, laptop1 *pd.Laptop, laptop2 *pd.Laptop) {
	json1, err := serializer.ProtobufToJSON(laptop1)
	require.NoError(t, err)

	json2, err := serializer.ProtobufToJSON(laptop2)
	require.NoError(t, err)

	require.Equal(t, json1, json2)
}




