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
	"encoding/binary"
	"github.com/crunchydata/crunchy-proxy/config"
	"github.com/golang/glog"
	"net"
)

func connect(client net.Conn) error {
	glog.V(2).Infoln("[proxy] connect start")

	config.Cfg.GetAllConnections()
	glog.V(2).Infoln("replicas cnt=%d\n", len(config.Cfg.Replicas))

	masterBuf := make([]byte, 4096)
	var masterReadLen int
	var msgType string
	var msgLen int
	var err error
	var clientLen int

	clientLen, err = client.Read(masterBuf)
	if err != nil {
		glog.Errorln("[proxy] error reading from client conn" + err.Error())
		return err
	}
	thelen := binary.BigEndian.Uint32(masterBuf[:4])
	theprotocol := binary.BigEndian.Uint32(masterBuf[4:8])
	glog.V(2).Infoln("the len=%d the protocol=%d\n", thelen, theprotocol)
	glog.V(2).Infoln("the msg=[%s] \n", string(masterBuf[8:]))

	LogProtocol("-->", "startup", masterBuf, clientLen)

	masterReadLen, err = config.Cfg.Master.TCPConn.Write(masterBuf[:clientLen])
	masterReadLen, err = config.Cfg.Master.TCPConn.Read(masterBuf)

	if err != nil {
		glog.Errorln("master WriteRead error:" + err.Error())
	}

	LogProtocol("<--", "master", masterBuf, masterReadLen)
	//here you will get an R authrequest or an N notice from the master

	msgType, _ = ProtocolMsgType(masterBuf)
	if msgType == "N" {
		//write to client only the master "N" response
		msgType, _ = ProtocolMsgType(masterBuf)
		glog.V(2).Infoln("sending N to client")
		_, err = client.Write(masterBuf[:masterReadLen])
		if err != nil {
			glog.Errorln("[proxy] closing client conn" + err.Error())
			return err
		}
		glog.V(2).Infoln("read client response after N was sent")
		clientLen, err = client.Read(masterBuf)
		if err != nil {
			glog.Errorln("[proxy] error reading from client conn" + err.Error())
			return err
		}

		//write client response to master
		masterReadLen, err = config.Cfg.Master.TCPConn.Write(masterBuf[:clientLen])
		masterReadLen, err = config.Cfg.Master.TCPConn.Read(masterBuf)
		if err != nil {
			glog.Errorln("master WriteRead error:" + err.Error())
		}
		msgType, _ = ProtocolMsgType(masterBuf)

		glog.V(2).Infoln("read from master after N msgType=" + msgType)

	}

	//
	//send R authentication request to client
	//
	msgType, msgLen = ProtocolMsgType(masterBuf)
	glog.V(2).Infoln("sending msgType to client: %s msgLen=%d\n",
		msgType, msgLen)
	glog.V(2).Infoln("should be R Auth msg here send R auth to client")
	_, err = client.Write(masterBuf[:masterReadLen])
	if err != nil {
		glog.Errorln("[proxy] closing client conn" + err.Error())
		return err
	}

	glog.V(2).Infoln("read client response after R was sent")
	clientLen, err = client.Read(masterBuf)
	if err != nil {
		glog.Errorln("[proxy] error reading from client conn" + err.Error())
		return err
	}

	LogProtocol("-->", "master", masterBuf, clientLen)

	//
	//process the 'p' password message from the client
	//
	masterReadLen, err = config.Cfg.Master.TCPConn.Write(masterBuf[:clientLen])
	masterReadLen, err = config.Cfg.Master.TCPConn.Read(masterBuf)
	if err != nil {
		glog.Errorln("master WriteRead error:" + err.Error())
	}
	msgType, msgLen = ProtocolMsgType(masterBuf)
	glog.V(2).Infoln("pt 5 got msgType " + msgType)
	if msgType == "R" {
		LogProtocol("<--", "AuthenticationOK", masterBuf, masterReadLen)
	}

	//write to client only the master response
	_, err = client.Write(masterBuf[:masterReadLen])
	if err != nil {
		glog.Errorln("[proxy] closing client conn" + err.Error())
		return err
	}

	glog.V(2).Infoln("end of connect logic..sending terminate")

	//after authenticating to the master, we terminate this connection
	//will use pool connections for the rest of the user session
	termMsg := GetTerminateMessage()
	masterReadLen, err = config.Cfg.Master.TCPConn.Write(termMsg)
	if err != nil {
		glog.Errorln("master WriteRead error on term msg:" + err.Error())
	}

	return nil

}
