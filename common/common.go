package common

import (
	"net"
)

const (
	NODE_ROLE_MASTER  string = "master"
	NODE_ROLE_REPLICA string = "replica"
)

type Node struct {
	HostPort string            `mapstructure:"hostport"` //remote host:port
	Role     string            `mapstructure:"role"`
	Metadata map[string]string `mapstructure:"metadata"`
	Healthy  bool              `mapstructure:"-"`
}

type Pool struct {
	Channel     chan int   `mapstructure:"-"`
	Connections []net.Conn `mapstructure:"-"`
}

type SSLConfig struct {
	Enable        bool   `mapstructure:"enable"`
	SSLMode       string `mapstructure:"sslmode"`
	SSLCert       string `mapstructure:"sslcert,omitempty"`
	SSLKey        string `mapstructure:"sslkey,omitempty"`
	SSLRootCA     string `mapstructure:"sslrootca,omitempty"`
	SSLServerCert string `mapstructure:"sslservercert,omitempty"`
	SSLServerKey  string `mapstructure:"sslserverkey,omitempty"`
	SSLServerCA   string `mapstructure:"sslserverca,omitempty"`
}

type Credentials struct {
	Username string            `mapstructure:"username"`
	Password string            `mapstructure:"password,omitempty"`
	Database string            `mapstructure:"database"`
	SSL      SSLConfig         `mapstructure:"ssl"`
	Options  map[string]string `mapstructure:"options"`
}

type HealthCheckConfig struct {
	Delay int    `mapstructure:"delay"`
	Query string `mapstructure:"query"`
}
