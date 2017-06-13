package server

import (
	"net"
	"sync"

	"github.com/crunchydata/crunchy-proxy/config"
	"github.com/crunchydata/crunchy-proxy/util/log"
)

type Server struct {
	admin     *AdminServer
	proxy     *ProxyServer
	waitGroup *sync.WaitGroup
}

func NewServer() *Server {
	s := &Server{
		waitGroup: &sync.WaitGroup{},
	}

	s.admin = NewAdminServer(s)

	s.proxy = NewProxyServer(s)

	return s
}

func (s *Server) Start() {
	proxyConfig := config.GetProxyConfig()
	adminConfig := config.GetAdminConfig()

	log.Info("Admin Server Starting...")
	adminListener, err := net.Listen("tcp", adminConfig.HostPort)

	if err != nil {
		log.Error(err.Error())
	}

	s.waitGroup.Add(1)
	go s.admin.Serve(adminListener)

	log.Info("Proxy Server Starting...")
	proxyListener, err := net.Listen("tcp", proxyConfig.HostPort)

	s.waitGroup.Add(1)
	go s.proxy.Serve(proxyListener)

	s.waitGroup.Wait()

	log.Info("Server Exiting...")
}
