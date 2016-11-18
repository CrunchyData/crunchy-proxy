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
	"github.com/crunchydata/crunchy-proxy/proxy/config"
	"log"
	"net"
)

func ListenAndServe(config *config.Config) {
	log.Println("[proxy] ListenAndServe config=" + config.Name)
	log.Println("[proxy] ListenAndServe listening on ipaddr=" + config.IPAddr)

	tcpAddr, err := net.ResolveTCPAddr("tcp", config.IPAddr)
	checkError(err)

	var listener net.Listener

	listener, err = net.ListenTCP("tcp", tcpAddr)
	checkError(err)

	handleListener(config, listener)
}

func handleListener(config *config.Config, listener net.Listener) {
	for {
		conn, err := listener.Accept()
		//log.Println("after Accept")
		if err != nil {
			continue
		}
		go handleClient(config, conn)
	}
}
func handleClient(cfg *config.Config, client net.Conn) {
	log.Println("[proxy] handleClient start")

	connect(cfg, client)

	defer client.Close()
	masterBuf := make([]byte, 4096)
	replicaBuf := make([]byte, 4096)
	var writeLen int
	var readLen int
	var msgType string
	var writeCase = false
	var err error
	var reqLen int
	var nextNode *config.Node
	var backendConn *net.TCPConn
	var poolIndex int

	for {

		reqLen, err = client.Read(masterBuf)
		if err != nil {
			log.Println("[proxy] error reading from client conn" + err.Error())
			return
		}

		msgType = ProtocolMsgType(masterBuf)
		LogProtocol("-->", "", masterBuf, reqLen)

		log.Println("here is a new msgType=" + msgType)

		//
		// adapt inbound data
		//err = cfg.Adapter.Do(&masterBuf, reqLen)
		//if err != nil {
		//log.Println("[proxy] error adapting inbound")
		//return
		//}

		//todo still needed?
		copy(replicaBuf, masterBuf)

		if msgType == "X" {
			log.Println("termination msg received")
			return
		} else if msgType == "Q" {
			poolIndex = -1
			writeCase = IsWriteAnno(masterBuf)
			if writeCase {
				backendConn = cfg.Master.TCPConn
				cfg.Master.Stats.Queries = cfg.Master.Stats.Queries + 1
				log.Printf("query writeCase sending to %s\n", cfg.Master.IPAddr)
				log.Println("+++++++++++incrementing writes=%d\n", cfg.Master.Stats.Queries)
			} else {
				nextNode, err = cfg.GetNextNode(writeCase)
				if err != nil {
					log.Println(err.Error())
					return
				}
				log.Println("+++++++++++incrementing reads=%d\n", nextNode.Stats.Queries)
				nextNode.Stats.Queries = nextNode.Stats.Queries + 1
				//get pool index from pool channel
				poolIndex = <-nextNode.Pool.Channel

				log.Printf("query readCase sending to %s pool Index=%d\n", nextNode.IPAddr, poolIndex)
				backendConn = nextNode.Pool.Connections[poolIndex]
			}

			writeLen, err = backendConn.Write(masterBuf[:reqLen])
			log.Printf("wrote outbuf reqLen=%d writeLen=%d\n", reqLen, writeLen)
			log.Printf("read masterBuf readLen=%d\n", readLen)
			if err != nil {
				log.Println(err.Error())
				log.Println("[proxy] error here")
			}
			readLen, err = backendConn.Read(masterBuf)
			if poolIndex != -1 {
				ReturnConnection(nextNode.Pool.Channel, poolIndex)
			}

			if err != nil {
				log.Println(err.Error())
				log.Println("attempting retry of query...")
				//right here is where retry logic occurs
				//mark as unhealthy the current node
				config.UpdateHealth(nextNode, false)

				//get next node as usual
				nextNode, err = cfg.GetNextNode(writeCase)
				if err != nil {
					log.Println("could not get node for query retry")
					log.Println(err.Error())
				} else {
					writeLen, err = nextNode.Pool.Connections[0].Write(masterBuf[:reqLen])
					readLen, err = nextNode.Pool.Connections[0].Read(masterBuf)
					if err != nil {
						log.Println("query retry failed")
						log.Println(err.Error())
					}
				}
			}

			writeLen, err = client.Write(masterBuf[:readLen])
			if err != nil {
				log.Println("[proxy] closing client conn" + err.Error())
				return
			}

			log.Printf("[proxy] wrote1 to pg client %d\n", writeLen)
		} else {

			log.Println("XXXX msgType here is " + msgType)

			writeLen, err = cfg.Master.TCPConn.Write(masterBuf[:reqLen])
			readLen, err = cfg.Master.TCPConn.Read(masterBuf)
			if err != nil {
				log.Println("master WriteRead error:" + err.Error())
			}

			msgType = ProtocolMsgType(masterBuf)

			//write to client only the master response
			writeLen, err = client.Write(masterBuf[:readLen])
			if err != nil {
				log.Println("[proxy] closing client conn" + err.Error())
				return
			}

			log.Printf("[proxy] wrote3 to pg client %d\n", writeLen)
		}

		//err = cfg.Adapter.Do(&masterBuf, readLen) //adapt the outbound msg
		//if err != nil {
		//log.Println("[proxy] error adapting outbound msg")
		//log.Println(err.Error())
		//}

	}
	log.Println("[proxy] closing client conn")
}
func checkError(err error) {
	if err != nil {
		log.Fatalf("Fatal	error:	%s", err.Error())
	}
}

/**
func RecvMessage(conn *net.TCPConn, r *[]byte) (byte, error) {
	// workaround for a QueryRow bug, see exec
	if cn.saveMessageType != 0 {
		t := cn.saveMessageType
		*r = cn.saveMessageBuffer
		cn.saveMessageType = 0
		cn.saveMessageBuffer = nil
		return t, nil
	}

	var scratch [512]byte

	x := scratch[:5]
	_, err := io.ReadFull(conn, x)
	if err != nil {
		return 0, err
	}

	// read the type and length of the message that follows
	t := x[0]
	n := int(binary.BigEndian.Uint32(x[1:])) - 4
	var y []byte
	if n <= len(scratch) {
		y = scratch[:n]
	} else {
		y = make([]byte, n)
	}
	_, err = io.ReadFull(conn, y)
	if err != nil {
		return 0, err
	}
	*r = y
	return t, nil
}
*/
