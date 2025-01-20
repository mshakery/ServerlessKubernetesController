package main

import (
	"context"
	"flag"
	"fmt"
	protos "github.com/mshakery/ServerlessController/protos"
	"github.com/mshakery/ServerlessController/writeToEtcd/resources"
	"google.golang.org/grpc"
	"log"
	"net"
	"time"
)

var (
	port = flag.Int("port", 50051, "The server port")
)

type server struct {
	protos.UnimplementedWriteToEtcdServer
}

func ResourceFactory(req *protos.ClientRequest) resources.Resource {
	switch req.OneofResource.(type) {
	case *protos.ClientRequest_Deployment:
		return &resources.DeploymentResource{Deployment: *req.GetDeployment()}
	case *protos.ClientRequest_Role:
		return &resources.RoleResource{Role: *req.GetRole()}
	case *protos.ClientRequest_RoleBinding:
		return &resources.RoleBindingResource{RoleBinding: *req.GetRoleBinding()}
	case *protos.ClientRequest_User:
		return &resources.UserResource{User: *req.GetUser()}
	case *protos.ClientRequest_Node:
		return &resources.NodeResource{Node: *req.GetNode()}
	case *protos.ClientRequest_Pod:
		return &resources.PodResource{Pod: *req.GetPod()}
	case *protos.ClientRequest_Hpa:
		return &resources.HpaResource{Hpa: *req.GetHpa()}
	}
	return nil
}

func (s *server) Apply(ctx context.Context, in *protos.ApplyRequest) (*protos.Response, error) {
	startTime := time.Now().UnixNano()
	resource := ResourceFactory(in.GetClientRequest())
	result := resources.CreateResourceInEtcd(ctx, resource)
	resp := protos.Response{Code: 0, Status: "Success"}
	if !result {
		resp.Code = -1
		resp.Status = "Cannot Write to ETCD"
	}
	fmt.Println("Time took to run function:", time.Now().UnixNano()-startTime)
	return &resp, nil
}

func main() {
	flag.Parse()
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", *port))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	s := grpc.NewServer()
	protos.RegisterWriteToEtcdServer(s, &server{})
	log.Printf("server listening at %v", lis.Addr())
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
