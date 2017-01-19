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
		go handleClient(conn)
	}
}

func handleClient(client net.Conn) {
	glog.V(2).Infoln("[proxy] handleClient start")

	/*
	 * TODO: handle client connection differently, perhaps refactor so that
	 * pool connections and client connections follow the same code path.
	 */
	err := connect(client)

	if err != nil {
		glog.Errorln("[proxy] client could not authenticate and connect")
		return
	}

	defer client.Close()

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
			writeCase, startCase, finishCase = IsWriteAnno(config.Cfg.ReadAnnotation,
				config.Cfg.StartAnnotation, config.Cfg.FinishAnnotation, masterBuf)
			glog.V(2).Infof("writeCase=%t startCase=%t finishCase=%t\n", writeCase, startCase, finishCase)
			if statementBlock {
				glog.V(2).Infof("inside a statementBlock") //keep using the same node and connection
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

			writeLen, err = backendConn.Write(masterBuf[:reqLen])
			glog.V(2).Infof("wrote outbuf reqLen=%d writeLen=%d\n", reqLen, writeLen)
			if err != nil {
				glog.Errorln(err.Error())
				glog.Errorln("[proxy] error here")
			}

			//write the query to backend then read and write
			//till we get Q from the backend
			err = processBackend(client, backendConn, masterBuf)

			if poolIndex != -1 {
				ReturnConnection(nextNode.Pool.Channel, poolIndex)
			}

			if err != nil {
				glog.Errorln(err.Error())
				glog.Errorln("attempting retry of query...")

				//right here is where retry logic occurs
				//mark as unhealthy the current node
				config.UpdateHealth(nextNode, false)

				//get next node as usual
				nextNode, err = config.Cfg.GetNextNode(writeCase)

				if err != nil {
					glog.Errorln("could not get node for query retry")
					glog.Errorln(err.Error())
				} else {
					writeLen, err = nextNode.Pool.Connections[0].Write(masterBuf[:reqLen])
					readLen, err = nextNode.Pool.Connections[0].Read(masterBuf)
					if err != nil {
						glog.Errorln("query retry failed")
						glog.Errorln(err.Error())
					}
				}
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

func processBackend(client net.Conn, backendConn *net.TCPConn, masterBuf []byte) error {
	var writeLen, msgLen, readLen int
	var msgType string
	var err error

	for {
		readLen, err = backendConn.Read(masterBuf)
		for startPos := 0; startPos < readLen; {
			msgType = string(masterBuf[0])

			msgLen = msgLen + 1 //add 1 for the message first byte
			//adapt msgs going back to client
			err = config.Cfg.Adapter.Do(masterBuf, readLen)
			if err != nil {
				glog.Errorln("[proxy] error adapting outbound" + err.Error())
			}

			writeLen, err = client.Write(masterBuf[startPos : msgLen+startPos])
			glog.V(3).Infof("[proxy] wrote1 to pg client %d\n", writeLen)
			startPos = startPos + msgLen
			glog.V(3).Infof("[proxy] startPos is now %d\n", startPos)
		}
		if msgType == "Z" {
			glog.V(2).Infof("[proxy] Z msg found")
			return err
		}
	}

	return err
}
