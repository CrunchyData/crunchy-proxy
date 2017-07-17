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
