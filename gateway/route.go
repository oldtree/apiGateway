package gateway

import (
	"log"
	"net/http"

	"github.com/julienschmidt/httprouter"
)

type RoutePathInfo struct {
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
	RouterMappingInfo map[string]*RoutePathInfo
	srv               *ApiService
}

func (r *Route) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	r.router.ServeHTTP(w, req)
}

func (r *Route) FillRouterInfo() {

}

func (r *Route) BuildRouteInfo() {
	for key, value := range r.RouterMappingInfo {
		switch value.OuterMethod {
		case "GET", "get":
			switch value.InnerMethod {
			case "GET", "get":
				r.router.GET(value.OuterPath, r.srv.HandleGetMethod)
			case "POST", "post":
				r.router.POST(value.OuterPath, r.srv.HandlePostMethod)
			case "PUT", "put":
				r.router.PUT(value.OuterPath, r.srv.HandlePutMethod)
			case "DELETE", "delete":
				r.router.DELETE(value.OuterPath, r.srv.HandleDeleteMethod)
			default:
				r.router.GET(value.OuterPath, r.srv.HandleGetMethod)
			}
		case "POST", "post":
			switch value.InnerMethod {
			case "GET", "get":
				r.router.GET(value.OuterPath, r.srv.HandleGetMethod)
			case "POST", "post":
				r.router.POST(value.OuterPath, r.srv.HandlePostMethod)
			case "PUT", "put":
				r.router.PUT(value.OuterPath, r.srv.HandlePutMethod)
			case "DELETE", "delete":
				r.router.DELETE(value.OuterPath, r.srv.HandleDeleteMethod)
			default:
				r.router.POST(value.OuterPath, r.srv.HandlePostMethod)
			}
		case "PUT", "put":
			switch value.InnerMethod {
			case "GET", "get":
				r.router.GET(value.OuterPath, r.srv.HandleGetMethod)
			case "POST", "post":
				r.router.POST(value.OuterPath, r.srv.HandlePostMethod)
			case "PUT", "put":
				r.router.PUT(value.OuterPath, r.srv.HandlePutMethod)
			case "DELETE", "delete":
				r.router.DELETE(value.OuterPath, r.srv.HandleDeleteMethod)
			default:
				r.router.PUT(value.OuterPath, r.srv.HandlePutMethod)
			}
		case "DELETE", "delete":
			switch value.InnerMethod {
			case "GET", "get":
				r.router.GET(value.OuterPath, r.srv.HandleGetMethod)
			case "POST", "post":
				r.router.POST(value.OuterPath, r.srv.HandlePostMethod)
			case "PUT", "put":
				r.router.PUT(value.OuterPath, r.srv.HandlePutMethod)
			case "DELETE", "delete":
				r.router.DELETE(value.OuterPath, r.srv.HandleDeleteMethod)
			default:
				r.router.DELETE(value.OuterPath, r.srv.HandleDeleteMethod)
			}
		default:
			log.Printf("path [%s] method [%s]is not support \n", key, value.OuterMethod)
		}
	}

}

func NewRoute(srv *ApiService) *Route {
	router := httprouter.New()
	return &Route{
		router:            router,
		RouterMappingInfo: make(map[string]*RoutePathInfo, 16),
		srv:               srv,
	}
}
