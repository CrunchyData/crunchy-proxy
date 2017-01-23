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

package proxy

import (
	"bytes"
	"encoding/binary"
	"github.com/crunchydata/crunchy-proxy/config"
	"github.com/golang/glog"
	"net"
)

func ListenAndServe() {
	var listener net.Listener

	tcpAddr, err := net.ResolveTCPAddr("tcp", config.Cfg.HostPort)
	checkError(err)

	listener, err = net.ListenTCP("tcp", tcpAddr)
	checkError(err)

	glog.Infof("[proxy] Listening on: %s", config.Cfg.HostPort)

	handleListener(listener)
}

func handleListener(listener net.Listener) {
	for {
		conn, err := listener.Accept()
		if err != nil {
			continue
		}
		go handleClientConnection(conn)
	}
}

func handleClientConnection(client net.Conn) {
	glog.V(2).Infoln("[proxy] handleClient start")

	authenticated := AuthenticateClient(client)

	// If the client could not authenticate then go no further.
	if !authenticated {
		glog.Errorln("[proxy] client could not authenticate and connect.")
		return
	}

	defer client.Close()

	var err error
	masterBuf := make([]byte, 4096)
	var writeLen int
	var readLen int
	var msgType string
	var writeCase, startCase, finishCase = false, false, false
	var reqLen int
	var nextNode *config.Node
	var backendConn *net.TCPConn
	var poolIndex int
	var statementBlock = false

	for {

		reqLen, err = client.Read(masterBuf)
		if err != nil {
			glog.Errorln("[proxy] error reading from client conn" + err.Error())
			return
		}

		msgType = string(masterBuf[0])

		// adapt inbound data
		err = config.Cfg.Adapter.Do(masterBuf, reqLen)
		if err != nil {
			glog.Errorln("[proxy] error adapting inbound" + err.Error())
		}

		if msgType == "X" {
			glog.V(2).Infoln("termination msg received")
			return
		} else if msgType == "Q" {
			poolIndex = -1

			// Determine if the query has an annotation.
			writeCase, startCase, finishCase = IsWriteAnno(config.Cfg.ReadAnnotation,
				config.Cfg.StartAnnotation, config.Cfg.FinishAnnotation, masterBuf)

			glog.V(2).Infof("writeCase=%t startCase=%t finishCase=%t\n", writeCase, startCase, finishCase)

			if statementBlock {
				glog.V(2).Infof("[proxy] inside a statementBlock") // keep using the same node and connection
			} else {
				if startCase {
					statementBlock = true
				}

				nextNode, err = config.Cfg.GetNextNode(writeCase)

				if err != nil {
					glog.Errorln(err.Error())
					return
				}

				//get pool index from pool channel
				poolIndex = <-nextNode.Pool.Channel

				glog.V(2).Infof("query sending to %s pool Index=%d\n", nextNode.HostPort, poolIndex)
				backendConn = nextNode.Pool.Connections[poolIndex]
			}

			if finishCase {
				statementBlock = false
				glog.V(2).Infof("outside a statementBlock")
			}

			nextNode.Stats.Queries = nextNode.Stats.Queries + 1

			var clientErr error
			clientErr, err = processBackend(client, backendConn, masterBuf, reqLen)
			if clientErr != nil {
				glog.Errorln(clientErr.Error())
				glog.Errorln("[proxy] client write error..giving up on statement")
				return
			}

			if err != nil {
				glog.Errorln(err.Error())
				glog.Errorln("[proxy] backend write error..retrying...")
				//retry logic
				//mark node as unhealthy
				config.UpdateHealth(nextNode, false)
				//return connection index to pool
				ReturnConnection(nextNode.Pool.Channel, poolIndex)
				//get new connection on master
				nextNode, err = config.Cfg.GetNextNode(true)
				poolIndex = <-nextNode.Pool.Channel
				backendConn = nextNode.Pool.Connections[poolIndex]
				clientErr, err = processBackend(client, backendConn, masterBuf, reqLen)
				if clientErr != nil {
					glog.Errorln(clientErr.Error())
					glog.Errorln("[proxy] client write erroron retry..giving up on statement")
					return
				}
				if err != nil {
					glog.Errorln("retry of query also failed, giving up on this statement...")
					ReturnConnection(nextNode.Pool.Channel, poolIndex)
					config.UpdateHealth(nextNode, false)
					writeLen, err = client.Write(GetTerminateMessage())
					if err != nil {
						glog.Errorln(err.Error())
					}
					return
				}
			}

			if poolIndex != -1 {
				ReturnConnection(nextNode.Pool.Channel, poolIndex)
			}

		} else {

			glog.V(2).Infoln("XXXX msgType here is " + msgType)

			writeLen, err = config.Cfg.Master.TCPConn.Write(masterBuf[:reqLen])
			readLen, err = config.Cfg.Master.TCPConn.Read(masterBuf)

			if err != nil {
				glog.Errorln("master WriteRead error:" + err.Error())
			}

			msgType = string(masterBuf[0])

			//write to client only the master response
			writeLen, err = client.Write(masterBuf[:readLen])

			if err != nil {
				glog.Errorln("[proxy] closing client conn" + err.Error())
				return
			}

			glog.V(2).Infof("[proxy] wrote3 to pg client %d\n", writeLen)
		}

	}
	glog.V(2).Infoln("[proxy] closing client conn")
}
func checkError(err error) {
	if err != nil {
		glog.Fatalf("Fatal	error:	%s", err.Error())
	}
}

func getMessageTypeAndLength(buf []byte) (string, int) {
	var msgLen int32

	// Read the message length.
	reader := bytes.NewReader(buf[1:5])
	binary.Read(reader, binary.BigEndian, &msgLen)

	glog.V(2).Infof("[protocol] %d msgLen\n", msgLen)

	return string(buf[0]), int(msgLen)
}

func processBackend(client net.Conn, backendConn *net.TCPConn, masterBuf []byte, reqLen int) (error, error) {
	var writeLen, msgLen, readLen int
	var msgType string
	var err error
	var clienterr error

	writeLen, err = backendConn.Write(masterBuf[:reqLen])
	glog.V(2).Infof("wrote outbuf reqLen=%d writeLen=%d\n", reqLen, writeLen)
	if err != nil {
		glog.Errorln("[proxy] error writing to backend " + err.Error())
		return clienterr, err
	}

	for {
		readLen, err = backendConn.Read(masterBuf)
		if err != nil {
			glog.Errorln("[proxy] error reading from backend " + err.Error())
			return clienterr, err
		}
		glog.V(6).Infof("read from backend..%d\n", readLen)

		for startPos := 0; startPos < readLen; {
			msgType, msgLen = getMessageTypeAndLength(masterBuf[startPos:])

			msgLen = msgLen + 1 // add 1 for the message first byte

			//adapt msgs going back to client
			err = config.Cfg.Adapter.Do(masterBuf, readLen)

			if err != nil {
				glog.Errorln("[proxy] error adapting outbound" + err.Error())
			}

			writeLen, clienterr = client.Write(masterBuf[startPos : msgLen+startPos])
			if clienterr != nil {
				glog.Errorln("[proxy] error writing to client " + err.Error())
				return clienterr, err
			}

			glog.V(3).Infof("[proxy] wrote1 to pg client %d\n", writeLen)

			startPos = startPos + msgLen

			glog.V(3).Infof("[proxy] startPos is now %d\n", startPos)
		}
		if msgType == "Z" {
			glog.V(2).Infof("[proxy] Z msg found")
			return clienterr, err
		}
	}

	return clienterr, err
}
