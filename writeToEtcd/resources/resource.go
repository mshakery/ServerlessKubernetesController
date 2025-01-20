package resources

import (
	"context"
	"github.com/golang/protobuf/proto"
	"github.com/mshakery/ServerlessController/etcdMiddleware"
	etcd "github.com/mshakery/ServerlessController/etcdMiddleware"
)

type Resource interface {
	GetKVs() map[string]proto.Message
	CreatePostHook(ctx context.Context) bool
}

func CreateResourceInEtcd(ctx context.Context, r Resource) bool {
	client, err := etcd.ConnectToEtcd()
	if err != nil {
		panic("Could not connect to etcd. ")
	}
	KVs := r.GetKVs()
	for k, v := range KVs {
		err := etcdMiddleware.WriteToEtcdFromPb(client, ctx, k, v)
		if err != nil {
			panic(err)
		}
	}
	return r.CreatePostHook(ctx)
}
