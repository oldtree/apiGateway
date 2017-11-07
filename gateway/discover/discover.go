package discover

const (
	DiscoverTypeEtcd       = "etcd"
	DiscoverTypeK8S        = "k8s"
	DiscoverTypeConsul     = "consul"
	DiscoverTypeRedis      = "Redis"
	DiscoverTypeBoltDb     = "boltdb"
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

type Adpater struct {
	DiscoverType string
	Dis          Discover
}
