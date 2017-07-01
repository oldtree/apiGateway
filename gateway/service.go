package gateway

import (
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"runtime"
	"sync"
	"time"

	"strings"

	"io"

	"github.com/julienschmidt/httprouter"
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
	DefaultService.LoadBlanceType = 0
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
		sysinfo.CpuProfile = string(runtime.CPUProfile())
		sysinfo.StartTime = starttime
		re := new(Result)
		re.Code = 1
		re.Description = "ok"
		re.Data = sysinfo
		w.Write(re.Json())
	})
	DefaultService.R.router.GET(fmt.Sprintf("/%s/app/handles", DefaultService.ServiceName), func(w http.ResponseWriter, req *http.Request, _ httprouter.Params) {
		var app []string
		re := new(Result)
		re.Code = 1
		re.Description = "ok"
		re.Data = app
		w.Write(re.Json())
	})

}

func init() {
	InitDefaultService()
}

type ApiService struct {
	BackendMap map[int]*Node `json:"backend_map,omitempty"`

	ServiceName       string `json:"service_name,omitempty"`
	Protocal          string `json:"protocal,omitempty"`
	LoadBlanceType    uint   `json:"load_blance_type,omitempty"`
	ReadWriteTimeout  int64  `json:"read_write_timeout,omitempty"`
	ConnectionTimeout int64  `json:"connection_timeout,omitempty"`

	R      *Route       `json:"r,omitempty"`
	client *http.Client `json:"client,omitempty"`

	sync.Mutex `json:"mutex"`
}

func NewApiService() *ApiService {
	apiSrv := &ApiService{
		BackendMap: make(map[int]*Node, 16),
	}
	return apiSrv
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
			ResponseHeaderTimeout: time.Duration(srv.ReadWriteTimeout) * time.Second,
			MaxIdleConns:          100,
			MaxIdleConnsPerHost:   10,
			IdleConnTimeout:       120 * time.Second,
		},
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return fmt.Errorf("redirect failed")
		},
	}
	return nil
}

func copyRequest(srcR *http.Request, urlStr string, method string) (*http.Request, error) {
	if srcR == nil {
		return nil, fmt.Errorf("src request if nil")
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

func (srv *ApiService) HandleGetMethod(w http.ResponseWriter, req *http.Request, _ httprouter.Params) {
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
		urlStr := fmt.Sprintf("%s://%s", srv.Protocal, node.Address)
		req2Endpoinrt, err = copyRequest(req, urlStr, "GET")
		if err != nil {
			log.Println(err)
			w.Write([]byte("forward request failed : " + err.Error()))
			return
		}
		response, err := srv.client.Do(req2Endpoinrt)
		if err != nil {
			log.Println(err)
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

func (srv *ApiService) HandlePostMethod(w http.ResponseWriter, req *http.Request, _ httprouter.Params) {
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
		urlStr := fmt.Sprintf("%s://%s", srv.Protocal, node.Address)
		req2Endpoinrt, err = copyRequest(req, urlStr, "POST")
		if err != nil {
			log.Println(err)
			w.Write([]byte("forward request failed : " + err.Error()))
			return
		}
		response, err := srv.client.Do(req2Endpoinrt)
		if err != nil {
			log.Println(err)
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

func (srv *ApiService) HandlePutMethod(w http.ResponseWriter, req *http.Request, _ httprouter.Params) {
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
		urlStr := fmt.Sprintf("%s://%s", srv.Protocal, node.Address)
		req2Endpoinrt, err = copyRequest(req, urlStr, "PUT")
		if err != nil {
			log.Println(err)
			w.Write([]byte("forward request failed : " + err.Error()))
			return
		}
		response, err := srv.client.Do(req2Endpoinrt)
		if err != nil {
			log.Println(err)
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

func (srv *ApiService) HandleDeleteMethod(w http.ResponseWriter, req *http.Request, _ httprouter.Params) {
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
		urlStr := fmt.Sprintf("%s://%s", srv.Protocal, node.Address)
		req2Endpoinrt, err = copyRequest(req, urlStr, "DELETE")
		if err != nil {
			log.Println(err)
			w.Write([]byte("forward request failed : " + err.Error()))
			return
		}
		response, err := srv.client.Do(req2Endpoinrt)
		if err != nil {
			log.Println(err)
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
	switch srv.LoadBlanceType {
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
			log.Println("recover panic : ", re)
		}
	}()
	srv.Lock()
	defer srv.Unlock()
	if _, ok := srv.BackendMap[n.NodeId]; ok {
		return fmt.Errorf("node [%d] is exist", n.NodeId)
	} else {
		srv.BackendMap[n.NodeId] = n
	}
	return nil
}

func (srv *ApiService) RemoveNodoe(n *Node) error {
	if n == nil {
		return fmt.Errorf("node info is nil")
	}
	defer func() {
		if re := recover(); re != nil {
			log.Println("recover panic : ", re)
		}
	}()
	srv.Lock()
	defer srv.Unlock()
	delete(srv.BackendMap, n.NodeId)
	return nil
}
