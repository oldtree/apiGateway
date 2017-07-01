package main

import (
	"net/http"
	"os"
	"os/signal"
	"runtime"
	"syscall"

	"github.com/oldtree/apiGateway/gateway"
)

func Init() {
	eginx := gateway.NewEngine()
	eginx.AddService(gateway.DefaultService)
	http.ListenAndServe(":2222", eginx)
}

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())

	Init()

	sc := make(chan os.Signal, 1)
	signal.Notify(sc,
		syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGQUIT,
	)
	<-sc
}
