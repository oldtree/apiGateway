package gateway

import (
	"errors"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/FlyCynomys/tools/log"
	"github.com/oldtree/apiGateway/gateway/etcdop"
	"github.com/oldtree/apiGateway/gateway/utils"
)

type Engine struct {
	SrvMap map[string]http.Handler
	sync.RWMutex
	Notice        chan *utils.Event
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
	temp := Engine{
		SrvMap:        make(map[string]http.Handler, 16),
		Notice:        make(chan *utils.Event, 16),
		EtcdOperation: etcdop.NewEtcdCluster(),
	}
	err := temp.EtcdOperation.Init([]string{"http://localhost:2379"})
	if err != nil {
		return nil
	}
	temp.EtcdOperation.EtcdEventCallback = temp.EventCallback
	return &temp
}

func (e *Engine) EventCallback(ee *etcdop.EtcdEvent) error {
	newevt := new(utils.Event)
	var err error
	path := strings.Split(string(ee.Key), "/")
	if path[0] == "nodes" {
		switch ee.Type {
		case "DELETE":
			newevt.EventType = utils.EventNodeDelete
			newevt.TimeStamp = time.Now().String()
			srv := NewApiService()
			err = srv.MappingApiServiceFromData(ee.Value)
			if err != nil {
				return err
			}
			newevt.Content = srv
		case "PUT":
			newevt.EventType = utils.EventNodeAdd
			newevt.TimeStamp = time.Now().String()
			srv := NewApiService()
			err = srv.MappingApiServiceFromData(ee.Value)
			if err != nil {
				return err
			}
			newevt.Content = srv
		default:
			log.Error("error node event ", ee)
			return errors.New("error node event ")
		}
	} else if path[0] == "service" {
		switch ee.Type {
		case "DELETE":
			newevt.EventType = utils.EventServiceDelete
			newevt.TimeStamp = time.Now().String()
			node := NewNodeFromData(ee.Value)
			newevt.Content = node
		case "PUT":
			newevt.EventType = utils.EventServiceAdd
			newevt.TimeStamp = time.Now().String()
			node := NewNodeFromData(ee.Value)
			newevt.Content = node
		default:
			log.Error("error service event ", ee)
			return errors.New("error service event")
		}
	}
	e.Notice <- newevt
	return nil
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
				case utils.EventServiceAdd:
					e.AddService((*ApiService)(content))
				case utils.EventServiceGet:
				case utils.EventServiceUpdate:
					e.UpdateService((*ApiService)(content))
				case utils.EventServiceDelete:
					e.DelService((*ApiService)(content))
				default:
					//log.Info(fmt.Printf("not support event [%s] \n", evt))
				}
			case *Node:
				switch evt.EventType {
				case utils.EventNodeGet:
				case utils.EventNodeAdd:
					e.AddServiceBackendNode((*Node)(content))
				case utils.EventNodeUpdate:
					e.AddServiceBackendNode((*Node)(content))
				case utils.EventNodeDelete:
					e.DeleteServiceBackendNode((*Node)(content))
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
