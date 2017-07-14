package pool

import (
	"net"
)

type Pool struct {
	connections chan net.Conn
	Name        string
	Capacity    int
}

func NewPool(name string, capacity int) *Pool {
	return &Pool{
		connections: make(chan net.Conn, capacity),
		Name:        name,
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

func (p *Pool) Len() int {
	return len(p.connections)
}
