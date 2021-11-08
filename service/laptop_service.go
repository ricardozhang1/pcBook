package service

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"io"
	"log"
	"pc_book/pd"
)

// 设定上传图片的最大大小
const maxImageSize = 1 << 20

// LaptopService is the server that provides laptop service
type LaptopService struct {
	pd.UnimplementedLaptopServiceServer
	laptopStore LaptopStore
	imageStore ImageStore
	ratingStore RatingStore
}

func NewLaptopService(laptopStore LaptopStore, imageStore ImageStore, ratingStore RatingStore) *LaptopService {
	return &LaptopService{
		laptopStore: laptopStore,
		imageStore: imageStore,
		ratingStore: ratingStore,
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
	if err := contextError(ctx); err != nil {
		return nil, err
	}

	// save the laptop to in-memory store
	err := server.laptopStore.Save(laptop)
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

	err := server.laptopStore.Search(stream.Context(), filter, func(laptop *pd.Laptop) error {
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

// UploadImage is a server-streaming RPC to search for image uploading
func (server *LaptopService) UploadImage(stream pd.LaptopService_UploadImageServer) error {
	req, err := stream.Recv()
	if err != nil {
		return logErr(status.Errorf(codes.Unknown, "cannot receive image info"))
	}

	laptopID := req.GetInfo().GetLaptopId()
	imageType := req.GetInfo().GetImageType()
	log.Printf("receive an upload-image request for laptop %s with image type %s", laptopID, imageType)

	laptop, err := server.laptopStore.Find(laptopID)
	if err != nil {
		return logErr(status.Errorf(codes.Internal, "cannot find laptop: %v", err))
	}

	if laptop == nil {
		return logErr(status.Errorf(codes.InvalidArgument, "laptop %s doesnt exist", laptopID))
	}

	imageData := bytes.Buffer{}
	imageSize := 0

	for {
		// check context error  用于上传图片超时
		if err := contextError(stream.Context()); err != nil {
			return err
		}

		log.Print("waiting to receive more data")
		req, err := stream.Recv()
		if err == io.EOF {
			log.Print("no more data")
			break
		}

		if err != nil {
			return logErr(status.Errorf(codes.Unknown, "cannot receive chunk data: %v", err))
		}

		chunk := req.GetChunkData()
		size := len(chunk)

		log.Printf("receive a chunk with size: %d", size)

		imageSize += size
		if imageSize > maxImageSize {
			return logErr(status.Errorf(codes.InvalidArgument, "image is too large: %d > %d", imageSize, maxImageSize))
		}

		_, err = imageData.Write(chunk)
		if err != nil {
			return logErr(status.Errorf(codes.Unknown, "cannot write image to store: %v", err))
		}
	}

	fmt.Println(laptopID, imageType)
	imageID, err := server.imageStore.Save(laptopID, imageType, imageData)
	if err != nil {
		return logErr(status.Errorf(codes.Internal, "cannot save image to store: %v", err))
	}

	res := &pd.UploadImageResponse{
		Id: imageID,
		Size: uint32(imageSize),
	}

	err = stream.SendAndClose(res)
	if err != nil {
		return logErr(status.Errorf(codes.Unknown, "cannot send response: %v", err))
	}

	log.Printf("saved image with id: %s, size: %d", imageID, imageSize)
	return nil
}

// RateLaptop is a bidirectional-streaming RPC that allows client to rate a stream of laptops
// with a score, and returns a stream of average score for each of them
func (server *LaptopService) RateLaptop(stream pd.LaptopService_RateLaptopServer) error  {
	for {
		if err := contextError(stream.Context()); err != nil {
			return err
		}

		req, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			return logErr(status.Errorf(codes.Unknown, "cannot receive stream request: %v", err))
		}

		laptopID := req.GetLaptopId()
		score := req.GetScore()

		log.Printf("receive a rate-laptop request: %s, score: %.2f", laptopID, score)

		found, err := server.laptopStore.Find(laptopID)
		if err != nil {
			return logErr(status.Errorf(codes.Internal, "cannot find laptop: %v", err))
		}
		if found == nil {
			return logErr(status.Errorf(codes.NotFound, "laptopId %s is not found", laptopID))
		}

		rating, err := server.ratingStore.Add(laptopID, score)
		if err != nil {
			return logErr(status.Errorf(codes.Internal, "cannot add rating to the score: %v", err))
		}

		res := &pd.RateLaptopResponse{
			LaptopId: laptopID,
			RateCount: rating.Count,
			AverageScore: rating.Sum / float64(rating.Count),
		}
		if err := stream.Send(res); err != nil {
			return logErr(status.Errorf(codes.Unknown, "cannot send stream response: %v", err))
		}

	}
	return nil
}

func logErr(err error) error {
	if err != nil {
		log.Print(err)
	}
	return err
}

// 统一处理context，超时的错误
func contextError(ctx context.Context) error {
	switch ctx.Err() {
	case context.Canceled:
		return logErr(status.Error(codes.DeadlineExceeded, "request is canceled"))
	case context.DeadlineExceeded:
		return logErr(status.Error(codes.DeadlineExceeded, "deadline is exceeded"))
	default:
		return nil
	}
}




