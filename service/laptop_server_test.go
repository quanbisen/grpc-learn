package service_test

import (
	"context"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"grpc-go/pb"
	"grpc-go/sample"
	"grpc-go/service"
	"testing"
)

func TestLaptopServer_CreateLaptop(t *testing.T) {
	t.Parallel()
	laptopNoID := sample.NewLaptop()
	laptopNoID.Id = ""

	laptopInvalidID := sample.NewLaptop()
	laptopInvalidID.Id = "invalid-uuid"

	laptopDuplicateID := sample.NewLaptop()
	storeDuplicateID := service.NewInMemoryLaptopStore()
	err := storeDuplicateID.Save(laptopDuplicateID)
	require.Nil(t, err)

	testCase := []struct {
		name        string
		laptop      *pb.Laptop
		store       service.LaptopStore
		imageStore  service.ImageStore
		ratingStore service.RateStore
		code        codes.Code
	}{
		{
			name:        "success_with_id",
			laptop:      sample.NewLaptop(),
			store:       service.NewInMemoryLaptopStore(),
			imageStore:  service.NewDiskImageStore("img"),
			ratingStore: service.NewInMemoryRatingStore(),
			code:        codes.OK,
		},
		{
			name:        "success_no_id",
			laptop:      laptopNoID,
			store:       service.NewInMemoryLaptopStore(),
			imageStore:  service.NewDiskImageStore("img"),
			ratingStore: service.NewInMemoryRatingStore(),
			code:        codes.OK,
		},
		{
			name:        "failure_invalid_id",
			laptop:      laptopInvalidID,
			store:       service.NewInMemoryLaptopStore(),
			imageStore:  service.NewDiskImageStore("img"),
			ratingStore: service.NewInMemoryRatingStore(),
			code:        codes.InvalidArgument,
		},
		{
			name:        "failure_duplicate_id",
			laptop:      laptopDuplicateID,
			store:       storeDuplicateID,
			imageStore:  service.NewDiskImageStore("img"),
			ratingStore: service.NewInMemoryRatingStore(),
			code:        codes.AlreadyExists,
		},
	}

	for i := range testCase {
		tc := testCase[i]
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			req := &pb.CreateLaptopRequest{Laptop: tc.laptop}
			server := service.NewLaptopServer(tc.store, tc.imageStore, tc.ratingStore)
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
