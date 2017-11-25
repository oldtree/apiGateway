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
	_ "net/http/pprof"

	"flag"

	"github.com/FlyCynomys/tools/log"
	"github.com/oldtree/apiGateway/gateway"
	cfg "github.com/oldtree/apiGateway/gateway/config"
	"github.com/oldtree/apiGateway/gateway/description/servicedesc"
	"github.com/oldtree/apiGateway/gateway/utils"
)

var configfile = flag.String("config", "config.json", "config file path")

func Init() {
	enginx := gateway.NewEngine()
	go enginx.Doorman()
	go func() {
		time.Sleep(time.Second * 5)
		e := new(utils.Event)
		e.EventType = utils.EventServiceAdd
		e.TimeStamp = time.Now().String()
		newservice := servicedesc.NewServiceDesc()
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
		err = srv.InitServiceHttpClient()
		if err != nil {
			log.Error(err)
			return
		}
		srv.R.BuildRouteInfo()
		e.Content = srv
		enginx.Notice <- e

		nodee := new(utils.Event)
		nodee.EventType = utils.EventNodeAdd
		nodee.TimeStamp = time.Now().String()
		data, err = ioutil.ReadFile("node.json")
		if err != nil {
			return
		}
		newnode := servicedesc.NewNodeDesc()
		json.Unmarshal(data, newnode)
		node := gateway.NewDefaultNode(newnode.SrvName, newnode.Address, newnode.Id, newnode.Hostname)
		nodee.Content = node
		enginx.Notice <- nodee
	}()
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
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  120 * time.Second,
		TLSConfig:    tlsconfig,
		Handler:      enginx,
		Addr:         cfg.GetConfig().Port,
	}
	srv.ListenAndServe()
}

func main() {
	flag.Parse()
	runtime.GOMAXPROCS(runtime.NumCPU())
	if err := cfg.LoadConfig(*configfile); err != nil {
		log.Error("load condig file failed : ", err.Error())
		return
	}
	log.Info("start init default server", cfg.GetConfig().Port)
	Init()
	log.Info("end init default server", cfg.GetConfig().EtcdConfig)

	sc := make(chan os.Signal, 1)

SYSCALL:

	signal.Notify(sc,
		syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGQUIT,
		syscall.SIGKILL,
		syscall.SIGTRAP,
	)
	select {
	case si := <-sc:
		switch si {
		case syscall.SIGINT:
			log.Info("signal : ", si.String())
			return
		case syscall.SIGTERM:
			log.Info("signal : ", si.String())
			goto SYSCALL
		case syscall.SIGQUIT:
			log.Info("signal : ", si.String())
			return
		case syscall.SIGKILL:
			log.Info("signal : ", si.String())
			return
		case syscall.SIGTRAP:
			log.Info("signal : ", si.String())
			goto SYSCALL
		default:
			goto SYSCALL
		}
	}

}
