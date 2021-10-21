package service_test

import (
	"context"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"pc_book/pd"
	"pc_book/sample"
	"pc_book/service"
	"testing"
)

func TestServerCreateLaptop(t *testing.T) {
	t.Parallel()

	laptopNoId := sample.NewLaptop()
	laptopNoId.Id = ""

	laptopInvalidId := sample.NewLaptop()
	laptopInvalidId.Id = "invalid-id"

	laptopDuplicateId := sample.NewLaptop()
	storeDuplicateId := service.NewInMemoryLaptopStore()
	err := storeDuplicateId.Save(laptopDuplicateId)
	require.Nil(t, err)

	testCase := []struct {
		name 	string
		laptop 	*pd.Laptop
		store 	service.LaptopStore
		code 	codes.Code
	}{
		{
			name: "success_with_id",
			laptop: sample.NewLaptop(),
			store: service.NewInMemoryLaptopStore(),
			code: codes.OK,
		},
		{
			name: "success_no_id",
			laptop: laptopNoId,
			store: service.NewInMemoryLaptopStore(),
			code: codes.OK,
		},
		{
			name: "failure_invalid_id",
			laptop: laptopInvalidId,
			store: service.NewInMemoryLaptopStore(),
			code: codes.InvalidArgument,
		},
		{
			name: "failure_duplicate_id",
			laptop: laptopDuplicateId,
			store: storeDuplicateId,
			code: codes.AlreadyExists,
		},
	}

	for i := range testCase {
		tc := testCase[i]

		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			req := &pd.CreateLaptopRequest{
				Laptop: tc.laptop,
			}

			server := service.NewLaptopService(tc.store)
			res, err := server.CreateLaptop(context.Background(), req)

			if tc.code == codes.OK {
				require.NoError(t, err)
				require.NotNil(t, res)

				require.NotEmpty(t, res.Id)
				if len(tc.laptop.Id) > 0 {
					require.Equal(t, tc.laptop.Id, res.Id)
				}
			} else {
				require.Error(t, err)
				require.Nil(t, res)
				st, ok := status.FromError(err)
				require.True(t, ok)
				require.Equal(t, tc.code, st.Code())
			}
		})
	}
}






