package discover

import (
	"fmt"
)

const (
	DiscoverTypeEtcd       = "etcd"
	DiscoverTypeK8S        = "k8s"
	DiscoverTypeConsul     = "consul"
	DiscoverTypeRedis      = "Redis"
	DiscoverTypeBoltDB     = "boltdb"
	DiscoverTypePostgreSQL = "postgresql"
)

type Docter interface {
	Health() bool
	DoctorsAdvice() string
	Surgery() bool
}

type Discover interface {
	Sub(interface{}) interface{}
	Pub(interface{}) bool
	BroadCast(interface{}) bool
}

type BuildDiscover func() (Discover, Docter)

var etcdfunc BuildDiscover
var k8sfunc BuildDiscover
var redisfunc BuildDiscover
var consulfunc BuildDiscover
var boltdbfunc BuildDiscover
var postresqlfunc BuildDiscover

type DiscoverAdpater struct {
	DiscoverType string
	Dis          Discover
	Docter       Docter
	Data         chan []byte
}

func (d *DiscoverAdpater) BuildAdapter(fn BuildDiscover) error {

	return nil
}

func (d *DiscoverAdpater) Watcher() {

}

func SetUpWatcher(discoverType string, dis *DiscoverAdpater) error {
	switch discoverType {
	case DiscoverTypeEtcd:
		dis.BuildAdapter(etcdfunc)
	case DiscoverTypeBoltDB:
		dis.BuildAdapter(boltdbfunc)
	case DiscoverTypeConsul:
		dis.BuildAdapter(consulfunc)
	case DiscoverTypeK8S:
		dis.BuildAdapter(k8sfunc)
	case DiscoverTypePostgreSQL:
		dis.BuildAdapter(postresqlfunc)
	case DiscoverTypeRedis:
		dis.BuildAdapter(redisfunc)
	default:
		return fmt.Errorf("error discover type : %s \n", discoverType)
	}
	return nil
}
