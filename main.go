package main

import (
	"net/http"
	"os"
	"os/signal"
	"runtime"
	"syscall"
	"time"

	"encoding/json"
	"fmt"
	"io/ioutil"

	"github.com/oldtree/apiGateway/gateway"
)

func Init() {
	enginx := gateway.NewEngine()
	go enginx.Doorman()
	go func() {
		time.Sleep(time.Second * 5)
		e := new(gateway.Event)
		e.EventType = gateway.EventServiceAdd
		e.TimeStamp = time.Now().String()
		newservice := gateway.NewServiceInfo()
		data, err := ioutil.ReadFile("sample.json")
		if err != nil {
			fmt.Println(err)
			return
		}
		err = json.Unmarshal(data, newservice)
		if err != nil {
			fmt.Println(err)
			return
		}
		srv := gateway.NewApiService()
		srv.MappingApiService(newservice)
		srv.R.BuildRouteInfo()
		e.Content = srv
		enginx.Notice <- e
	}()
	newservice := gateway.NewServiceInfo()
	data, err := ioutil.ReadFile("sample.json")
	if err != nil {
		fmt.Println(err)
		return
	}
	err = json.Unmarshal(data, newservice)
	if err != nil {
		fmt.Println(err)
		return
	}
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
