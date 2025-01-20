package main

import (
	"context"
	"fmt"
	etcd "github.com/mshakery/ServerlessController/etcdMiddleware"
	"github.com/mshakery/ServerlessController/protos"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"log"
	"path"
	"strings"
	"time"
)

func main() {
	startTime := time.Now().UnixNano()
	client, err := etcd.ConnectToEtcd()
	if err != nil {
		panic("Could not connect to etcd. ")
	}
	defer client.Close()
	ctx := context.TODO()
	read, err := etcd.ReadFromEtcd(client, ctx, "/cluster/resources/node/", true)
	if err != nil {
	}
	for _, kv := range read.Kvs {
		if strings.Count(string(kv.Key), "/") == 4 {
			NodeName := path.Base(string(kv.Key))
			conn, err2 := grpc.NewClient("node-checker.default.10.101.174.165.sslip.io:80", grpc.WithTransportCredentials(insecure.NewCredentials()))

			if err2 != nil {
				log.Fatalf("node-checker: could not connect 1: %v", err2)
			}
			c := protos.NewNodeCheckerClient(conn)

			in := protos.NodeNamee{Name: NodeName}
			_, err3 := c.CheckNode(context.Background(), &in)

			if err3 != nil {
				log.Fatalf("node-checker: could not connect 2: %v", err3)
			}
			log.Printf("node %s checked\n", NodeName)
		}
	}
	fmt.Println("Time took to run nodeCheckerCronjob function:", time.Now().UnixNano()-startTime)
}
