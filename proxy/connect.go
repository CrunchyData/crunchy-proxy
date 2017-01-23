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
	"io"
	"net"
)

/*
 * Send a message.
 *
 * connection - the connection to which to send the message.
 * message - the message to send.
 */
func send(connection net.Conn, message []byte) {
	_, err := connection.Write(message)

	if err != nil {
		glog.Errorln("[proxy] Error sending message.")
		glog.Errorf("[proxy] %s\n", err.Error())
	}
}

/*
 * Receive a message. Returns the message and the number of bytes received.
 *
 * connection - the connection from which to receive the message.
 */
func receive(connection net.Conn) ([]byte, int, error) {
	buffer := make([]byte, 4096)
	length, err := connection.Read(buffer)

	if err != nil {
		glog.Errorln("[proxy] Error receiving response.")
		glog.Errorf("[proxy] %s\n", err.Error())
	}

	return buffer, length, err
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
	var err error

	/* Establish a connection with the master node. */
	master, _ := net.DialTCP("tcp", nil, config.Cfg.Master.TCPAddr)

	/* Receive the startup message from client. */
	message, length, _ := receive(client)

	/* Realy the startup message to master node. */
	send(master, message[:length])

	/* Receive startup response. */
	message, length, _ = receive(master)

	/*
	 * While the response for the master node is not an AuthenticationOK or
	 * ErrorResponse keep relaying the mesages to/from the client/master.
	 */
	messageType := getMessageType(message)
	for !isAuthenticationOk(message) && (messageType != ERROR_MESSAGE_TYPE) {
		send(client, message[:length])
		message, length, err = receive(client)

		/*
		 * Must check that the client has not closed the connection.  This in
		 * particular is specific to 'psql' when it prompts for a password.
		 * Apparently, when psql prompts the user for a password it closes the
		 * original connection, and then creates a new one. Eventually the
		 * following send/receives would timeout and no 'meaningful' messages
		 * are relayed. This would ultimately cause an infinite loop.  Thus it
		 * is better to short circuit here if the client connection has been
		 * closed.
		 */
		if (err != nil) && (err == io.EOF) {
			glog.Infoln("The client closed the connection.")
			glog.Infoln("If the client is 'psql' and the authentication method " +
				"was 'password', then this behavior is expected.")
			return false
		}

		send(master, message[:length])
		message, length, _ = receive(master)

		send(client, message[:length])
		messageType = getMessageType(message)
	}

	/*
	 * If the last response from the master node was AuthenticationOK, then
	 * terminate the connection and return 'true' for a successful
	 * authentication of the client.
	 *
	 * If the last response from the master is NOT AuthenticationOK, it is safe
	 * to assume that it was an ErrorResponse as all over message types are
	 * covered by the above.  As well
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
