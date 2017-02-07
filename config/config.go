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
package config

import (
	"encoding/json"
	"errors"
	"flag"
	"github.com/crunchydata/crunchy-proxy/adapter"
	"github.com/golang/glog"
	"io/ioutil"
	"log"
	"math/rand"
	"net"
	"os"
	"sync"
	"time"
)

const (
	DEFAULT_READ_ANNOTATION   string = "read"
	DEFAULT_START_ANNOTATION  string = "start"
	DEFAULT_FINISH_ANNOTATION string = "finish"
)

type NodeStats struct {
	Queries int `json:"-"`
}

type NodePool struct {
	Channel     chan int       `json:"-"`
	Connections []*net.TCPConn `json:"-"`
}

type PoolConfig struct {
	Capacity int `json:"capacity"`
}

type PGCredentials struct {
	Username string `json:"username"`
	Password string `json:"password"`
	Database string `json:"database"`
}

type Healthcheck struct {
	Delay int    `json:"delay"` //remote host:port
	Query string `json:"query"`
}

type Node struct {
	HostPort     string            `json:"hostport"` //remote host:port
	Metadata     map[string]string `json:"metadata"`
	Healthy      bool              `json:"-"`
	HCConnection net.Conn          `json:"-"`
	TCPAddr      *net.TCPAddr      `json:"-"`
	TCPConn      *net.TCPConn      `json:"-"`
	Pool         NodePool          `json:"-"`
	Stats        NodeStats         `json:"-"`
}

type Adapter struct {
	AdapterType string                 `json:"adaptertype"`
	Metadata    map[string]interface{} `json:"metadata"`
}

type Config struct {
	Name             string          `json:"name"`
	HostPort         string          `json:"hostport"`      //listen on host:port
	AdminHostPort    string          `json:"adminhostport"` //listen on host:port
	ReadAnnotation   string          `json:"readannotation"`
	StartAnnotation  string          `json:"startannotation"`
	FinishAnnotation string          `json:"finishannotation"`
	Credentials      PGCredentials   `json:"credentials"`
	Pool             PoolConfig      `json:"pool"`
	Master           Node            `json:"master"`
	Replicas         []Node          `json:"replicas"`
	Adapters         []Adapter       `json:"adapters"`
	Healthcheck      Healthcheck     `json:"healthcheck"`
	Adapter          adapter.Adapter `json:"-"`
}

var Cfg Config

func (c Config) Print() {
	str, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		glog.Errorln(err)
	}
	glog.V(2).Infoln(string(str))

}

func (c Config) PrintNodeInfo() {
	// Print the master node information.
	glog.Infoln("[config] ---- Master Node Information ----")
	glog.Infof("[config] master host = %s\n", c.Master.HostPort)

	// Print the replica node information.
	glog.Infoln("[config] ---- Replica Node Information ----")
	for i, replica := range c.Replicas {
		glog.Infof("[config] replica %d host = %s\n", i, replica.HostPort)
	}
}

func PrintExample() {
	var ds = make([]Adapter, 1)
	ds[0] = Adapter{
		AdapterType: "audit",
	}
	ds[0].Metadata = make(map[string]interface{})
	ds[0].Metadata["Age"] = 6
	ds[0].Metadata["Filepath"] = "/tmp/audit.log"
	var pool = PoolConfig{
		Capacity: 2}
	cred := PGCredentials{
		Username: "logging",
		Password: "audit",
		Database: "database1"}

	var ms = Node{
		HostPort: "master:5432"}

	ms.Metadata = make(map[string]string)

	var rs = make([]Node, 2)
	rs[0] = Node{
		HostPort: "replica1:5432"}
	rs[0].Metadata = make(map[string]string)
	rs[1] = Node{
		HostPort: "replica2:5432"}
	rs[1].Metadata = make(map[string]string)
	var hs Healthcheck
	hs.Delay = 10
	hs.Query = "select now()"

	c := Config{
		Name:        "sampleconfig",
		HostPort:    "localhost:5432",
		Master:      ms,
		Pool:        pool,
		Credentials: cred,
		Replicas:    rs,
		Healthcheck: hs,
		Adapters:    ds}

	str, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		glog.Errorln(err)
	}
	glog.V(2).Infoln(string(str))
}

func ReadConfig() {

	var filePath string
	flag.StringVar(&filePath, "config", "", "a configuration file")
	flag.Parse()
	glog.V(2).Infoln("[config]" + filePath + " is the config path")
	if filePath == "" {
		glog.Errorln("-config command option required")
		os.Exit(1)
	}

	var err error
	var byt []byte

	if byt, err = ioutil.ReadFile(filePath); err != nil {
		panic(err)
	}
	if err = json.Unmarshal(byt, &Cfg); err != nil {
		panic(err)
	}

	Cfg.Master.TCPAddr, err = net.ResolveTCPAddr("tcp4", Cfg.Master.HostPort)
	checkError(err)
	for n := 0; n < len(Cfg.Replicas); n++ {
		Cfg.Replicas[n].TCPAddr, err = net.ResolveTCPAddr("tcp4", Cfg.Replicas[n].HostPort)
		checkError(err)
	}

	if Cfg.ReadAnnotation == "" {
		Cfg.ReadAnnotation = DEFAULT_READ_ANNOTATION
		glog.Infof("[config] ReadAnnotation is not specified, using default: %s\n",
			Cfg.ReadAnnotation)
	}

	if Cfg.StartAnnotation == "" {
		Cfg.StartAnnotation = DEFAULT_START_ANNOTATION
		glog.Infof("[config] StartAnnotation is not specified, using default: %s\n",
			Cfg.StartAnnotation)
	}

	if Cfg.FinishAnnotation == "" {
		Cfg.FinishAnnotation = DEFAULT_FINISH_ANNOTATION
		glog.Infof("[config] FinishAnnotation is not specified, using default: %s\n",
			Cfg.FinishAnnotation)
	}

	glog.V(2).Infof("[config] %s is the ReadAnnotation", Cfg.ReadAnnotation)
	glog.V(2).Infof("[config] %s is the StartAnnotation", Cfg.StartAnnotation)
	glog.V(2).Infof("[config] %s is the FinishAnnotation", Cfg.FinishAnnotation)

}

func (c *Config) GetAllConnections() {
	var err error
	glog.V(2).Infoln("dialing " + c.Master.HostPort)
	c.Master.TCPConn, err = net.DialTCP("tcp", nil, c.Master.TCPAddr)
	if err != nil {
		glog.Errorln(err.Error())
	}

}

func (c *Config) SetupAdapters() {
	glog.Infoln("---- Setup Adapters ----")

	var ds []adapter.Decorator = make([]adapter.Decorator, 0)

	for i := 0; i < len(c.Adapters); i++ {
		glog.V(2).Infof("---- Setup '%q' Adapter ----", c.Adapters[i])
		switch c.Adapters[i].AdapterType {
		case "audit":

			glog.V(2).Infof("---- added audit adapter")
			ds = append(ds, adapter.Audit(c.Adapters[i].Metadata, log.New(os.Stdout, "[audit adapter]", 0)))
		default:
			glog.Errorf("Invalid adapter: %s", c.Adapters[i])
		}
	}

	c.Adapter = adapter.ThisDecorate(adapter.MockAdapter{}, ds)
}

//eventually this would be a load balancer algorithm function
func (c *Config) GetNextNode(writeCase bool) (*Node, error) {

	var err error
	var rCnt = len(c.Replicas)

	if writeCase || rCnt == 0 {
		if !c.Master.Healthy {
			glog.V(2).Infoln("master is unhealthy!")
			return &c.Master, errors.New("unhealthy master")
		}
		glog.V(2).Infoln("writeCase so using master as node...")
		return &c.Master, err
	}

	var replicaHealthy = false

	for i := 0; i < len(c.Replicas); i++ {
		if c.Replicas[i].Healthy {
			glog.V(2).Infoln("picked replica that was healthy")
			replicaHealthy = true
		}
	}

	if rCnt == 1 && replicaHealthy == false {
		glog.V(2).Infoln("no replicas are healthy..using master")
		if !c.Master.Healthy {
			glog.V(2).Infoln("master is unhealthy!")
			return &c.Master, errors.New("unhealthy master")
		}
		return &c.Master, err
	}

	//for now, use a simple random number generator to pick
	//the next replica...I estimate that most replica counts will
	//be typically very low, mostly less than 5, so this simple
	//algorithm will probably suffice until we support
	//multiple or plugable load balancing algorithms
	//also, this algorithm doesn't include the master as a reader
	//which someone might want

	myrand := random(0, rCnt)
	if !c.Replicas[myrand].Healthy {
		glog.V(2).Infoln("random replica was not healthy")

		//find first healthy replica
		for i := 0; i < len(c.Replicas); i++ {
			if c.Replicas[i].Healthy {
				glog.V(2).Infoln("picked replica that was healthy")
				return &c.Replicas[i], err
			}
		}

		glog.V(2).Infoln("no healthy replica found")

		if c.Master.Healthy {
			glog.V(2).Infoln("master is healthy will use instead of replica!")
			return &c.Master, err
		}

		glog.V(2).Infoln("master is unhealthy and no healthy replica found")

		return &c.Master, errors.New("master and all replicas are unhealthy")
	}

	return &c.Replicas[myrand], err
}

//give us a random number between min and less than max
func random(min, max int) int {
	rand.Seed(time.Now().UTC().UnixNano())
	return rand.Intn(max-min) + min
}

func checkError(err error) {
	if err != nil {
		glog.Fatalf("Fatal error:  %s", err.Error())
	}
}

func containsMapValues(m1 map[string]string, m2 map[string]string) bool {
	for k, v := range m1 {
		if m2[k] == v {
			glog.Fatalf("%s found in m2\n", v)
		} else {
			glog.V(2).Infof("%s not found in m2\n", v)
			return false
		}
	}
	return true
}

func UpdateHealth(node *Node, status bool) {
	var mutex = &sync.Mutex{}
	mutex.Lock()
	node.Healthy = status
	mutex.Unlock()
}
