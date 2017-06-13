package pool

import (
	"net"
)

type Pool struct {
	connections chan net.Conn
	Capacity    int
}

func NewPool(capacity int) *Pool {
	return &Pool{
		connections: make(chan net.Conn, capacity),
		Capacity:    capacity,
	}
}

func (p *Pool) Add(connection net.Conn) {
	p.connections <- connection
}

func (p *Pool) Next() net.Conn {
	return <-p.connections
}

func (p *Pool) Return(connection net.Conn) {
	p.connections <- connection
}
