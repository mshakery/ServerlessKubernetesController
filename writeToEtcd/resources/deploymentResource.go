package resources

import (
	"context"
	"fmt"
	"github.com/golang/protobuf/proto"
	"github.com/google/uuid"
	"github.com/mshakery/ServerlessController/protos"
	"sync"
)

type DeploymentResource struct {
	Deployment protos.Deployment
}

func (dr *DeploymentResource) GetKVs() map[string]proto.Message {
	var podNames []string
	for i := int32(0); i < dr.Deployment.GetSpec().GetReplicas(); i++ {
		newUuid := uuid.New().String()
		newPodName := fmt.Sprintf("%s-%s", dr.Deployment.Metadata.Name, newUuid)
		podNames = append(podNames, newPodName)
	}
	dr.Deployment.Status.Replicas = dr.Deployment.GetSpec().GetReplicas()
	dr.Deployment.Status.PodNames = podNames
	KVs := make(map[string]proto.Message)
	key := fmt.Sprintf("/cluster/resources/Deployment/%s/%s", dr.Deployment.Metadata.Namespace, dr.Deployment.Metadata.Name)
	KVs[key] = &dr.Deployment
	key = fmt.Sprintf("/cluster/resources/Deployment/%s/%s/status", dr.Deployment.Metadata.Namespace, dr.Deployment.Metadata.Name)
	KVs[key] = dr.Deployment.GetStatus()
	return KVs
}

func (dr *DeploymentResource) CreatePostHook(ctx context.Context) bool {
	var pods []protos.Pod
	for _, podName := range dr.Deployment.GetStatus().GetPodNames() {
		newPod := protos.Pod{}
		newPod.Metadata = dr.Deployment.GetMetadata()
		newPod.Metadata.Name = podName
		newPod.Metadata.Uid = podName
		newPod.Spec = dr.Deployment.GetSpec().GetTemplate().GetSpec()
		newPod.Status = &protos.PodStatus{}
		newPod.Status.Worker = &protos.Worker{Worker: "-1"}
		newPod.Status.ResourceUsage = &protos.ResourceUsage{ResourceUsage: make(map[string]string)}
		pods = append(pods, newPod)
	}

	var wg sync.WaitGroup
	resultChan, result := make(chan bool), true

	for _, pod := range pods {
		wg.Add(1)
		go func(pod protos.Pod) {
			defer wg.Done()
			podResource := PodResource{Pod: pod}
			success := CreateResourceInEtcd(ctx, &podResource)
			resultChan <- success
		}(pod)
	}

	go func() {
		wg.Wait()
		close(resultChan)
	}()
	for success := range resultChan {
		result = result && success
	}

	return result
}
