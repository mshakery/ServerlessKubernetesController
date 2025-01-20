package etcdMiddleware

import (
	"context"
	"fmt"
	"github.com/golang/protobuf/proto"
	"go.etcd.io/etcd/api/v3/v3rpc/rpctypes"
	clientv3 "go.etcd.io/etcd/client/v3"
	"reflect"
	"time"
)

var endpoints = []string{"14.102.10.254:2379"}

func ConnectToEtcd() (*clientv3.Client, error) {
	cli, err := clientv3.New(clientv3.Config{
		Endpoints:   endpoints,
		DialTimeout: 5 * time.Second,
	})
	if err != nil {
		fmt.Println("Failed to connect to etcd:", err)
		return nil, err
	}
	return cli, nil
	/* dont forget to:
	defer func(cli *clientv3.Client) {
		err := cli.Close()
		if err != nil {
		}
	}(cli)
	*/
}

func WriteToEtcd(cli *clientv3.Client, ctx context.Context, key string, val string) error {
	_, err := cli.Put(ctx, key, val)
	if err != nil {
		switch err {
		case context.Canceled:
			fmt.Printf("ctx is canceled by another routine: %v\n", err)
		case context.DeadlineExceeded:
			fmt.Printf("ctx is attached with a deadline is exceeded: %v\n", err)
		case rpctypes.ErrEmptyKey:
			fmt.Printf("client-side error: %v\n", err)
		default:
			fmt.Printf("bad cluster endpoints, which are not etcd servers: %v\n", err)
		}
		return err
	}
	return nil
}

func ReadFromEtcd(cli *clientv3.Client, ctx context.Context, key string, prefix bool) (*clientv3.GetResponse, error) {
	var resp *clientv3.GetResponse
	var err error
	if prefix {
		resp, err = cli.Get(ctx, key, clientv3.WithPrefix())
	} else {
		resp, err = cli.Get(ctx, key)
	}
	if err != nil {
		switch err {
		case context.Canceled:
			fmt.Printf("ctx is canceled by another routine: %v\n", err)
		case context.DeadlineExceeded:
			fmt.Printf("ctx is attached with a deadline is exceeded: %v\n", err)
		case rpctypes.ErrEmptyKey:
			fmt.Printf("client-side error: %v\n", err)
		default:
			fmt.Printf("bad cluster endpoints, which are not etcd servers: %v\n", err)
		}
		return nil, err
	}
	for _, ev := range resp.Kvs {
		fmt.Printf("%s : %s\n", ev.Key, ev.Value)
	}
	return resp, nil
}

func WriteToEtcdFromPb(cli *clientv3.Client, ctx context.Context, key string, val proto.Message) error {
	marshalled, err := proto.Marshal(val)
	if err != nil {
		return fmt.Errorf("failed to marshal proto message: %w", err)
	}
	err = WriteToEtcd(cli, ctx, key, string(marshalled))
	return err
}

func ReadOneFromEtcdToPb(cli *clientv3.Client, ctx context.Context, key string, to proto.Message) error {
	read, err := ReadFromEtcd(cli, ctx, key, false)
	if err != nil {
		return err
	}
	if read.Count == 0 {
		return fmt.Errorf("key not found: %s", key)
	}

	kv := read.Kvs[0]
	if err := proto.Unmarshal(kv.Value, to); err != nil {
		return fmt.Errorf("failed to unmarshal proto: %w", err)
	}

	return nil
}

func ReadManyFromEtcdToPb(cli *clientv3.Client, ctx context.Context, keyPrefix string, to []proto.Message) error {
	read, err := ReadFromEtcd(cli, ctx, keyPrefix, true)
	if err != nil {
		return err
	}
	if read.Count == 0 {
		return fmt.Errorf("key not found: %s", keyPrefix)
	}

	to = to[:0] // clears the "to"

	for _, kv := range read.Kvs {
		message := reflect.New(reflect.TypeOf(to).Elem()).Interface().(proto.Message)
		if err := proto.Unmarshal(kv.Value, message); err != nil {
			return fmt.Errorf("failed to unmarshal proto: %w", err)
		}
		to = append(to, message)
	}
	return nil
}

func DeleteFromEtcdByPrefix(cli *clientv3.Client, ctx context.Context, prefix string) error {
	resp, err := cli.Delete(ctx, prefix, clientv3.WithPrefix())
	if err != nil {
		switch err {
		case context.Canceled:
			fmt.Printf("ctx is canceled by another routine: %v\n", err)
		case context.DeadlineExceeded:
			fmt.Printf("ctx is attached with a deadline is exceeded: %v\n", err)
		case rpctypes.ErrEmptyKey:
			fmt.Printf("client-side error: %v\n", err)
		default:
			fmt.Printf("bad cluster endpoints, which are not etcd servers: %v\n", err)
		}
		return err
	}

	fmt.Printf("deleted %d keys with prefix '%s'.\n", resp.Deleted, prefix)
	return nil
}
