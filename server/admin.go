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
	"database/sql"
	"fmt"
	"net"
	"time"

	_ "github.com/lib/pq" // required
	"golang.org/x/net/context"
	"google.golang.org/grpc"

	"github.com/crunchydata/crunchy-proxy/common"
	"github.com/crunchydata/crunchy-proxy/config"
	pb "github.com/crunchydata/crunchy-proxy/server/serverpb"
	"github.com/crunchydata/crunchy-proxy/util/grpcutil"
	"github.com/crunchydata/crunchy-proxy/util/log"
)

type AdminServer struct {
	grpc       *grpc.Server
	server     *Server
	nodeHealth map[string]bool
}

func NewAdminServer(s *Server) *AdminServer {
	admin := &AdminServer{
		server:     s,
		nodeHealth: make(map[string]bool, 0),
	}

	admin.grpc = grpc.NewServer()

	pb.RegisterAdminServer(admin.grpc, admin)

	return admin
}

func (s *AdminServer) Nodes(ctx context.Context, req *pb.NodeRequest) (*pb.NodeResponse, error) {
	var response pb.NodeResponse

	response.Nodes = make(map[string]string, 0)

	for name, node := range config.GetNodes() {
		response.Nodes[name] = node.HostPort
	}

	return &response, nil
}

func (s *AdminServer) Pools(ctx context.Context, req *pb.PoolRequest) (*pb.PoolResponse, error) {
	var response pb.PoolResponse

	response.Pools = append(response.Pools, "Pool")

	return &response, nil
}

func (s *AdminServer) Shutdown(req *pb.ShutdownRequest, stream pb.Admin_ShutdownServer) error {
	// Stop the Proxy Server
	s.server.proxy.Stop()

	// Stop the Admin grpc Server
	s.grpc.Stop()

	return nil
}

func (s *AdminServer) Health(ctx context.Context, req *pb.HealthRequest) (*pb.HealthResponse, error) {
	var response pb.HealthResponse

	response.Health = s.nodeHealth

	return &response, nil
}

func (s *AdminServer) Statistics(context.Context, *pb.StatisticsRequest) (*pb.StatisticsResponse, error) {
	var response pb.StatisticsResponse

	response.Queries = s.server.proxy.Stats()

	return &response, nil
}

func (s *AdminServer) Version(context.Context, *pb.VersionRequest) (*pb.VersionResponse, error) {
	var response pb.VersionResponse

	response.Version = "1.0.0beta"

	return &response, nil
}

func (s *AdminServer) Serve(l net.Listener) {
	log.Infof("Admin Server listening on: %s", l.Addr())
	defer s.server.waitGroup.Done()

	go s.startHealthCheck()

	err := s.grpc.Serve(l)
	l.Close()

	if !grpcutil.IsClosedConnection(err) {
		log.Infof("Server Error: %s", err)
	}
}

func (s *AdminServer) startHealthCheck() {
	nodes := config.GetNodes()
	hcConfig := config.GetHealthCheckConfig()

	for {
		for name, node := range nodes {
			/* Connect to node */
			conn, err := getDBConnection(node)

			if err != nil {
				log.Errorf("healthcheck: error creating connection to '%s'", name)
				log.Errorf("healthcheck: %s", err.Error())
			}

			/* Perform Health Check Query */
			rows, err := conn.Query(hcConfig.Query)

			if err != nil {
				log.Errorf("healthcheck: query failed: %s", err.Error())
				s.nodeHealth[name] = false
				continue
			}

			rows.Close()

			/* Update health status */
			s.nodeHealth[name] = true

			conn.Close()
		}

		time.Sleep(time.Duration(hcConfig.Delay) * time.Second)
	}
}

func getDBConnection(node common.Node) (*sql.DB, error) {
	host, port, _ := net.SplitHostPort(node.HostPort)
	creds := config.GetCredentials()

	connectionString := fmt.Sprintf("host=%s port=%s ", host, port)
	connectionString += fmt.Sprintf(" user=%s", creds.Username)
	connectionString += fmt.Sprintf(" database=%s", creds.Database)

	connectionString += fmt.Sprintf(" sslmode=%s", creds.SSL.SSLMode)
	connectionString += " application_name=proxy_healthcheck"

	if creds.Password != "" {
		connectionString += fmt.Sprintf(" password=%s", creds.Password)
	}

	if creds.SSL.Enable {
		connectionString += fmt.Sprintf(" sslcert=%s", creds.SSL.SSLCert)
		connectionString += fmt.Sprintf(" sslkey=%s", creds.SSL.SSLKey)
		connectionString += fmt.Sprintf(" sslrootcert=%s", creds.SSL.SSLRootCA)
	}

	/* Build connection string. */
	for key, value := range creds.Options {
		connectionString += fmt.Sprintf(" %s=%s", key, value)
	}

	log.Debugf("healthcheck: Opening connection with parameters: %s",
		connectionString)

	dbConn, err := sql.Open("postgres", connectionString)

	if err != nil {
		log.Errorf("healthcheck: Error creating connection : %s", err.Error())
	}

	return dbConn, err
}
