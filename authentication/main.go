package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/golang/protobuf/proto"
	etcd "github.com/mshakery/ServerlessController/etcdMiddleware"
	protos "github.com/mshakery/ServerlessController/protos"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/reflection"
	"log"
	"net"
	"time"
)

var (
	port = flag.Int("port", 50051, "The server port")
)

type server struct {
	protos.UnimplementedAuthenticationServer
}

func (s *server) Auth(ctx context.Context, in *protos.AuthenticationRequest) (*protos.Response, error) {
	startTime := time.Now().UnixNano()
	/*
		should reade from ETCD. key is token. checks if value exists.
		if exists, value is uid.
		then makes request for authorization.

		creates Authorizationrequest. uid = uid from etcd.
		then calls authorizationrequest.
	*/
	client, err := etcd.ConnectToEtcd()
	if err != nil {
		panic("Could not connect to etcd. ")
	}
	defer client.Close()

	craftedKey := fmt.Sprintf("/cluster/resources/user/%s", in.Token)
	read, err := etcd.ReadFromEtcd(client, ctx, craftedKey, false)
	if err != nil {
	}

	resp := protos.Response{}

	if read.Count == 0 {
		resp.Code = -1
		resp.Status = "Token does not exist."
		return &resp, nil
	} else {
		for _, kv := range read.Kvs {
			// fmt.Printf("k: %s, v: %s\n", kv.Key, kv.Value)
			resp.Code = 0
			resp.Status = "User exists."

			user := protos.User{}
			err := proto.Unmarshal(kv.Value, &user)
			if err != nil {
				return nil, err
			}

			fmt.Println("Time took to run function:", time.Now().UnixNano()-startTime)
			authResp, err := callAuthorization(ctx, in.ClientRequest, user.Uid)
			if err != nil {
				return nil, err
			}
			return authResp, nil
		}
	}
	return nil, nil
}

func callAuthorization(ctx context.Context, client_request *protos.ClientRequest, uid string) (*protos.Response, error) {
	conn, err := grpc.NewClient("authorization.default.10.101.174.165.sslip.io:80", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("could not connect: %v", err)
	}
	defer conn.Close()
	c := protos.NewAuthorizationClient(conn)

	resp, err := c.Authorize(ctx, &protos.AuthorizationRequest{Uid: uid, ClientRequest: client_request})
	if err != nil {
		log.Fatalf("could not connect: %v", err)
	}
	return resp, nil
}

func main() {
	flag.Parse()
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", *port))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	s := grpc.NewServer()
	protos.RegisterAuthenticationServer(s, &server{})
	reflection.Register(s)
	log.Printf("server listening at %v", lis.Addr())
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
