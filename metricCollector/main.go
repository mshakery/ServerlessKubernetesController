package main

import (
	"context"
	"flag"
	"fmt"
	etcd "github.com/mshakery/ServerlessController/etcdMiddleware"
	protos "github.com/mshakery/ServerlessController/protos"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/reflection"
	"log"
	"net"
	"strconv"
	"time"
)

var (
	port = flag.Int("port", 50051, "The server port")
)

type server struct {
	protos.UnimplementedMetricCollectorServer
}

func (s *server) GatherMetric(ctx context.Context, in *protos.NodeName) (*protos.Empty, error) {
	startTime := time.Now().UnixNano()
	host := fmt.Sprintf("%s:50051", in.GetName())

	conn, err2 := grpc.NewClient(host, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err2 != nil {
		log.Printf("kubelet node: could not connect: %v", err2)
		return &protos.Empty{}, nil
	}
	client, err := etcd.ConnectToEtcd()
	if err != nil {
		log.Printf("kubelet node: could not connect: %v", err)
		return &protos.Empty{}, nil
	}

	c := protos.NewKubeletClient(conn)

	inp := protos.Empty{}
	response, err3 := c.Metric(context.Background(), &inp)

	if err3 != nil {
		log.Printf("metrics: could not connect: %v", err3)
		return &protos.Empty{}, nil
	}
	nodeCpuUsage, nodeMemoryUsage := 0, 0
	for _, podMetric := range response.GetPodMetrics() {
		podKey := fmt.Sprintf("/cluster/resources/pod/%s/%s/usage", podMetric.GetNamespace(), podMetric.GetName())
		ru := protos.ResourceUsage{ResourceUsage: make(map[string]string)}
		ru.ResourceUsage["cpu"] = podMetric.GetCpuUsage()
		ru.ResourceUsage["memory"] = podMetric.GetMemoryUsage()

		err = etcd.WriteToEtcdFromPb(client, ctx, podKey, &ru)
		if err != nil {
			return nil, err
		}
		podCpuUsageInt, _ := strconv.Atoi(podMetric.GetCpuUsage())
		podMemoryUsageInt, _ := strconv.Atoi(podMetric.GetMemoryUsage())
		nodeCpuUsage += podCpuUsageInt
		nodeMemoryUsage += podMemoryUsageInt
	}
	nodeKey := fmt.Sprintf("/cluster/resources/node/%s", in.GetName())
	NodeAllocatableKey := fmt.Sprintf("/cluster/resources/node/%s/allocatable", in.GetName())
	NodeLastUpdateKey := fmt.Sprintf("/cluster/resources/node/%s/update/time", in.GetName())
	var nodeObject = protos.Node{}
	err = etcd.ReadOneFromEtcdToPb(client, ctx, nodeKey, &nodeObject)
	if err != nil {
		return nil, err
	}
	nodeFreeResource := nodeObject.GetStatus().GetCapacity()
	nodeCpuCapacity, _ := strconv.Atoi(nodeFreeResource.Resources["cpu"])
	nodeMemoryCapacity, _ := strconv.Atoi(nodeFreeResource.Resources["memory"])
	nodeFreeResource.Resources["cpu"] = strconv.Itoa(nodeCpuCapacity - nodeCpuUsage)
	nodeFreeResource.Resources["memory"] = strconv.Itoa(nodeMemoryCapacity - nodeMemoryUsage)
	err = etcd.WriteToEtcdFromPb(client, ctx, NodeAllocatableKey, nodeFreeResource)
	if err != nil {
		return nil, err
	}
	err = etcd.WriteToEtcd(client, ctx, NodeLastUpdateKey, time.Now().Format("2006-01-02 15:04:05"))
	if err != nil {
		return nil, err
	}
	fmt.Println("Time took to run function:", time.Now().UnixNano()-startTime)
	return &protos.Empty{}, nil
}

func main() {
	flag.Parse()
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", *port))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	s := grpc.NewServer()
	protos.RegisterMetricCollectorServer(s, &server{})
	reflection.Register(s)
	log.Printf("server listening at %v", lis.Addr())
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
