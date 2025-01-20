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
	read, err := etcd.ReadFromEtcd(client, ctx, "/cluster/resources/node/", true)
	if err != nil {
	}
	for _, kv := range read.Kvs {
		if strings.Count(string(kv.Key), "/") == 4 {
			NodeName := path.Base(string(kv.Key))
			conn, err2 := grpc.NewClient("metric-collector.default.10.101.174.165.sslip.io:80", grpc.WithTransportCredentials(insecure.NewCredentials()))

			if err2 != nil {
				log.Fatalf("metrics: could not connect 1: %v", err2)
			}
			c := protos.NewMetricCollectorClient(conn)

			in := protos.NodeName{Name: NodeName}
			_, err3 := c.GatherMetric(context.Background(), &in)

			if err3 != nil {
				log.Fatalf("metrics: could not connect 2: %v", err3)
			}
			log.Printf("metrics are collected for node %s\n", NodeName)
			fmt.Println("Time since start:", time.Now().UnixNano()-startTime)
		}
	}
	fmt.Println("Time took to run entire function:", time.Now().UnixNano()-startTime)
}
