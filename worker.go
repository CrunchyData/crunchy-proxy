package main

import (
	"log"
	"net"
	"time"
)

type ConnectionPool struct {
	Name        string
	Channel     chan int
	Connections []string
}

var pool1, pool2 ConnectionPool

func SetupPools(name string) ConnectionPool {
	cp := ConnectionPool{}
	cp.Channel = make(chan int, 2)
	log.Println("setupPools")
	cp.Name = "pool1"
	cp.Connections = []string{name + "-a", name + "-b"}
	for i := 0; i < len(cp.Connections); i++ {
		cp.Channel <- i
	}
	log.Println("setupPools done")
	return cp
}

func returnTheMerch(ch chan int, merch int) {
	log.Printf("returning the pool connection %d\n", merch)
	ch <- merch
}

func handler(c net.Conn) {
	var poolChannel chan int
	poolChannel = pool1.Channel
	poolIndex := <-poolChannel
	defer returnTheMerch(poolChannel, poolIndex)
	log.Printf("handler called with poolIndex=%d\n", poolIndex)
	time.Sleep(8000 * time.Millisecond)
	c.Write([]byte("ok"))
	c.Close()
}

func server(l net.Listener) {
	for {
		c, err := l.Accept()
		if err != nil {
			continue
		}
		go handler(c)
	}
}

func main() {
	l, err := net.Listen("tcp", ":5000")
	if err != nil {
		panic(err)
	}

	pool1 = SetupPools("pool1")
	pool2 = SetupPools("pool2")

	go server(l)
	for {
		log.Println("server main loop sleeping")
		time.Sleep(10 * time.Second)
	}
}
