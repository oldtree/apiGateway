package gateway

import (
	"fmt"
	"log"
	"net/http"

	"github.com/julienschmidt/httprouter"
)

//RoutePathInfo : route description info ,get from config file or ETCD watch client
type RoutePathInfo struct {
	Description string `json:"description,omitempty"`
	ActionType  uint   `json:"action_type,omitempty"`
	InnerPath   string `json:"inner_path,omitempty"`
	InnerMethod string `json:"inner_method,omitempty"` //inner service request method
	OuterPath   string `json:"outer_path,omitempty"`
	OuterMethod string `json:"method,omitempty"` // outer access request method
	Auth        bool   `json:"auth,omitempty"`
	Status      uint   `json:"status,omitempty"`
}

//Route : route is a real functional compoment
//router :include router(with route parsed info)
//srvBelong :node register witch srv
//RouterMappingInfo : route info
type Route struct {
	router            *httprouter.Router
	RouterMappingInfo []*RoutePathInfo
	srvBelong         *ApiService
}

//ServeHTTP : interface method for http.Handler
func (r *Route) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	r.router.ServeHTTP(w, req)
}

//BuildRouteInfo : use RouterMappingInfo to build functional *httprouter.Router
func (r *Route) BuildRouteInfo() {
	for _, value := range r.RouterMappingInfo {
		outerPath := fmt.Sprintf("/%s/%s", r.srvBelong.ServiceName, value.OuterPath)
		switch value.InnerMethod {
		case "GET", "get":
			switch value.OuterMethod {
			case "GET", "get":
				r.router.GET(outerPath, r.srvBelong.HandleGetMethod)
			case "POST", "post":
				r.router.POST(outerPath, r.srvBelong.HandlePostMethod)
			case "PUT", "put":
				r.router.PUT(outerPath, r.srvBelong.HandlePutMethod)
			case "DELETE", "delete":
				r.router.DELETE(outerPath, r.srvBelong.HandleDeleteMethod)
			default:
				r.router.GET(outerPath, r.srvBelong.HandleGetMethod)
			}
		case "POST", "post":
			switch value.OuterMethod {
			case "GET", "get":
				r.router.GET(outerPath, r.srvBelong.HandleGetMethod)
			case "POST", "post":
				r.router.POST(outerPath, r.srvBelong.HandlePostMethod)
			case "PUT", "put":
				r.router.PUT(outerPath, r.srvBelong.HandlePutMethod)
			case "DELETE", "delete":
				r.router.DELETE(outerPath, r.srvBelong.HandleDeleteMethod)
			default:
				r.router.POST(outerPath, r.srvBelong.HandlePostMethod)
			}
		case "PUT", "put":
			switch value.OuterMethod {
			case "GET", "get":
				r.router.GET(outerPath, r.srvBelong.HandleGetMethod)
			case "POST", "post":
				r.router.POST(outerPath, r.srvBelong.HandlePostMethod)
			case "PUT", "put":
				r.router.PUT(outerPath, r.srvBelong.HandlePutMethod)
			case "DELETE", "delete":
				r.router.DELETE(outerPath, r.srvBelong.HandleDeleteMethod)
			default:
				r.router.PUT(outerPath, r.srvBelong.HandlePutMethod)
			}
		case "DELETE", "delete":
			switch value.OuterMethod {
			case "GET", "get":
				r.router.GET(outerPath, r.srvBelong.HandleGetMethod)
			case "POST", "post":
				r.router.POST(outerPath, r.srvBelong.HandlePostMethod)
			case "PUT", "put":
				r.router.PUT(outerPath, r.srvBelong.HandlePutMethod)
			case "DELETE", "delete":
				r.router.DELETE(outerPath, r.srvBelong.HandleDeleteMethod)
			default:
				r.router.DELETE(outerPath, r.srvBelong.HandleDeleteMethod)
			}
		default:
			log.Printf("path method [%s]is not support \n", value.OuterMethod)
		}
	}

}

func NewRoute(srv *ApiService) *Route {
	router := httprouter.New()
	return &Route{
		router:            router,
		RouterMappingInfo: *new([]*RoutePathInfo),
		srvBelong:         srv,
	}
}
