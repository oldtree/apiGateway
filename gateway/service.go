package gateway

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"net/http/httptrace"
	"net/http/pprof"
	"os"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/FlyCynomys/tools/log"

	"github.com/julienschmidt/httprouter"
	"github.com/oldtree/apiGateway/gateway/servicedesc"
	"github.com/oldtree/apiGateway/gateway/utils"
)

const (
	LoadblanceRoundRobinType = 1 << iota
	LoadblanceRandomtype
	LoadblanceWeighttype
)

var DefaultService = new(ApiService)
var starttime = time.Now()

func InitDefaultService() {
	DefaultService.client = nil
	DefaultService.BackendMap = nil
	DefaultService.ServiceName = "debug"
	DefaultService.Protocal = "http"
	DefaultService.LoadBalanceType = 0
	DefaultService.R = NewRoute(DefaultService)

	DefaultService.R.router.GET(fmt.Sprintf("/%s/info", DefaultService.ServiceName), func(w http.ResponseWriter, req *http.Request, _ httprouter.Params) {
		type Info struct {
			GOPATH     string    `json:"gopath,omitempty"`
			GOROOT     string    `json:"goroot,omitempty"`
			CPU        int       `json:"cpu,omitempty"`
			CpuProfile string    `json:"cpu_profile,omitempty"`
			Goroutine  int       `json:"goroutine,omitempty"`
			CgoCall    int64     `json:"cgo_call,omitempty"`
			PWD        string    `json:"pwd,omitempty"`
			StartTime  time.Time `json:"start_time,omitempty"`
		}
		var sysinfo Info
		sysinfo.GOPATH = os.Getenv("GOPATH")
		sysinfo.GOROOT = os.Getenv("GOROOT")
		sysinfo.CPU = runtime.NumCPU()
		sysinfo.Goroutine = runtime.NumGoroutine()
		sysinfo.CgoCall = runtime.NumCgoCall()
		sysinfo.PWD, _ = os.Getwd()
		sysinfo.StartTime = starttime
		re := new(utils.Result)
		re.Code = 1
		re.Description = "ok"
		re.Data = sysinfo
		w.Write(re.Json())
	})
	DefaultService.R.router.GET(fmt.Sprintf("/%s/app/handles", DefaultService.ServiceName), func(w http.ResponseWriter, req *http.Request, _ httprouter.Params) {
		re := new(utils.Result)
		re.Code = 1
		re.Description = "ok"
		re.Data = nil
		w.Write(re.Json())
	})
	DefaultService.R.router.GET(fmt.Sprintf("/favicon.ico", DefaultService.ServiceName), func(w http.ResponseWriter, req *http.Request, _ httprouter.Params) {
		data, err := ioutil.ReadFile("favicon.ico")
		re := new(utils.Result)
		if err != nil {
			re.Code = -1
			re.Description = "error"
			re.Data = err
			w.Write(re.Json())
		} else {
			re.Code = 1
			re.Description = "ok"
			re.Data = data
			w.Write(re.Json())
		}
		w.Write(re.Json())
	})
	DefaultService.R.router.GET(fmt.Sprintf("/%s/cmd", DefaultService.ServiceName), func(w http.ResponseWriter, req *http.Request, _ httprouter.Params) {
		pprof.Cmdline(w, req)
	})
	DefaultService.R.router.GET(fmt.Sprintf("/%s/profile", DefaultService.ServiceName), func(w http.ResponseWriter, req *http.Request, _ httprouter.Params) {
		pprof.Profile(w, req)
	})
	DefaultService.R.router.GET(fmt.Sprintf("/%s/trace", DefaultService.ServiceName), func(w http.ResponseWriter, req *http.Request, _ httprouter.Params) {
		pprof.Trace(w, req)
	})
	DefaultService.R.router.GET(fmt.Sprintf("/%s/symbol", DefaultService.ServiceName), func(w http.ResponseWriter, req *http.Request, _ httprouter.Params) {
		pprof.Symbol(w, req)
	})
	DefaultService.R.router.GET(fmt.Sprintf("/%s/goroutine", DefaultService.ServiceName), func(w http.ResponseWriter, req *http.Request, _ httprouter.Params) {
		pprof.Handler("goroutine").ServeHTTP(w, req)
	})
	DefaultService.R.router.GET(fmt.Sprintf("/%s/heap", DefaultService.ServiceName), func(w http.ResponseWriter, req *http.Request, _ httprouter.Params) {
		pprof.Handler("heap").ServeHTTP(w, req)
	})
	DefaultService.R.router.GET(fmt.Sprintf("/%s/block", DefaultService.ServiceName), func(w http.ResponseWriter, req *http.Request, _ httprouter.Params) {
		pprof.Handler("block").ServeHTTP(w, req)
	})

}

func init() {
	InitDefaultService()
}

type ApiService struct {
	BackendMap map[int]*Node `json:"backend_map,omitempty"`

	ServiceName       string `json:"service_name,omitempty"`
	Version           string `json:"version,omitempty"`
	Protocal          string `json:"protocal,omitempty"`
	LoadBalanceType   uint   `json:"load_balance_type,omitempty"`
	ReadWriteTimeout  int64  `json:"read_write_timeout,omitempty"`
	ConnectionTimeout int64  `json:"connection_timeout,omitempty"`

	R      *Route       `json:"route,omitempty"`
	client *http.Client `json:"client,omitempty"`

	XEtag map[string]string `json:"x_etag,omitempty"`

	sync.Mutex `json:"mutex"`
	OnlineTime time.Time `json:"online_time,omitempty"`
}

func NewApiService() *ApiService {
	apiSrv := &ApiService{
		BackendMap: make(map[int]*Node, 16),
	}
	apiSrv.R = NewRoute(apiSrv)
	return apiSrv
}

func (srv *ApiService) MappingApiServiceFromData(data []byte) error {
	if len(data) <= 0 {
		return nil
	}
	si := new(servicedesc.ServiceDesc)
	err := json.Unmarshal(data, si)
	if err != nil {
		return err
	}
	err = srv.MappingApiService(si)
	return nil
}

func (srv *ApiService) MappingApiService(si *servicedesc.ServiceDesc) error {
	if si == nil {
		return ErrMappingServiceInfoFailed
	}
	////////start mapping service info//////////
	srv.ServiceName = si.ServiceName
	srv.Version = si.Version
	srv.Protocal = si.Protocal
	srv.LoadBalanceType = si.LoadBalanceType
	srv.ReadWriteTimeout = int64(si.ReadWriteTimeout)
	srv.ConnectionTimeout = int64(si.ConnectionTimeout)
	if si.Createtime == "" {
		srv.OnlineTime = time.Now()
	} else {
		srv.OnlineTime, _ = time.Parse("2006-01-02 15:04:05", si.Createtime)
	}

	if si.Api != nil {
		if si.Api.Info != nil {
			for _, Pathvalue := range si.Api.Info {
				for _, methodValue := range Pathvalue.HandleMethed {
					temp := &RoutePathInfo{
						InnerPath:   Pathvalue.InnerRoutePath,
						OuterPath:   Pathvalue.OuterRoutePath,
						Description: Pathvalue.OuterRoutePath,
						InnerMethod: methodValue.InnerMethod,
						OuterMethod: methodValue.OuterMethod,
						Auth:        Pathvalue.NeedAuth,
						Status:      Pathvalue.Status,
					}
					srv.R.RouterMappingInfo = append(srv.R.RouterMappingInfo, temp)

				}
			}
		}
	}
	if len(si.XEtag) >= 0 {
		srv.XEtag = make(map[string]string, 16)
		for key, value := range si.XEtag {
			srv.XEtag[key] = value
		}
	} else {
		srv.XEtag = nil
	}
	return nil
}

func (srv *ApiService) InitServiceHttpClient() error {
	srv.client = &http.Client{
		Timeout: time.Duration(0),
		Transport: &http.Transport{
			DialContext: (&net.Dialer{
				Timeout:   30 * time.Second,
				KeepAlive: 30 * time.Second,
				DualStack: true,
			}).DialContext,
			DisableCompression:    false,
			DisableKeepAlives:     false,
			ResponseHeaderTimeout: 120 * time.Second,
			MaxIdleConns:          100,
			MaxIdleConnsPerHost:   10,
			IdleConnTimeout:       300 * time.Second,
		},
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return fmt.Errorf("redirect failed")
		},
	}
	return nil
}

func (srv *ApiService) MiddleWareWrap(method string, innerpath string, middleware interface{}) httprouter.Handle {

	switch method {
	case "GET":
		return srv.HandleWrapGetMethod(innerpath)
	case "POST":
		return srv.HandleWrapPostMethod(innerpath)
	case "PUT":
		return srv.HandleWrapPutMethod(innerpath)
	case "DELETE":
		return srv.HandleWrapDeleteMethod(innerpath)
	default:
		return func(w http.ResponseWriter, req *http.Request, _ httprouter.Params) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(404)
			w.Write([]byte("request method not support"))
			return
		}
	}
	return func(w http.ResponseWriter, req *http.Request, _ httprouter.Params) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(404)
		w.Write([]byte("request method not support"))
		return
	}
}

func (srv *ApiService) HandleWrapGetMethod(innerpath string) httprouter.Handle {
	return func(w http.ResponseWriter, req *http.Request, _ httprouter.Params) {
		var (
			status        int8
			node          *Node
			req2Endpoinrt *http.Request
			err           error
			retry         uint
		)
		retry = 0
		for status == 0 {
			node = srv.SelectNode()
			if node == nil {
				w.Header().Set("Connection", "keep-alive")
				w.WriteHeader(500)
				w.Write([]byte("Internel server error"))
				return
			}
			urlStr := fmt.Sprintf("%s://%s/%s?%s", srv.Protocal, node.Address, innerpath, req.URL.RawQuery)
			log.Info(urlStr)
			req2Endpoinrt, err = copyRequest(req, urlStr, "GET")
			if err != nil {
				log.Debug(err)
				w.Write([]byte("forward request failed : " + err.Error()))
				return
			}
			response, err := srv.client.Do(req2Endpoinrt)
			if err != nil {
				log.Debug(err)
				status = -1
				w.Write([]byte("forward request failed : " + err.Error()))
				return
			}
			defer response.Body.Close()
			switch response.StatusCode / 100 {
			case 2:
				status = 1
				node.Status = NodeStatusOK
				for k, v := range response.Header {
					for _, vv := range v {
						w.Header().Set(k, vv)
					}
				}
				w.Header().Set("Connection", "keep-alive")
				w.WriteHeader(response.StatusCode)
				io.Copy(w, response.Body)
				return
			case 3:
				status = 1
				node.Status = NodeStatusOK
				if urlStrRedirect := response.Header.Get("Location"); urlStrRedirect != "" {
					http.Redirect(w, req, urlStrRedirect, http.StatusFound)
					return
				}
				w.Header().Set("Connection", "keep-alive")
				w.WriteHeader(response.StatusCode)
				w.Write([]byte("rediret to url"))
				return
			case 4:
				status = 1
				node.Status = NodeStatusOK
				for k, v := range response.Header {
					for _, vv := range v {
						w.Header().Set(k, vv)
					}
				}
				w.Header().Set("Connection", "keep-alive")
				w.WriteHeader(response.StatusCode)
				io.Copy(w, response.Body)
				return
			case 5:
				if retry > 3 {
					status = 0
					node.Status = NodeStatusError
					w.Header().Set("Connection", "keep-alive")
					w.WriteHeader(response.StatusCode)
					w.Write([]byte("Internel server error"))
					return
				} else {
					status = 0
					node.Status = NodeStatusError
					retry = retry + 1
					continue
				}
			default:
			}
			status = 1
		}
		return
	}
}

func (srv *ApiService) HandleWrapPostMethod(innerpath string) httprouter.Handle {
	return func(w http.ResponseWriter, req *http.Request, _ httprouter.Params) {
		var (
			status        int8
			node          *Node
			req2Endpoinrt *http.Request
			err           error
			retry         uint
		)
		retry = 0
		for status == 0 {
			node = srv.SelectNode()
			if node == nil {
				return
			}
			urlStr := fmt.Sprintf("%s://%s/%s?%s", srv.Protocal, node.Address, innerpath, req.URL.RawQuery)
			log.Info(urlStr)
			req2Endpoinrt, err = copyRequest(req, urlStr, "POST")
			if err != nil {
				log.Debug(err)
				w.Write([]byte("forward request failed : " + err.Error()))
				return
			}
			response, err := srv.client.Do(req2Endpoinrt)
			if err != nil {
				log.Debug(err)
				status = -1
				w.Write([]byte("forward request failed : " + err.Error()))
				return
			}
			defer response.Body.Close()
			switch response.StatusCode / 100 {
			case 2:
				status = 1
				node.Status = NodeStatusOK
				for k, v := range response.Header {
					for _, vv := range v {
						w.Header().Set(k, vv)
					}
				}
				w.Header().Set("Connection", "keep-alive")
				w.WriteHeader(response.StatusCode)
				io.Copy(w, response.Body)
				return
			case 3:
				status = 1
				node.Status = NodeStatusOK
				if urlStrRedirect := response.Header.Get("Location"); urlStrRedirect != "" {
					http.Redirect(w, req, urlStrRedirect, http.StatusFound)
					return
				}
				w.Header().Set("Connection", "keep-alive")
				w.WriteHeader(response.StatusCode)
				w.Write([]byte("rediret to url"))
				return
			case 4:
				status = 1
				node.Status = NodeStatusOK
				for k, v := range response.Header {
					for _, vv := range v {
						w.Header().Set(k, vv)
					}
				}
				w.Header().Set("Connection", "keep-alive")
				w.WriteHeader(response.StatusCode)
				io.Copy(w, response.Body)
				return
			case 5:
				if retry > 3 {
					status = 0
					node.Status = NodeStatusError
					w.Header().Set("Connection", "keep-alive")
					w.WriteHeader(response.StatusCode)
					w.Write([]byte("Internel server error"))
					return
				} else {
					status = 0
					node.Status = NodeStatusError
					retry = retry + 1
					continue
				}
			default:
			}
			status = 1
		}
		return
	}
}

func (srv *ApiService) HandleWrapPutMethod(innerpath string) httprouter.Handle {
	return func(w http.ResponseWriter, req *http.Request, _ httprouter.Params) {
		var (
			status        int8
			node          *Node
			req2Endpoinrt *http.Request
			err           error
			retry         uint
		)
		retry = 0
		for status == 0 {
			node = srv.SelectNode()
			if node == nil {
				w.Header().Set("Connection", "keep-alive")
				w.WriteHeader(500)
				w.Write([]byte("Internel server error"))
				return
			}
			urlStr := fmt.Sprintf("%s://%s/%s?%s", srv.Protocal, node.Address, innerpath, req.URL.RawQuery)
			log.Info(urlStr)
			req2Endpoinrt, err = copyRequest(req, urlStr, "PUT")
			if err != nil {
				log.Debug(err)
				w.Write([]byte("forward request failed : " + err.Error()))
				return
			}
			response, err := srv.client.Do(req2Endpoinrt)
			if err != nil {
				log.Debug(err)
				status = -1
				w.Write([]byte("forward request failed : " + err.Error()))
				return
			}
			defer response.Body.Close()
			switch response.StatusCode / 100 {
			case 2:
				status = 1
				node.Status = NodeStatusOK
				for k, v := range response.Header {
					for _, vv := range v {
						w.Header().Set(k, vv)
					}
				}
				w.Header().Set("Connection", "keep-alive")
				w.WriteHeader(response.StatusCode)
				io.Copy(w, response.Body)
				return
			case 3:
				status = 1
				node.Status = NodeStatusOK
				if urlStrRedirect := response.Header.Get("Location"); urlStrRedirect != "" {
					http.Redirect(w, req, urlStrRedirect, http.StatusFound)
					return
				}
				w.Header().Set("Connection", "keep-alive")
				w.WriteHeader(response.StatusCode)
				w.Write([]byte("rediret to url"))
				return
			case 4:
				status = 1
				node.Status = NodeStatusOK
				for k, v := range response.Header {
					for _, vv := range v {
						w.Header().Set(k, vv)
					}
				}
				w.Header().Set("Connection", "keep-alive")
				w.WriteHeader(response.StatusCode)
				io.Copy(w, response.Body)
				return
			case 5:
				if retry > 3 {
					status = 0
					node.Status = NodeStatusError
					w.Header().Set("Connection", "keep-alive")
					w.WriteHeader(response.StatusCode)
					w.Write([]byte("Internel server error"))
					return
				} else {
					status = 0
					node.Status = NodeStatusError
					retry = retry + 1
					continue
				}
			default:
			}
			status = 1
		}
		return
	}
}

func (srv *ApiService) HandleWrapDeleteMethod(innerpath string) httprouter.Handle {
	return func(w http.ResponseWriter, req *http.Request, _ httprouter.Params) {
		var (
			status        int8
			node          *Node
			req2Endpoinrt *http.Request
			err           error
			retry         uint
		)
		retry = 0
		for status == 0 {
			node = srv.SelectNode()
			if node == nil {
				w.Header().Set("Connection", "keep-alive")
				w.WriteHeader(500)
				w.Write([]byte("Internel server error"))
				return
			}
			urlStr := fmt.Sprintf("%s://%s/%s?%s", srv.Protocal, node.Address, innerpath, req.URL.RawQuery)
			log.Info(urlStr)
			req2Endpoinrt, err = copyRequest(req, urlStr, "DELETE")
			if err != nil {
				log.Debug(err)
				w.Write([]byte("forward request failed : " + err.Error()))
				return
			}
			response, err := srv.client.Do(req2Endpoinrt)
			if err != nil {
				log.Debug(err)
				status = -1
				w.Write([]byte("forward request failed : " + err.Error()))
				return
			}
			defer response.Body.Close()
			switch response.StatusCode / 100 {
			case 2:
				status = 1
				node.Status = NodeStatusOK
				for k, v := range response.Header {
					for _, vv := range v {
						w.Header().Set(k, vv)
					}
				}
				w.Header().Set("Connection", "keep-alive")
				w.WriteHeader(response.StatusCode)
				io.Copy(w, response.Body)
				return
			case 3:
				status = 1
				node.Status = NodeStatusOK
				if urlStrRedirect := response.Header.Get("Location"); urlStrRedirect != "" {
					http.Redirect(w, req, urlStrRedirect, http.StatusFound)
					return
				}
				w.Header().Set("Connection", "keep-alive")
				w.WriteHeader(response.StatusCode)
				w.Write([]byte("rediret to url"))
				return
			case 4:
				status = 1
				node.Status = NodeStatusOK
				for k, v := range response.Header {
					for _, vv := range v {
						w.Header().Set(k, vv)
					}
				}
				w.Header().Set("Connection", "keep-alive")
				w.WriteHeader(response.StatusCode)
				io.Copy(w, response.Body)
				return
			case 5:
				if retry > 3 {
					status = 0
					node.Status = NodeStatusError
					w.Header().Set("Connection", "keep-alive")
					w.WriteHeader(response.StatusCode)
					w.Write([]byte("Internel server error"))
					return
				} else {
					status = 0
					node.Status = NodeStatusError
					retry = retry + 1
					continue
				}
			default:
			}
			status = 1
		}
		return
	}
}

func (srv *ApiService) SelectNode() *Node {
	if len(srv.BackendMap) == 0 {
		return nil
	}
	var n *Node
	if len(srv.BackendMap) == 1 {
		for _, value := range srv.BackendMap {
			n = value
		}
		if n.Status == NodeStatusError {
			return nil
		}
		return n
	}
	switch srv.LoadBalanceType {
	case LoadblanceRoundRobinType:
		return nil
	case LoadblanceRandomtype:
		return nil
	case LoadblanceWeighttype:
		return nil
	}
	return n
}

func (srv *ApiService) AddNode(n *Node) error {
	if n == nil {
		return fmt.Errorf("node info is nil")
	}
	defer func() {
		if re := recover(); re != nil {
			log.Info(re)
		}
	}()
	srv.Lock()
	defer srv.Unlock()
	if _, ok := srv.BackendMap[n.NodeID]; ok {
		return fmt.Errorf("node [%d] is exist", n.NodeID)
	} else {
		srv.BackendMap[n.NodeID] = n
	}
	return nil
}

func (srv *ApiService) UpdateNode(n *Node) error {
	if n == nil {
		return fmt.Errorf("node info is nil")
	}
	defer func() {
		if re := recover(); re != nil {
			log.Info(re)
		}
	}()
	srv.Lock()
	defer srv.Unlock()
	if _, ok := srv.BackendMap[n.NodeID]; ok {
		return fmt.Errorf("node [%d] is exist", n.NodeID)
	} else {
		srv.BackendMap[n.NodeID] = n
	}
	return nil
}

func (srv *ApiService) RemoveNode(n *Node) error {
	if n == nil {
		return fmt.Errorf("node info is nil")
	}
	defer func() {
		if re := recover(); re != nil {
			log.Info(re)
		}
	}()
	srv.Lock()
	defer srv.Unlock()
	delete(srv.BackendMap, n.NodeID)
	return nil
}

func copyResuqetWithTrace(srcR *http.Request, urlStr string, method string) (*http.Request, error) {
	ctx, cancelfunc := context.WithDeadline(context.Background(), time.Now().Add(time.Second*30))
	ctx = httptrace.WithClientTrace(ctx, nil)
	defer cancelfunc()

	return nil, nil
}

func copyRequest(srcR *http.Request, urlStr string, method string) (*http.Request, error) {
	if srcR == nil {
		return nil, fmt.Errorf("src request is nil")
	}
	var nReq *http.Request
	var err error
	switch method {
	case "GET", "get":
		nReq, err = http.NewRequest(method, urlStr, nil)
		if err != nil {
			return nil, err
		}
	case "PUT", "put", "POST", "post":

		if contentType := srcR.Header.Get("Content-Type"); strings.Contains(contentType, "x-www-form-urlencoded") {
			srcR.ParseForm()
			nReq, err = http.NewRequest(method, urlStr, strings.NewReader(srcR.PostForm.Encode()))
			if err != nil {
				return nil, err
			}
		} else {
			nReq, err = http.NewRequest(method, urlStr, srcR.Body)
			if err != nil {
				return nil, err
			}
		}
	case "delete", "DELETE":
		nReq, err = http.NewRequest(method, urlStr, nil)
		if err != nil {
			return nil, err
		}
	default:
		return nil, fmt.Errorf("unsupport method")
	}

	nReq.Header = make(http.Header, len(srcR.Header))
	for key, value := range srcR.Header {
		va := make([]string, len(value))
		copy(va, value)
		nReq.Header[key] = va
	}
	return nReq, nil
}
