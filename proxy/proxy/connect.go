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
	"github.com/crunchydata/crunchy-proxy/proxy/config"
	"log"
	"net"
)

func connect(cfg *config.Config, client net.Conn) {
	log.Println("[proxy] connect start")

	cfg.GetAllConnections()
	log.Printf("replicas cnt=%d\n", len(cfg.Replicas))

	masterBuf := make([]byte, 4096)
	replicaBuf := make([]byte, 4096)
	var masterReadLen int
	var msgType string
	var err error
	var clientLen int

	clientLen, err = client.Read(masterBuf)
	if err != nil {
		log.Println("[proxy] error reading from client conn" + err.Error())
		return
	}
	copy(replicaBuf, masterBuf)
	thelen := binary.BigEndian.Uint32(masterBuf[:4])
	theprotocol := binary.BigEndian.Uint32(masterBuf[4:8])
	log.Printf("the len=%d the protocol=%d\n", thelen, theprotocol)
	log.Printf("the msg=[%s] \n", string(masterBuf[8:]))

	LogProtocol("-->", "startup", masterBuf, clientLen)

	masterReadLen, err = cfg.Master.TCPConn.Write(masterBuf[:clientLen])
	masterReadLen, err = cfg.Master.TCPConn.Read(masterBuf)

	if err != nil {
		log.Println("master WriteRead error:" + err.Error())
	}

	LogProtocol("<--", "master", masterBuf, masterReadLen)
	//here you will get an R authrequest or an N notice from the master

	if msgType == "N" {
		//write to client only the master "N" response
		msgType = ProtocolMsgType(masterBuf)
		log.Println("sending N to client")
		_, err = client.Write(masterBuf[:masterReadLen])
		if err != nil {
			log.Println("[proxy] closing client conn" + err.Error())
			return
		}

		//read the next message from the master
		masterReadLen, err = cfg.Master.TCPConn.Read(masterBuf)
		msgType = ProtocolMsgType(masterBuf)
		log.Println("read from master after N msgType=" + msgType)
		copy(replicaBuf, masterBuf)
	}

	//
	//send R authentication request to client
	//
	msgType = ProtocolMsgType(masterBuf)
	log.Println("sending msgType to client:" + msgType)
	log.Println("should be R Auth msg here send R auth to client")
	_, err = client.Write(masterBuf[:masterReadLen])
	if err != nil {
		log.Println("[proxy] closing client conn" + err.Error())
		return
	}

	log.Println("read client response after R was sent")
	clientLen, err = client.Read(masterBuf)
	if err != nil {
		log.Println("[proxy] error reading from client conn" + err.Error())
		return
	}

	LogProtocol("-->", "master", masterBuf, clientLen)

	//
	//process the 'p' password message from the client
	//
	masterReadLen, err = cfg.Master.TCPConn.Write(masterBuf[:clientLen])
	masterReadLen, err = cfg.Master.TCPConn.Read(masterBuf)
	if err != nil {
		log.Println("master WriteRead error:" + err.Error())
	}
	msgType = ProtocolMsgType(masterBuf)
	log.Println("pt 5 got msgType " + msgType)
	if msgType == "R" {
		LogProtocol("<--", "AuthenticationOK", masterBuf, masterReadLen)
	}

	//write to client only the master response
	_, err = client.Write(masterBuf[:masterReadLen])
	if err != nil {
		log.Println("[proxy] closing client conn" + err.Error())
		return
	}

	log.Println("end of connect logic")

}
