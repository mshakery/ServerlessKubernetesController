package main

import (
	"context"
	"flag"
	"fmt"
	protos "github.com/mshakery/ServerlessController/protos"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"log"
	"net"
	"strconv"
	"time"
)

var (
	port = flag.Int("port", 50051, "Metric Port")
)

type server struct {
	protos.UnimplementedKubeletServer
}

var podsToRun []*protos.Pod

func (s *server) RunAPod(ctx context.Context, pod *protos.Pod) (*protos.Empty, error) {
	// command
	podsToRun = append(podsToRun, pod)
	fmt.Printf(strconv.Itoa(len(podsToRun)))
	return &protos.Empty{}, nil
}

func (s *server) DeleteAPod(ctx context.Context, in *protos.Pod) (*protos.Empty, error) {
	// command
	idx := -1
	for i, pod := range podsToRun {
		if pod.GetMetadata().GetNamespace() == in.GetMetadata().GetNamespace() && pod.GetMetadata().GetName() == in.GetMetadata().GetName() {
			idx = i
		}
	}
	if idx == -1 {
		podsToRun = append(podsToRun[:idx], podsToRun[idx+1:]...)
	}
	return &protos.Empty{}, nil
}

func (s *server) Metric(ctx context.Context, in *protos.Empty) (*protos.NodeMetrics, error) {
	startTime := time.Now().UnixNano()
	metrics := protos.NodeMetrics{}
	var result []*protos.PodMetrics
	for _, pod := range podsToRun {
		pm := protos.PodMetrics{}
		pm.Name = pod.GetMetadata().GetName()
		pm.Namespace = pod.GetMetadata().GetNamespace()
		CpuUsage, MemoryUsage := 0, 0
		for _, container := range pod.Spec.Containers {
			CCpuUsage, _ := strconv.Atoi(container.GetResources().Requests["cpu"])
			CMemoryUsage, _ := strconv.Atoi(container.GetResources().Requests["memory"])
			CpuUsage += CCpuUsage
			MemoryUsage += CMemoryUsage
		}
		// TODO: it must be the maximum usage and request
		pm.CpuUsage = strconv.Itoa(CpuUsage)
		pm.MemoryUsage = strconv.Itoa(MemoryUsage)
		result = append(result, &pm)
	}
	metrics.PodMetrics = result
	fmt.Println("Time took to run function:", time.Now().UnixNano()-startTime)
	return &metrics, nil
}

func main() {
	flag.Parse()
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", *port))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	s := grpc.NewServer()
	protos.RegisterKubeletServer(s, &server{})
	reflection.Register(s)
	log.Printf("server listening at %v", lis.Addr())
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
