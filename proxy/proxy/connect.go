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

	//write to client only the master "N" response
	_, err = client.Write(masterBuf[:masterReadLen])
	if err != nil {
		log.Println("[proxy] closing client conn" + err.Error())
		return
	}

	//read the startup part 2 message from the client
	clientLen, err = client.Read(masterBuf)
	if err != nil {
		log.Println("[proxy] error reading from client conn" + err.Error())
		return
	}
	copy(replicaBuf, masterBuf)

	msgType = ProtocolMsgType(masterBuf)
	//log.Println("msgType here is " + msgType)
	LogProtocol("-->", msgType, masterBuf, clientLen)

	masterReadLen, err = cfg.Master.TCPConn.Write(masterBuf[:clientLen])
	masterReadLen, err = cfg.Master.TCPConn.Read(masterBuf)
	//masterReadLen, err = WriteRead("master", cfg.Master.TCPConn, clientLen, masterBuf)
	if err != nil {
		log.Println("master WriteRead error:" + err.Error())
	}
	/**
	_, err = WriteRead("replica0", cfg.Replicas[0].TCPConn, clientLen, replicaBuf)
	if err != nil {
		log.Println("replica WriteRead error: " + err.Error())
	}
	replicaSalt = AuthenticationRequest(replicaBuf)
	log.Printf("picked out replicaSalt of %x\n", replicaSalt)
	*/

	//msgType = ProtocolMsgType(masterBuf)
	//log.Println("pt 3 got here msgType=" + msgType)
	LogProtocol("<--", "master", masterBuf, masterReadLen)

	//
	//send R authentication request to client
	//
	_, err = client.Write(masterBuf[:masterReadLen])
	if err != nil {
		log.Println("[proxy] closing client conn" + err.Error())
		return
	}

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
