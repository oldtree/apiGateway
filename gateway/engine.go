package gateway

import (
	"net/http"
	"strings"
	"sync"
)

type Engine struct {
	SrvMap map[string]http.Handler
	sync.RWMutex
}

var defaultEngine *Engine
var singlef *sync.Once

func DefaultEngine() *Engine {
	singlef.Do(func() {
		defaultEngine = NewEngine()
	})
	return defaultEngine
}

func NewEngine() *Engine {
	return &Engine{
		SrvMap: make(map[string]http.Handler, 16),
	}
}

func (e *Engine) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	servicePath := strings.TrimLeft(r.URL.Path, "/")
	servicePath = strings.Split(servicePath, "/")[0]
	println(servicePath)
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
	e.SrvMap[srv.ServiceName] = srv.R.router
}

func (e *Engine) DelService(srv *ApiService) {
	if srv == nil {
		return
	}
	e.Lock()
	defer e.Unlock()
	delete(e.SrvMap, srv.ServiceName)
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
