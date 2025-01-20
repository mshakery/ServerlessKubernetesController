package resources

import (
	"context"
	"fmt"
	"github.com/golang/protobuf/proto"
	"github.com/mshakery/ServerlessController/protos"
)

type HpaResource struct {
	Hpa protos.Hpa
}

func (hr *HpaResource) GetKVs() map[string]proto.Message {
	KVs := make(map[string]proto.Message)
	key := fmt.Sprintf("/cluster/resources/hpa/%s/%s", hr.Hpa.GetMetadata().GetNamespace(), hr.Hpa.GetMetadata().GetName())
	KVs[key] = &hr.Hpa
	return KVs
}

func (hr *HpaResource) CreatePostHook(ctx context.Context) bool {
	return true
}
