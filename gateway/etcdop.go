package gateway

import (
	"context"
	"fmt"
	"log"
	"time"

	v3 "github.com/coreos/etcd/clientv3"
	"github.com/coreos/etcd/etcdserver/api/v3rpc/rpctypes"
)

func NewEtcdCluster() *EtcdCluster {
	return &EtcdCluster{}
}

type EtcdCluster struct {
	ClusterAddress    []string
	Proxy             string
	ConnectionTimeout int64
	RequestTimeout    int64
	cli               *v3.Client
	WatchRespChan     chan v3.WatchResponse
}

func (e *EtcdCluster) Init(address []string) error {
	var err error
	e.cli, err = v3.New(v3.Config{
		Endpoints:        address,
		AutoSyncInterval: time.Duration(600) * time.Second,
		DialTimeout:      time.Duration(5) * time.Second,
	})

	if err != nil {
		return err
	}
	return nil
}

func (e *EtcdCluster) WatchDir(dirKey string) error {
	if dirKey == "" {
		return fmt.Errorf("empty watch dir", dirKey)
	}
	defer func() {
		if re := recover(); re != nil {
			log.Println("recover panic : ", re)
		}
	}()
	ctx := context.Background()
	watcher := e.cli.Watch(ctx, dirKey)
	for change := range watcher {
		if len(change.Events) <= 0 {
			continue
		}
		switch change.Events[0].Type {
		case v3.EventTypeDelete:
		case v3.EventTypePut:
		}
	}
	return nil
}

func (e *EtcdCluster) Close() {
	if err := e.cli.Close(); err != nil {
		log.Println(err)
	}
	return
}

func (e *EtcdCluster) AddKeyValue(key, value string) error {
	ctx, cancelfunc := context.WithTimeout(context.Background(), time.Duration(e.RequestTimeout)*time.Second)
	resp, err := e.cli.Put(ctx, key, value)
	defer cancelfunc()
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

func (e *EtcdCluster) GetKeyValue(key string) (string, error) {
	ctx, cancelfunc := context.WithTimeout(context.Background(), time.Duration(e.RequestTimeout)*time.Second)
	resp, err := e.cli.Get(ctx, key)
	defer cancelfunc()
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
		return "", err
	}
	if len(resp.Kvs) <= 0 {
		return "", fmt.Errorf("key [%s] get empty value", key)
	} else {
		fmt.Printf("key [%s] get value list :[%s]", key, resp.Kvs)
	}
	value := string(resp.Kvs[0].Value)
	return value, nil
}

func (e *EtcdCluster) DeleteKeyValue(key string) error {
	ctx, cancelfunc := context.WithTimeout(context.Background(), time.Duration(e.RequestTimeout)*time.Second)
	resp, err := e.cli.Delete(ctx, key)
	defer cancelfunc()
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
