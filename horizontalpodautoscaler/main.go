package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/google/uuid"
	etcd "github.com/mshakery/ServerlessController/etcdMiddleware"
	protos "github.com/mshakery/ServerlessController/protos"
	clientv3 "go.etcd.io/etcd/client/v3"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/reflection"
	"log"
	"net"
	"strconv"
	"sync"
	"time"
)

var (
	port = flag.Int("port", 50051, "The server port")
)

type server struct {
	protos.UnimplementedHorizontalPodAutoscalerServer
}

func Abs(x float32) float32 {
	if x < 0 {
		return -x
	}
	return x
}

func CalculateAverageResourceUsage(client *clientv3.Client, ctx context.Context, namespace string, podNames []string, resourceType string) float32 {
	sum := 0
	count := 0
	for _, podName := range podNames {
		key := fmt.Sprintf("/cluster/resources/pod/%s/%s/usage", namespace, podName)
		var ru protos.ResourceUsage
		err := etcd.ReadOneFromEtcdToPb(client, ctx, key, &ru)
		if err != nil {
			continue
		}
		i, err := strconv.Atoi(ru.GetResourceUsage()[resourceType])
		if err != nil {
			// ... handle error
			//log.Fatalf("str to int error: %v", err)
			continue
		}
		sum += i
		count += 1
	}
	return float32(sum) / float32(count)
}

func (s *server) Scale(ctx context.Context, in *protos.HpaName) (*protos.Empty, error) {
	startTime := time.Now().UnixNano()
	conn, err := grpc.NewClient("write-to-etcd.default.10.101.174.165.sslip.io:80", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("could not connect: %v", err)
	}
	defer conn.Close()
	clientWriteToEtcd := protos.NewWriteToEtcdClient(conn)

	client, err := etcd.ConnectToEtcd()
	if err != nil {
		panic("Could not connect to etcd. ")
	}
	var hpa protos.Hpa
	key := fmt.Sprintf("/cluster/resources/hpa/%s/%s", in.GetNamespace(), in.GetName())
	err = etcd.ReadOneFromEtcdToPb(client, ctx, key, &hpa)
	if err != nil {
		log.Fatalf("cant read from etcd: %v", err)
	}
	deploymentStatusKey := fmt.Sprintf("/cluster/resources/Deployment/%s/%s/status", in.GetNamespace(), hpa.GetSpec().GetTargetResource())
	var deploymentStatus protos.DeploymentStatus
	err = etcd.ReadOneFromEtcdToPb(client, ctx, deploymentStatusKey, &deploymentStatus)
	if err != nil {
		log.Fatalf("cant read from etcd: %v", err)
	}
	deploymentKey := fmt.Sprintf("/cluster/resources/Deployment/%s/%s", in.GetNamespace(), hpa.GetSpec().GetTargetResource())
	var deployment protos.Deployment
	err = etcd.ReadOneFromEtcdToPb(client, ctx, deploymentKey, &deployment)
	if err != nil {
		log.Fatalf("cant read from etcd: %v", err)
	}
	resourceAverage := CalculateAverageResourceUsage(client, ctx, in.GetNamespace(), deploymentStatus.PodNames, hpa.GetSpec().GetMetrics().GetMetricType())
	targetValue := resourceAverage / hpa.GetSpec().GetMetrics().GetTargetValue()
	if Abs(targetValue/hpa.GetSpec().GetMetrics().GetTargetAverageUtilization()) <= 0.9 || Abs(targetValue/hpa.GetSpec().GetMetrics().GetTargetAverageUtilization()) > 1.1 {
		newReplica := int32(float32(deploymentStatus.GetReplicas()) * (targetValue / hpa.GetSpec().GetMetrics().GetTargetAverageUtilization()))
		if newReplica > hpa.GetSpec().GetMaxReplicas() {
			newReplica = hpa.GetSpec().GetMaxReplicas()
		}
		if newReplica < hpa.GetSpec().GetMinReplicas() {
			newReplica = hpa.GetSpec().GetMinReplicas()
		}
		currentReplica := deploymentStatus.Replicas
		currrentPods := deploymentStatus.PodNames

		wg := new(sync.WaitGroup)
		if newReplica > deploymentStatus.Replicas {
			for i := int32(0); i < newReplica-deploymentStatus.Replicas; i++ {
				currentReplica += 1
				newUuid := uuid.New().String()
				newPodName := fmt.Sprintf("%s-%s", hpa.GetSpec().GetTargetResource(), newUuid)
				currrentPods = append(currrentPods, newPodName)
				newPod := protos.Pod{}
				newPod.Metadata = deployment.GetMetadata()
				newPod.Metadata.Name = newPodName
				newPod.Metadata.Uid = newPodName
				newPod.Spec = deployment.GetSpec().GetTemplate().GetSpec()
				newPod.Status = &protos.PodStatus{}
				newPod.Status.Worker = &protos.Worker{Worker: "-1"}
				newPod.Status.ResourceUsage = &protos.ResourceUsage{ResourceUsage: make(map[string]string)}

				applyReq := &protos.ApplyRequest{
					ClientRequest: &protos.ClientRequest{Operation: "create", OneofResource: &protos.ClientRequest_Pod{Pod: &newPod}},
				}
				etcd.WriteToEtcdFromPb(client, ctx, deploymentStatusKey, &protos.DeploymentStatus{Replicas: currentReplica, PodNames: currrentPods})
				wg.Add(1)
				go func(ctx context.Context, applyReq *protos.ApplyRequest) {
					defer wg.Done()
					clientWriteToEtcd.Apply(ctx, applyReq)
				}(ctx, applyReq)
			}
		} else if newReplica < deploymentStatus.Replicas {
			for i := int32(0); i < deploymentStatus.Replicas-newReplica; i++ {
				removedPod := currrentPods[len(currrentPods)-1]
				currentReplica -= 1
				currrentPods = currrentPods[:len(currrentPods)-1]
				err = etcd.WriteToEtcdFromPb(client, ctx, deploymentStatusKey, &protos.DeploymentStatus{Replicas: currentReplica, PodNames: currrentPods})
				if err != nil {
					log.Fatalf("cant write to etcd: %v", err)
				}

				wg.Add(1)
				go func() {
					defer wg.Done()
					podWorkerKey := fmt.Sprintf("/cluster/resources/pod/%s/%s/worker", in.GetNamespace(), removedPod)
					var podWorker protos.Worker
					err = etcd.ReadOneFromEtcdToPb(client, ctx, podWorkerKey, &podWorker)
					if err != nil {
						return
					}
					workerKey := fmt.Sprintf("%s:50051", podWorker.GetWorker())
					conn, err := grpc.NewClient(workerKey, grpc.WithTransportCredentials(insecure.NewCredentials()))
					if err != nil {
						log.Fatalf("could not connect: %v", err)
					}
					defer conn.Close()
					clientKubelet := protos.NewKubeletClient(conn)
					clientKubelet.DeleteAPod(ctx, &protos.Pod{Metadata: &protos.Metadata{Name: removedPod, Namespace: in.GetNamespace()}})
					podKey := fmt.Sprintf("/cluster/resources/pod/%s/%s", in.GetNamespace(), removedPod)
					_, err = client.Delete(ctx, podKey, clientv3.WithPrefix())
					if err != nil {
						return
					}
				}()
			}
		}
		wg.Wait()
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
	protos.RegisterHorizontalPodAutoscalerServer(s, &server{})
	reflection.Register(s)
	log.Printf("server listening at %v", lis.Addr())
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
