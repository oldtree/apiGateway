package gateway

import (
	"log"
	"time"

	v3 "github.com/coreos/etcd/clientv3"
)

func NewEtcdCluster() *EtcdCluster {
	return &EtcdCluster{}
}

type EtcdCluster struct {
	ClusterAddress    []string
	Proxy             string
	ConnectionTimeout int64
	cli               *v3.Client
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

func (e *EtcdCluster) Close() {
	if err := e.cli.Close(); err != nil {
		log.Println(err)
	}
	return
}

func (e *EtcdCluster) AddKeyValue(key, value []byte) error {

	return nil
}
