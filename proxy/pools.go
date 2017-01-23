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
package proxy

import (
	"github.com/crunchydata/crunchy-proxy/config"
	"github.com/golang/glog"
	"net"
)

func ReturnConnection(ch chan int, connIndex int) {
	glog.V(2).Infof("returning poolIndex %d\n", connIndex)
	ch <- connIndex
}

func SetupPools() {
	for i := 0; i < len(config.Cfg.Replicas); i++ {
		SetupPoolForNode(&config.Cfg.Replicas[i])
	}

	SetupPoolForNode(&config.Cfg.Master)
}

func connectPool(hostPort string) (*net.TCPConn, error) {
	var address *net.TCPAddr
	var connection *net.TCPConn
	var err error

	// TODO: Add some error checking here for each of these statements.
	address, err = net.ResolveTCPAddr("tcp4", hostPort)
	connection, err = net.DialTCP("tcp", nil, address)

	return connection, err
}

func SetupPoolForNode(node *config.Node) {
	var connection *net.TCPConn
	var err error

	glog.Infof("[pool] Setting up connection pool for %s", node.HostPort)

	/*
	 * Each node has its own connection pool. The size of the pool is defined
	 * by 'Capacity'. To set up each connection, first authenticate and second
	 * add the connection the pool.
	 */
	node.Pool.Channel = make(chan int, config.Cfg.Pool.Capacity)
	node.Pool.Connections = make([]*net.TCPConn, config.Cfg.Pool.Capacity)

	for j := 0; j < config.Cfg.Pool.Capacity; j++ {

		// Create a new pool connection.
		connection, err = connectPool(node.HostPort)

		if err != nil {
			glog.Errorf("[pool] Error creating connection for node: %s\n", err.Error())
		}

		// Authenticate the new pool connection.
		glog.Infof("[pool] Authenticating connection %d.\n", j)
		authenticated := Authenticate(connection)

		// Add the connection to the nodes connection pool.
		if authenticated {
			glog.Infof("[pool] Adding connection %d to pool.\n", j)
			node.Pool.Connections[j] = connection
			node.Pool.Channel <- j
		} else {
			glog.Errorln("[pool] Error occurred authenticating.")
		}

	}
}
