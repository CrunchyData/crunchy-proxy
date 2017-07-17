/*
Copyright 2017 Crunchy Data Solutions, Inc.
Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at
http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

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
