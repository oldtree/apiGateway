package gateway

import (
	"fmt"
	"log"
	"net/http"

	"github.com/julienschmidt/httprouter"
)

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

type Route struct {
	router            *httprouter.Router
	RouterMappingInfo []*RoutePathInfo
	srv               *ApiService
}

func (r *Route) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	r.router.ServeHTTP(w, req)
}

func (r *Route) FillRouterInfo() {

}

func (r *Route) BuildRouteInfo() {
	for _, value := range r.RouterMappingInfo {
		outerPath := fmt.Sprintf("/%s/%s", r.srv.ServiceName, value.OuterPath)
		switch value.InnerMethod {
		case "GET", "get":
			switch value.OuterMethod {
			case "GET", "get":
				r.router.GET(outerPath, r.srv.HandleGetMethod)
			case "POST", "post":
				r.router.POST(outerPath, r.srv.HandlePostMethod)
			case "PUT", "put":
				r.router.PUT(outerPath, r.srv.HandlePutMethod)
			case "DELETE", "delete":
				r.router.DELETE(outerPath, r.srv.HandleDeleteMethod)
			default:
				r.router.GET(outerPath, r.srv.HandleGetMethod)
			}
		case "POST", "post":
			switch value.OuterMethod {
			case "GET", "get":
				r.router.GET(outerPath, r.srv.HandleGetMethod)
			case "POST", "post":
				r.router.POST(outerPath, r.srv.HandlePostMethod)
			case "PUT", "put":
				r.router.PUT(outerPath, r.srv.HandlePutMethod)
			case "DELETE", "delete":
				r.router.DELETE(outerPath, r.srv.HandleDeleteMethod)
			default:
				r.router.POST(outerPath, r.srv.HandlePostMethod)
			}
		case "PUT", "put":
			switch value.OuterMethod {
			case "GET", "get":
				r.router.GET(outerPath, r.srv.HandleGetMethod)
			case "POST", "post":
				r.router.POST(outerPath, r.srv.HandlePostMethod)
			case "PUT", "put":
				r.router.PUT(outerPath, r.srv.HandlePutMethod)
			case "DELETE", "delete":
				r.router.DELETE(outerPath, r.srv.HandleDeleteMethod)
			default:
				r.router.PUT(outerPath, r.srv.HandlePutMethod)
			}
		case "DELETE", "delete":
			switch value.OuterMethod {
			case "GET", "get":
				r.router.GET(outerPath, r.srv.HandleGetMethod)
			case "POST", "post":
				r.router.POST(outerPath, r.srv.HandlePostMethod)
			case "PUT", "put":
				r.router.PUT(outerPath, r.srv.HandlePutMethod)
			case "DELETE", "delete":
				r.router.DELETE(outerPath, r.srv.HandleDeleteMethod)
			default:
				r.router.DELETE(outerPath, r.srv.HandleDeleteMethod)
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
		srv:               srv,
	}
}
