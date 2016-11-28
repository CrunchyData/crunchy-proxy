package config

import (
	"encoding/json"
	"errors"
	"flag"
	"github.com/crunchydata/crunchy-proxy/adapter"
	"io/ioutil"
	"log"
	"math/rand"
	"net"
	"os"
	"sync"
	"time"
)

type NodeStats struct {
	Queries int `json:"-"`
}

type NodePool struct {
	Channel     chan int       `json:"-"`
	Connections []*net.TCPConn `json:"-"`
}

type PoolConfig struct {
	Enabled  bool `json:"enabled"`
	Capacity int  `json:"capacity"`
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
	IPAddr       string            `json:"ipaddr"` //remote host:port
	Metadata     map[string]string `json:"metadata"`
	Healthy      bool              `json:"-"`
	HCConnection net.Conn          `json:"-"`
	TCPAddr      *net.TCPAddr      `json:"-"`
	TCPConn      *net.TCPConn      `json:"-"`
	Pool         NodePool          `json:"-"`
	Stats        NodeStats         `json:"-"`
}

type Config struct {
	Name        string          `json:"name"`
	IPAddr      string          `json:"ipaddr"`      //listen on host:port
	AdminIPAddr string          `json:"adminipaddr"` //listen on host:port
	Credentials PGCredentials   `json:"credentials"`
	Pool        PoolConfig      `json:"pool"`
	Master      Node            `json:"master"`
	Replicas    []Node          `json:"replicas"`
	Adapters    []string        `json:"adapters"`
	Healthcheck Healthcheck     `json:"healthcheck"`
	Adapter     adapter.Adapter `json:"-"`
}

func (c Config) Print() {
	str, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		log.Println(err)
	}
	log.Println(string(str))

}
func (c Config) PrintNodeInfo(msg string) {
	log.Printf("----Master Info %s----\n", msg)
	log.Printf("master=%s ", c.Master.IPAddr)
	log.Printf("----Replica Info %s----\n", msg)
	for i := 0; i < len(c.Replicas); i++ {
		log.Printf("replica=%s ", c.Replicas[i].IPAddr)
	}
}

func PrintExample() {
	ds := []string{"logging", "audit"}
	var pool = PoolConfig{
		Enabled:  true,
		Capacity: 2}
	cred := PGCredentials{
		Username: "logging",
		Password: "audit",
		Database: "database1"}

	var ms = Node{
		IPAddr: "master:5432"}

	ms.Metadata = make(map[string]string)

	var rs = make([]Node, 2)
	rs[0] = Node{
		IPAddr: "replica1:5432"}
	rs[0].Metadata = make(map[string]string)
	rs[1] = Node{
		IPAddr: "replica2:5432"}
	rs[1].Metadata = make(map[string]string)
	var hs Healthcheck
	hs.Delay = 10
	hs.Query = "select now()"

	c := Config{
		Name:        "sampleconfig",
		IPAddr:      "localhost:5432",
		Master:      ms,
		Pool:        pool,
		Credentials: cred,
		Replicas:    rs,
		Healthcheck: hs,
		Adapters:    ds}

	str, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		log.Println(err)
	}
	log.Println(string(str))
}

func ReadConfig() Config {

	var filePath string
	flag.StringVar(&filePath, "config", "", "a configuration file")
	flag.Parse()
	log.Println("[config]" + filePath + " is the config path")
	if filePath == "" {
		log.Println("-config command option required")
		os.Exit(1)
	}

	var cfg Config
	var err error
	var byt []byte

	if byt, err = ioutil.ReadFile(filePath); err != nil {
		panic(err)
	}
	if err = json.Unmarshal(byt, &cfg); err != nil {
		panic(err)
	}

	cfg.Master.TCPAddr, err = net.ResolveTCPAddr("tcp4", cfg.Master.IPAddr)
	checkError(err)
	for n := 0; n < len(cfg.Replicas); n++ {
		cfg.Replicas[n].TCPAddr, err = net.ResolveTCPAddr("tcp4", cfg.Replicas[n].IPAddr)
		checkError(err)
	}

	return cfg
}

func (n *Node) GetConnection() (*net.TCPConn, error) {
	conn, err := net.DialTCP("tcp", nil, n.TCPAddr)
	return conn, err
}

func (c *Config) GetAllConnections() {

	var err error
	log.Println("dialing " + c.Master.IPAddr)
	c.Master.TCPConn, err = net.DialTCP("tcp", nil, c.Master.TCPAddr)
	if err != nil {
		log.Println(err.Error())
	}

}

func (c *Config) SetupAdapters() {
	var ds []adapter.Decorator = make([]adapter.Decorator, 0)
	for i := 0; i < len(c.Adapters); i++ {

		switch c.Adapters[i] {
		case "audit":
			ds = append(ds, adapter.Audit(log.New(os.Stdout, "[audit adapter]", 0)))
		case "logging":
			ds = append(ds, adapter.Logging(log.New(os.Stdout, "[log adapter]", 0)))
		default:
			log.Println("config found invalid adapter:" + c.Adapters[i])
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
			log.Println("master is unhealthy!")
			return &c.Master, errors.New("unhealthy master")
		}
		log.Println("writeCase so using master as node...")
		return &c.Master, err
	}

	var replicaHealthy = false

	for i := 0; i < len(c.Replicas); i++ {
		if c.Replicas[i].Healthy {
			log.Println("picked replica that was healthy")
			replicaHealthy = true
		}
	}

	if rCnt == 1 && replicaHealthy == false {
		log.Println("no replicas are healthy..using master")
		if !c.Master.Healthy {
			log.Println("master is unhealthy!")
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
		log.Println("random replica was not healthy")
		//find first healthy replica
		for i := 0; i < len(c.Replicas); i++ {
			if c.Replicas[i].Healthy {
				log.Println("picked replica that was healthy")
				return &c.Replicas[i], err
			}
		}

		log.Println("no healthy replica found")
		if c.Master.Healthy {
			log.Println("master is healthy will use instead of replica!")
			return &c.Master, err
		}
		log.Println("master is unhealthy and no healthy replica found")
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
		log.Fatalf("Fatal   error:  %s", err.Error())
	}
}

func containsMapValues(m1 map[string]string, m2 map[string]string) bool {
	for k, v := range m1 {
		if m2[k] == v {
			log.Printf("%s found in m2\n", v)
		} else {
			log.Printf("%s not found in m2\n", v)
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
