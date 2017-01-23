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

func send(connection net.Conn, message []byte) {
	_, err := connection.Write(message)

	if err != nil {
		glog.Errorln("[proxy] Error sending message.")
		glog.Errorf("[proxy] %s\n", err.Error())
	}
}

func receive(connection net.Conn) ([]byte, int) {
	buffer := make([]byte, 4096)
	length, err := connection.Read(buffer)

	if err != nil {
		glog.Errorln("[proxy] Error receiving response.")
		glog.Errorf("[proxy] %s\n", err.Error())
	}

	glog.Infof("[proxy] Received %d bytes.", length)

	return buffer, length
}

/*
 * Establish and authenticate client connection to the backend.
 *
 * This function simply handles the passing of messages from the client to the
 * backend necessary for startup/authentication of a connection. All
 * communication is between the client and the master node. If the client
 * authenticates successfully with the master node, then 'true' is returned and
 * the authenticating connection is terminated.
 */
func AuthenticateClient(client net.Conn) bool {
	glog.Infoln("[proxy] Start client connection.")

	/* Establish a connection with the master node. */
	master, _ := net.DialTCP("tcp", nil, config.Cfg.Master.TCPAddr)

	/* Receive the startup message from client. */
	message, length := receive(client)

	/* Realy the startup message to master node. */
	send(master, message[:length])

	/* Receive startup response. */
	message, length = receive(master)

	/*
	 * While the response for the master node is not an AuthenticationOK or
	 * ErrorResponse keep relaying the mesages to/from the client/master.
	 */
	messageType := getMessageType(message)
	for !isAuthenticationOk(message) && (messageType != ERROR_MESSAGE_TYPE) {
		send(client, message[:length])
		message, length = receive(client)
		send(master, message[:length])
		message, length = receive(master)
		send(client, message[:length])
		messageType = getMessageType(message)
	}

	/*
	 * If the last response from the master node was AuthenticationOK, then
	 * terminate the connection and return 'true'.
	 *
	 * If the last response from the master is NOT AuthenticationOK, it is safe
	 * to assume that it was an ErrorResponse as all over message types are
	 * covered by the above.
	 */
	if isAuthenticationOk(message) {
		termMsg := GetTerminateMessage()
		send(master, termMsg)
		master.Close()
		return true
	} else {
		glog.Errorln("[proxy] Error occurred on client startup.")
		return false
	}
}
