package client

import (
	"context"
	"google.golang.org/grpc"
	"grpc-go/pb"
	"time"
)

type AuthClient struct {
	service  pb.AuthServiceClient
	username string
	password string
}

func NewAuthClient(cc *grpc.ClientConn, username, password string) *AuthClient {
	service := pb.NewAuthServiceClient(cc)
	return &AuthClient{
		service:  service,
		username: username,
		password: password,
	}
}

func (client *AuthClient) Login() (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*500000)
	defer cancel()
	req := &pb.LoginRequest{
		Username: client.username,
		Password: client.password,
	}
	resp, err := client.service.Login(ctx, req)
	if err != nil {
		return "", err
	}
	return resp.GetAccessToken(), nil
}
