package resources

import (
	"context"
	"fmt"
	"github.com/golang/protobuf/proto"
	"github.com/mshakery/ServerlessController/protos"
)

type RoleBindingResource struct {
	RoleBinding protos.RoleBinding
}

func (rb *RoleBindingResource) GetKVs() map[string]proto.Message {
	KVs := make(map[string]proto.Message)
	key := fmt.Sprintf("/cluster/resources/role_binding/%s/%s", rb.RoleBinding.Metadata.Namespace, rb.RoleBinding.Metadata.Name)
	KVs[key] = &rb.RoleBinding
	return KVs
}

func (rb *RoleBindingResource) CreatePostHook(ctx context.Context) bool {
	return true
}
