package resources

import (
	"context"
	"fmt"
	"github.com/golang/protobuf/proto"
	"github.com/mshakery/ServerlessController/protos"
)

type RoleResource struct {
	Role protos.Role
}

func (rr *RoleResource) GetKVs() map[string]proto.Message {
	KVs := make(map[string]proto.Message)
	key := fmt.Sprintf("/cluster/resources/role/%s/%s", rr.Role.Metadata.Namespace, rr.Role.Metadata.Name)
	KVs[key] = &rr.Role
	return KVs
}

func (rr *RoleResource) CreatePostHook(ctx context.Context) bool {
	return true
}
