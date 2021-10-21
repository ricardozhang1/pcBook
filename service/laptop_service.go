package service

import (
	"context"
	"errors"
	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"log"
	"pc_book/pd"
)

// LaptopService is the server that provides laptop service
type LaptopService struct {
	pd.UnimplementedLaptopServiceServer
	Store LaptopStore
}

func NewLaptopService(store LaptopStore) *LaptopService {
	return &LaptopService{
		Store: store,
	}
}

// CreateLaptop is a unary RPC to create a new laptop.
func (server *LaptopService) CreateLaptop(ctx context.Context, req *pd.CreateLaptopRequest) (*pd.CreateLaptopResponse, error) {
	laptop := req.GetLaptop()
	log.Printf("receive a create-laptop request with id: %s", laptop.Id)

	if len(laptop.Id) > 0 {
		// check if it's valid uuid
		_, err := uuid.Parse(laptop.Id)
		if err != nil {
			return nil, status.Errorf(codes.InvalidArgument, "laptop ID is not a valid UUID: %v", err)
		}
	} else {
		id, err := uuid.NewRandom()
		if err != nil {
			return nil, status.Errorf(codes.Internal, "can not generate a new laptop ID: %v", err)
		}
		laptop.Id = id.String()
	}

	// some heavy process
	//time.Sleep(time.Second * 6)

	if ctx.Err() == context.Canceled {
		log.Println("request is canceled")
		return nil, status.Error(codes.DeadlineExceeded, "request is canceled")
	}

	if ctx.Err() == context.DeadlineExceeded {
		log.Println("deadline is exceeded")
		return nil, status.Error(codes.DeadlineExceeded, "deadline is exceeded")
	}

	// save the laptop to in-memory store
	err := server.Store.Save(laptop)
	if err != nil {
		code := codes.Internal
		if errors.Is(err, ErrAlreadyExists) {
			code = codes.AlreadyExists
		}
		return nil, status.Errorf(code, "can not save laptop to store: %v", err)
	}

	log.Printf("save laptop with ID: %v", laptop.Id)
	res := &pd.CreateLaptopResponse{
		Id: laptop.Id,
	}
	return res, nil
}

// SearchLaptop is a server-streaming RPC to search for laptop
func (server *LaptopService) SearchLaptop(req *pd.SearchLaptopRequest, stream pd.LaptopService_SearchLaptopServer) error {
	filter := req.GetFilter()
	log.Printf("recieve a search-laptop request with filter: %v", filter)

	err := server.Store.Search(stream.Context(), filter, func(laptop *pd.Laptop) error {
		res := &pd.SearchLaptopResponse{Laptop: laptop}
		err := stream.Send(res)
		if err != nil {
			return err
		}

		log.Printf("sent laptop with id: %s", laptop.GetId())
		return nil
	})
	if err != nil {
		return err
	}
	return nil
}




