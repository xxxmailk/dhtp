package serve

import (
	"dhtp/conf"
	"log"
	"net"
	"net/http"
	"sync"
)

func HTTPStart(wg *sync.WaitGroup) {
	defer func() {
		if err := recover(); err != nil {
			panic(err)
		}
	}()
	defer wg.Done()
	c := conf.GetConf()
	listen := net.JoinHostPort(c.Http.HttpIP, c.Http.HttpPort)
	http.Handle("/", http.FileServer(http.Dir(c.Http.MountPath)))
	log.Printf("starting http server %s and handle on path: %s",listen, c.Http.MountPath)
	panic(http.ListenAndServe(listen, nil))
}
