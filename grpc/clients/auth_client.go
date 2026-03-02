package clients

import (
	"context"
	"log"
	"time"

	"github.com/Bimidu/ctse-order-service/internal/config"
	authpb "github.com/Bimidu/ctse-order-service/proto/auth"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type AuthClient struct {
	conn   *grpc.ClientConn
	client authpb.AuthServiceClient
}

var Auth *AuthClient

func InitAuthClient() {
	addr := config.App.AuthServiceAddr
	conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Printf("Warning: could not connect to auth service at %s: %v", addr, err)
		return
	}
	Auth = &AuthClient{
		conn:   conn,
		client: authpb.NewAuthServiceClient(conn),
	}
	log.Printf("Auth gRPC client connected to %s", addr)
}

func (a *AuthClient) VerifyToken(token string) (*authpb.VerifyTokenResponse, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	return a.client.VerifyToken(ctx, &authpb.VerifyTokenRequest{Token: token})
}

func (a *AuthClient) GetUserInfo(userID string) (*authpb.UserInfo, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	return a.client.GetUserInfo(ctx, &authpb.GetUserInfoRequest{UserId: userID})
}

func (a *AuthClient) UpdatePurchaseCount(userID string, increment int32) (*authpb.UpdatePurchaseCountResponse, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	return a.client.UpdatePurchaseCount(ctx, &authpb.UpdatePurchaseCountRequest{
		UserId:    userID,
		Increment: increment,
	})
}

func (a *AuthClient) Close() {
	if a.conn != nil {
		a.conn.Close()
	}
}
