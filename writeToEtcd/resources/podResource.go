package resources

import (
	"context"
	"fmt"
	"github.com/golang/protobuf/proto"
	"github.com/mshakery/ServerlessController/protos"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"log"
)

type PodResource struct {
	Pod protos.Pod
}

func (pr *PodResource) GetKVs() map[string]proto.Message {
	KVs := make(map[string]proto.Message)
	key := fmt.Sprintf("/cluster/resources/pod/%s/%s", pr.Pod.Metadata.Namespace, pr.Pod.Metadata.Name)
	KVs[key] = &pr.Pod
	key = fmt.Sprintf("/cluster/resources/pod/%s/%s/usage", pr.Pod.Metadata.Namespace, pr.Pod.Metadata.Name)
	KVs[key] = pr.Pod.GetStatus().GetResourceUsage()
	key = fmt.Sprintf("/cluster/resources/pod/%s/%s/worker", pr.Pod.Metadata.Namespace, pr.Pod.Metadata.Name)
	KVs[key] = pr.Pod.GetStatus().GetWorker()
	return KVs
}

func (pr *PodResource) CreatePostHook(ctx context.Context) bool {
	conn, err2 := grpc.NewClient("scheduler.default.10.101.174.165.sslip.io:80", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err2 != nil {
		log.Fatalf("scheduler node: could not connect: %v", err2)
	}
	c := protos.NewSchedulerClient(conn)
	in := protos.PodDetail{Name: pr.Pod.GetMetadata().GetName(), Namespace: pr.Pod.GetMetadata().GetNamespace()}
	_, err := c.Schedule(ctx, &in)
	if err != nil {
		log.Fatalf("schedule error: %v", err)
	}
	return true
}
