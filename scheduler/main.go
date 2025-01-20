package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/golang/protobuf/proto"
	etcd "github.com/mshakery/ServerlessController/etcdMiddleware"
	protos "github.com/mshakery/ServerlessController/protos"
	clientv3 "go.etcd.io/etcd/client/v3"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"log"
	"net"
	"sort"
	"strconv"
	"strings"
	"time"
)

var (
	port = flag.Int("port", 50051, "The server port")
)

type server struct {
	protos.UnimplementedSchedulerServer
}

type NodeDetail struct {
	name       string
	freeCpu    int
	freeMemory int
}

func getNodesWithSufficientResource(ctx context.Context, cli *clientv3.Client, cpuMin int, memoryMin int) []NodeDetail {
	startTime := time.Now().UnixNano()
	var result []NodeDetail
	response, _ := etcd.ReadFromEtcd(cli, ctx, "/cluster/resources/node/", true)
	for _, kv := range response.Kvs {
		if strings.Count(string(kv.Key), "/") == 5 && strings.HasSuffix(string(kv.Key), "/allocatable") {
			nodeAllocatable := protos.Capacity{}
			nd := NodeDetail{}
			nd.name = strings.Split(string(kv.Key), "/")[4]
			err := proto.Unmarshal(kv.Value, &nodeAllocatable)
			if err != nil {
				panic(err)
			}
			nd.freeCpu, _ = strconv.Atoi(nodeAllocatable.Resources["cpu"])
			nd.freeMemory, _ = strconv.Atoi(nodeAllocatable.Resources["memory"])
			unschedulableKey := fmt.Sprintf("/cluster/resources/node/%s/unschedulable", nd.name)
			var unschedulable protos.Unschedulable
			etcd.ReadOneFromEtcdToPb(cli, ctx, unschedulableKey, &unschedulable)
			if nd.freeCpu >= cpuMin && nd.freeMemory >= memoryMin && unschedulable.Condition == false {
				result = append(result, nd)
			}
		}
	}
	sort.Slice(result, func(i, j int) bool {
		return result[i].freeMemory > result[j].freeMemory
	})
	fmt.Println("Time took to schedule a pod in ns:", time.Now().UnixNano()-startTime)
	return result
}

func calculatePodRequestedResource(pod *protos.Pod) (cpuReq int, memoryReq int) {
	cpuReq, memoryReq = 0, 0
	for _, container := range pod.GetSpec().GetContainers() {
		cCpuReq, _ := strconv.Atoi(container.Resources.Requests["cpu"])
		cMemoryReq, _ := strconv.Atoi(container.Resources.Requests["memory"])
		cpuReq += cCpuReq
		memoryReq += cMemoryReq
	}
	return cpuReq, memoryReq
}

func (s *server) Schedule(ctx context.Context, in *protos.PodDetail) (*protos.Empty, error) {
	startTime := time.Now().UnixNano()
	client, err := etcd.ConnectToEtcd()
	if err != nil {
		panic("Could not connect to etcd. ")
	}

	podKey := fmt.Sprintf("/cluster/resources/pod/%s/%s", in.GetNamespace(), in.GetName())
	var podObject protos.Pod
	etcd.ReadOneFromEtcdToPb(client, ctx, podKey, &podObject)
	cpuReq, memoryReq := calculatePodRequestedResource(&podObject)
	nodesWithSufficientResource := getNodesWithSufficientResource(ctx, client, cpuReq, memoryReq)
	if len(nodesWithSufficientResource) == 0 {
		log.Fatalf("no free node!")
		return &protos.Empty{}, nil
	}

	podWorkerKey := fmt.Sprintf("/cluster/resources/pod/%s/%s/worker", in.GetNamespace(), in.GetName())
	bestWorker := protos.Worker{Worker: nodesWithSufficientResource[0].name}
	etcd.WriteToEtcdFromPb(client, ctx, podWorkerKey, &bestWorker)

	host := fmt.Sprintf("%s:50051", bestWorker.GetWorker())
	conn, err2 := grpc.NewClient(host, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err2 != nil {
		log.Fatalf("kubelet node: could not connect: %v", err2)
	}
	c := protos.NewKubeletClient(conn)

	resp, err3 := c.RunAPod(context.Background(), &podObject)
	fmt.Printf("im feeling lucky %s", resp, &podObject)
	if err3 != nil {
		log.Fatalf("kubelet run a pod error: %v", err3)
	}
	fmt.Println("Time took to schedule a pod in ns:", time.Now().UnixNano()-startTime)
	return &protos.Empty{}, nil
}

func main() {
	flag.Parse()
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", *port))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	s := grpc.NewServer()
	protos.RegisterSchedulerServer(s, &server{})
	log.Printf("server listening at %v", lis.Addr())
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
