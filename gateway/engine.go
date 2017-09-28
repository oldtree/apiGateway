package gateway

import (
	"net/http"
	"strings"
	"sync"

	"github.com/FlyCynomys/tools/log"
	"github.com/oldtree/apiGateway/gateway/etcdop"
)

type Engine struct {
	SrvMap map[string]http.Handler
	sync.RWMutex
	Notice        chan *Event
	EtcdOperation *etcdop.EtcdCluster
}

var defaultEngine *Engine
var singleinitf *sync.Once

func DefaultEngine() *Engine {
	singleinitf.Do(func() {
		defaultEngine = NewEngine()
	})
	return defaultEngine
}

func NewEngine() *Engine {
	return &Engine{
		SrvMap: make(map[string]http.Handler, 16),
		Notice: make(chan *Event, 16),
	}
}

func (e *Engine) Doorman() {
	defer func() {
		if re := recover(); re != nil {
			log.Info(re)
		}
	}()
	for {
		select {
		case evt := <-e.Notice:
			//log.Info(fmt.Printf("event [%d] happend [%s] \n", evt.EventType, evt.Content))
			switch content := evt.Content.(type) {
			case *ApiService:
				switch evt.EventType {
				case EventServiceAdd:
					e.AddService((*ApiService)(content))
				case EventServiceGet:
				case EventServiceUpdate:
					e.UpdateService((*ApiService)(content))
				case EventServiceDelete:
					e.DelService((*ApiService)(content))
				default:
					//log.Info(fmt.Printf("not support event [%s] \n", evt))
				}
			case *Node:
				switch evt.EventType {
				case EventServiceNodeGet:
				case EventServiceNodeAdd:
					e.AddServiceBackendNode((*Node)(content))
				case EventServiceNodeUpdate:
					e.AddServiceBackendNode((*Node)(content))
				case EventServiceNodeDelete:
					e.AddServiceBackendNode((*Node)(content))
				default:
					//log.Info(fmt.Printf("not support event [%s] \n", evt))
				}
			default:
				//log.Info(fmt.Printf("not support event [%s] \n", evt))
			}
		}
	}
}

func (e *Engine) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	servicePath := strings.TrimLeft(r.URL.Path, "/")
	servicePath = strings.Split(servicePath, "/")[0]
	if servicePath == "" {
		http.Error(w, "Forbidden", 403)
		return
	}
	e.RLock()
	handler := e.SrvMap[servicePath]
	e.RUnlock()
	if handler != nil {
		handler.ServeHTTP(w, r)
	} else {
		http.Error(w, "Forbidden", 403)
	}
}

func (e *Engine) AddService(srv *ApiService) {
	if srv == nil {
		return
	}
	e.Lock()
	defer e.Unlock()
	e.SrvMap[srv.ServiceName] = srv.R
}

func (e *Engine) UpdateService(srv *ApiService) {
	if srv == nil {
		return
	}
	e.Lock()
	defer e.Unlock()
	e.SrvMap[srv.ServiceName] = srv.R
}

func (e *Engine) DelService(srv *ApiService) {
	if srv == nil {
		return
	}
	e.Lock()
	defer e.Unlock()
	delete(e.SrvMap, srv.ServiceName)
}

func (e *Engine) AddServiceBackendNode(nd *Node) error {
	if nd == nil {
		return nil
	}
	e.RLock()
	route := e.SrvMap[nd.ServeiceName]
	e.RUnlock()
	err := route.(*Route).srvBelong.AddNode(nd)
	if err != nil {
		return err
	}
	return nil
}

func (e *Engine) DeleteServiceBackendNode(nd *Node) error {
	if nd == nil {
		return nil
	}
	e.RLock()
	route := e.SrvMap[nd.ServeiceName]
	e.RUnlock()

	err := route.(*Route).srvBelong.RemoveNode(nd)
	if err != nil {
		return err
	}
	return nil
}

func (e *Engine) UpdateServiceBackendNode(nd *Node) error {
	if nd == nil {
		return nil
	}
	e.RLock()
	route := e.SrvMap[nd.ServeiceName]
	e.RUnlock()

	err := route.(*Route).srvBelong.UpdateNode(nd)
	if err != nil {
		return err
	}
	return nil
}

func (e *Engine) GetAppserviceList() []string {
	e.RLock()
	defer e.RUnlock()
	var applist []string
	for key, _ := range e.SrvMap {
		applist = append(applist, key)
	}
	return applist
}
