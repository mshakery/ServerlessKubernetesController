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
	ctx := context.Background()
	read, err := etcd.ReadFromEtcd(client, ctx, "/cluster/resources/hpa", true)
	if err != nil {
	}
	for _, kv := range read.Kvs {
		HpaNamespace := strings.Split(string(kv.Key), "/")[4]
		HpaName := path.Base(string(kv.Key))

		conn, err2 := grpc.NewClient("horizontal-pod-autoscaler.default.10.101.174.165.sslip.io:80", grpc.WithTransportCredentials(insecure.NewCredentials()))

		if err2 != nil {
			log.Fatalf("hpa: could not connect 1: %v", err2)
		}
		c := protos.NewHorizontalPodAutoscalerClient(conn)

		in := protos.HpaName{Name: HpaName, Namespace: HpaNamespace}
		_, err3 := c.Scale(context.Background(), &in)

		if err3 != nil {
			log.Fatalf("hpa: could not connect 2: %v", err3)
		}
		log.Printf("hpa is run for hpa %s %s\n", HpaName, HpaNamespace)
	}
	fmt.Println("Time took to run nodeCheckerCronjob function:", time.Now().UnixNano()-startTime)
}
