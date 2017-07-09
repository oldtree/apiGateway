package main

import (
	"net/http"
	"os"
	"os/signal"
	"runtime"
	"syscall"

	"fmt"

	"github.com/oldtree/apiGateway/gateway"
)

func Init() {
	enginx := gateway.NewEngine()
	enginx.AddService(gateway.DefaultService)
	http.ListenAndServe(":2222", enginx)
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
	select {
	case si := <-sc:
		fmt.Println("recv signal : ", si.String())
	}

}
