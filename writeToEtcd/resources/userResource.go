package resources

import (
	"context"
	"fmt"
	"github.com/golang/protobuf/proto"
	"github.com/mshakery/ServerlessController/protos"
)

type UserResource struct {
	User protos.User
}

func (ur *UserResource) GetKVs() map[string]proto.Message {
	KVs := make(map[string]proto.Message)
	key := fmt.Sprintf("/cluster/resources/user/%s", ur.User.Token)
	KVs[key] = &ur.User
	return KVs
}

func (ur *UserResource) CreatePostHook(ctx context.Context) bool {
	return true
}
