/*
Copyright 2016 Crunchy Data Solutions, Inc.
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
	"log"
	"net"
)

func ReturnConnection(ch chan int, connIndex int) {
	log.Printf("returning poolIndex %d\n", connIndex)
	ch <- connIndex
}

func SetupPools(c *config.Config) {
	if !c.Pool.Enabled {
		log.Println("[pool] pooling not enabled")
		return
	}

	log.Println("[pool] pooling enabled")

	for i := 0; i < len(c.Replicas); i++ {
		setupPoolForNode(c, &c.Replicas[i])
	}

	setupPoolForNode(c, &c.Master)

}

func setupPoolForNode(c *config.Config, node *config.Node) {
	var err error

	node.Pool.Channel = make(chan int, c.Pool.Capacity)
	node.Pool.Connections = make([]*net.TCPConn, c.Pool.Capacity)
	for j := 0; j < c.Pool.Capacity; j++ {
		node.Pool.Channel <- j
		//add a connection to the node pool
		log.Printf("[pool] adding conn to node %s pool\n", node.IPAddr)
		node.Pool.Connections[j], err = node.GetConnection()
		if err != nil {
			log.Println("error in getting pool conn for node " + err.Error())
		}
		Authenticate(c, node, node.Pool.Connections[j])
	}
}
