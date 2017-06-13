package server

import (
	"net"

	"github.com/crunchydata/crunchy-proxy/proxy"
	"github.com/crunchydata/crunchy-proxy/util/log"
)

type ProxyServer struct {
	ch       chan bool
	server   *Server
	listener net.Listener
}

func NewProxyServer(s *Server) *ProxyServer {
	proxy := &ProxyServer{}
	proxy.ch = make(chan bool)
	proxy.server = s

	return proxy
}

func (s *ProxyServer) Serve(l net.Listener) error {
	log.Infof("Proxy Server listening on: %s", l.Addr())
	defer s.server.waitGroup.Done()
	s.listener = l

	p := proxy.NewProxy()

	for {

		select {
		case <-s.ch:
			return nil
		default:
		}

		conn, err := l.Accept()

		if err != nil {
			continue
		}

		go p.HandleConnection(conn)
	}
}

func (s *ProxyServer) Stop() {
	s.listener.Close()
	close(s.ch)
}
