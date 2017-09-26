package main

import (
	"crypto/tls"
	"net/http"
	"os"
	"os/signal"
	"runtime"
	"syscall"
	"time"

	"encoding/json"
	"fmt"
	"io/ioutil"

	"flag"

	"github.com/oldtree/apiGateway/gateway"
	"github.com/oldtree/apiGateway/gateway/servicedesc"
)

var config = flag.String("config", "config.json", "config file path")

func Init() {
	enginx := gateway.NewEngine()
	go enginx.Doorman()
	go func() {
		time.Sleep(time.Second * 5)
		e := new(gateway.Event)
		e.EventType = gateway.EventServiceAdd
		e.TimeStamp = time.Now().String()
		newservice := servicedesc.NewServiceInfo()
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
	newservice := servicedesc.NewServiceInfo()
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

	//https://blog.cloudflare.com/exposing-go-on-the-internet/
	tlsconfig := &tls.Config{
		// Causes servers to use Go's default ciphersuite preferences,
		// which are tuned to avoid attacks. Does nothing on clients.
		PreferServerCipherSuites: true,
		// Only use curves which have assembly implementations
		CurvePreferences: []tls.CurveID{
			tls.CurveP256,
			tls.X25519, // Go 1.8 only
		},
		MinVersion: tls.VersionTLS12,
		CipherSuites: []uint16{
			tls.TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384,
			tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
			tls.TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305, // Go 1.8 only
			tls.TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305,   // Go 1.8 only
			tls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256,
			tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
		},
	}
	srv := http.Server{
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  120 * time.Second,
		TLSConfig:    tlsconfig,
		Handler:      enginx,
		Addr:         ":2222",
	}
	srv.ListenAndServe()
}

func main() {
	flag.Parse()
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
