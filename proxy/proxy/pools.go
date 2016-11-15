package proxy

import (
	"github.com/crunchydata/crunchy-proxy/proxy/config"
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
	var err error

	for i := 0; i < len(c.Replicas); i++ {
		c.Replicas[i].Pool.Channel = make(chan int, c.Pool.Capacity)
		c.Replicas[i].Pool.Connections = make([]*net.TCPConn, c.Pool.Capacity)
		for j := 0; j < c.Pool.Capacity; j++ {
			c.Replicas[i].Pool.Channel <- j
			//add a connection to the node pool
			log.Printf("[pool] adding conn to replica %s pool\n", c.Replicas[i].IPAddr)
			c.Replicas[i].Pool.Connections[j], err = c.Replicas[i].GetConnection()
			if err != nil {
				log.Println("error in getting pool conn for replica " + err.Error())
			}
			Authenticate(c, c.Replicas[i], c.Replicas[i].Pool.Connections[j])
		}
	}
}
