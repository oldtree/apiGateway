package gateway

import (
	"net/http"
	"testing"

	"github.com/julienschmidt/httprouter"
)

func TestRoute_ServeHTTP(t *testing.T) {
	type fields struct {
		router            *httprouter.Router
		RouterMappingInfo []*RoutePathInfo
		srvBelong         *ApiService
	}
	type args struct {
		w   http.ResponseWriter
		req *http.Request
	}
	tests := []struct {
		name   string
		fields fields
		args   args
	}{
	// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &Route{
				router:            tt.fields.router,
				RouterMappingInfo: tt.fields.RouterMappingInfo,
				srvBelong:         tt.fields.srvBelong,
			}
			r.ServeHTTP(tt.args.w, tt.args.req)
		})
	}
}
