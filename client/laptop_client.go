package client

import (
	"bufio"
	"context"
	"fmt"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"grpc-go/pb"
	"io"
	"log"
	"os"
	"path/filepath"
	"time"
)

type LaptopClient struct {
	service pb.LaptopServiceClient
}

func NewLaptopClient(conn *grpc.ClientConn) *LaptopClient {
	service := pb.NewLaptopServiceClient(conn)
	return &LaptopClient{
		service: service,
	}
}

func (client *LaptopClient) CreateLaptop(laptop *pb.Laptop) {
	req := &pb.CreateLaptopRequest{
		Laptop: laptop,
	}

	// set timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	res, err := client.service.CreateLaptop(ctx, req)
	if err != nil {
		st, ok := status.FromError(err)
		if ok && st.Code() == codes.AlreadyExists {
			// not a big deal
			log.Print("laptop already exists")
		} else {
			log.Fatal("cannot create laptop: ", err)
		}
		return
	}
	log.Printf("create laptop with id: %s\n", res.Id)
}

func (client *LaptopClient) SearchLaptop(filter *pb.Filter) {
	log.Printf("search filter: %v", filter)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*500000000)
	defer cancel()

	req := &pb.SearchLaptopRequest{
		Filter: filter,
	}
	stream, err := client.service.SearchLaptop(ctx, req)
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
		log.Print("- found: ", laptop.GetId())
		log.Print("+ brand: ", laptop.GetBrand())
		log.Print("+ name: ", laptop.GetName())
		log.Print("+ cpu cores: ", laptop.GetCpu().GetNumberCores())
		log.Print("+ cpu min ghz: ", laptop.GetCpu().GetMinGhz())
		log.Print("+ ram: ", laptop.GetRam())
		log.Print("+ price: ", laptop.GetPriceUsd())
	}
}

func (client *LaptopClient) UploadImage(laptopId string, filename string) {
	file, err := os.Open(filename)
	if err != nil {
		log.Fatal("cannot open image file: ", err)
	}
	defer file.Close()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*600)
	defer cancel()
	req := &pb.UploadImageRequest{
		Data: &pb.UploadImageRequest_Info{
			Info: &pb.ImageInfo{
				LaptopId:  laptopId,
				ImageType: filepath.Ext(filename),
			},
		},
	}
	stream, err := client.service.UploadImage(ctx)
	if err != nil {
		log.Fatal("cannot upload image: ", err)
	}
	err = stream.Send(req)
	if err != nil {
		log.Fatal("cannot send image info to server: ", err, stream.RecvMsg(nil))
	}
	reader := bufio.NewReader(file)
	buffer := make([]byte, 102400)
	for {
		n, err := reader.Read(buffer)
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatal("cannot read chunk to buffer: ", err)
		}
		req := &pb.UploadImageRequest{
			Data: &pb.UploadImageRequest_ChunkData{
				ChunkData: buffer[:n],
			},
		}
		err = stream.Send(req)
		if err != nil {
			log.Fatal("cannot send chunk to server: ", err)
		}
	}
	res, err := stream.CloseAndRecv()
	if err != nil {
		log.Fatal("cannot receive response: ", err)
	}
	log.Printf("image upload with id: %s, size: %d", res.GetId(), res.GetSize())
}

func (client *LaptopClient) RateLaptop(laptopIds []string, scores []float64) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*500000)
	defer cancel()

	stream, err := client.service.RateLaptop(ctx)
	if err != nil {
		return fmt.Errorf("cannot rate laptop: %v", err)
	}

	waitResponse := make(chan error)

	go func() {
		for {
			res, err := stream.Recv()
			if err == io.EOF {
				log.Println("no more response")
				waitResponse <- nil
				return
			}
			if err != nil {
				waitResponse <- fmt.Errorf("cannot receive stream response: %v", err)
				return
			}

			log.Println("receive response: ", res)
		}
	}()

	for i, laptopId := range laptopIds {
		req := &pb.RateLaptopRequest{
			LaptopId: laptopId,
			Score:    scores[i],
		}

		err = stream.Send(req)
		if err != nil {
			return fmt.Errorf("cannot send stream request: %v - %v", err, stream.RecvMsg(nil))
		}
		log.Println("send request: ", req)
	}

	err = stream.CloseSend()
	if err != nil {
		return fmt.Errorf("cannot close send: %v", err)
	}

	err = <-waitResponse
	return err
}
