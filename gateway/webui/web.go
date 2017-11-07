package webui

import (
	"net/http"
)

func Server() {

	mux := http.NewServeMux()
	mux.HandleFunc("/", func(resp http.ResponseWriter, req *http.Request) {
		resp.Write([]byte("gateway"))
	})
	mux.HandleFunc("/admin", func(resp http.ResponseWriter, req *http.Request) {
		resp.Write([]byte("admin"))
	})
	mux.HandleFunc("/statistic", func(resp http.ResponseWriter, req *http.Request) {
		resp.Write([]byte("statistic"))
	})
	mux.HandleFunc("/service", func(resp http.ResponseWriter, req *http.Request) {
		resp.Write([]byte("statistic"))
	})
	mux.HandleFunc("/service/endpoint", func(resp http.ResponseWriter, req *http.Request) {
		resp.Write([]byte("statistic"))
	})
	http.ListenAndServe(":8080", mux)
	return
}
