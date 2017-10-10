package etcdop

import (
	"context"
	"fmt"
	"time"

	"github.com/FlyCynomys/tools/log"

	v3 "github.com/coreos/etcd/clientv3"
	"github.com/coreos/etcd/etcdserver/api/v3rpc/rpctypes"
)

/*
/v2/keys/gateway  服务信息的根目录
/v2/keys/gateway/service 服务的目录
/v2/keys/gateway/service/api 服务的api映射信息目录
/v2/keys/gateway/service/node 服务的后端可工作节点信息
/v2/keys/gateway/service/config  服务的配置信息
/v2/keys/gateway/service/errors  服务的返回错误信息配置
*/

type EventCallback func(ee *EtcdEvent) error

func NewEtcdCluster() *EtcdCluster {
	return &EtcdCluster{}
}

type EtcdEvent struct {
	Key   []byte `json:"key,omitempty"`
	Value []byte `json:"value,omitempty"`
	Index uint64 `json:"index,omitempty"`
	Type  string `json:"type,omitempty"`
}

type EtcdCluster struct {
	ClusterAddress    []string
	Proxy             string
	ConnectionTimeout int64
	RequestTimeout    int64
	cli               *v3.Client
	WatchResp         v3.WatchChan

	EtcdEventCallback EventCallback
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
			log.Info(re)
		}
	}()
	ctx := context.Background()
	e.WatchResp = e.cli.Watch(ctx, dirKey)
	for change := range e.WatchResp {
		if len(change.Events) <= 0 {
			continue
		}
		eec := new(EtcdEvent)
		switch change.Events[0].Type {
		case v3.EventTypeDelete:
			eec.Key = append(eec.Key, change.Events[0].Kv.Key...)
			eec.Value = append(eec.Value, change.Events[0].Kv.Value...)
			eec.Index = change.Header.GetRaftTerm()
			eec.Type = "DELETE"
			e.EtcdEventCallback(eec)
		case v3.EventTypePut:
			eec.Key = append(eec.Key, change.Events[0].Kv.Key...)
			eec.Value = append(eec.Value, change.Events[0].Kv.Value...)
			eec.Index = change.Header.GetRaftTerm()
			eec.Type = "PUT"
			e.EtcdEventCallback(eec)
		}
	}
	return nil
}

func (e *EtcdCluster) Close() {
	if err := e.cli.Close(); err != nil {
		log.Error(err)
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
	log.Info(resp)
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
	log.Info(resp)
	return nil
}
