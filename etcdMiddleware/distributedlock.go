package etcdMiddleware

import (
	"context"
	"fmt"
	clientv3 "go.etcd.io/etcd/client/v3"
)

func AcquireLock(cli *clientv3.Client, lockKey string, ttl int64, waitTillAvailable bool, ctx context.Context) (*clientv3.LeaseGrantResponse, error) {
	leaseResp, err := cli.Grant(ctx, ttl)
	if err != nil {
		return nil, fmt.Errorf("failed to create lease: %v", err)
	}

	txn := cli.Txn(ctx)
	txn.If(clientv3.Compare(clientv3.CreateRevision(lockKey), "=", 0)).
		Then(clientv3.OpPut(lockKey, "locked", clientv3.WithLease(leaseResp.ID)))

	txnResp, err := txn.Commit()
	if err != nil {
		return nil, fmt.Errorf("failed to commit txn: %v", err)
	}

	if !txnResp.Succeeded {
		if waitTillAvailable {
			err := WatchLock(cli, lockKey, ctx)
			if err != nil {
				return nil, fmt.Errorf("err when watching for lock: %v", err)
			}
			return AcquireLock(cli, lockKey, ttl, false, ctx)
		} else {
			return nil, fmt.Errorf("lock exists from before")
		}
	}

	return leaseResp, nil
}

func ReleaseLock(client *clientv3.Client, leaseID clientv3.LeaseID, ctx context.Context) error {
	_, err := client.Revoke(ctx, leaseID)
	if err != nil {
		return fmt.Errorf("failed to revoke lease: %v", err)
	}
	return nil
}

func WatchLock(client *clientv3.Client, lockKey string, ctx context.Context) error {
	rch := client.Watch(ctx, lockKey)

	for wresp := range rch {
		if wresp.Canceled {
			fmt.Printf("watch canceled: %v\n", wresp.Err())
			return wresp.Err()
		}
		for _, ev := range wresp.Events {
			if ev.Type == clientv3.EventTypeDelete {
				return nil
			}
		}
	}
	return nil
}
