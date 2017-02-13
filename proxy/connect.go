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
	"io"
	"net"
)

/*
 * Send a message.
 *
 * connection - the connection to which to send the message.
 * message - the message to send.
 */
func send(connection net.Conn, message []byte) error {
	_, err := connection.Write(message)

	return err
}

/*
 * Receive a message. Returns the message and the number of bytes received.
 *
 * connection - the connection from which to receive the message.
 */
func receive(connection net.Conn) ([]byte, int, error) {
	buffer := make([]byte, 4096)
	length, err := connection.Read(buffer)

	return buffer, length, err
}

func isSSLRequest(message []byte) bool {
	var messageLength int32
	var sslCode int32

	reader := bytes.NewReader(message[0:4])
	binary.Read(reader, binary.BigEndian, &messageLength)
	reader.Reset(message[4:8])
	binary.Read(reader, binary.BigEndian, &sslCode)

	return (messageLength == 8 && sslCode == SSL_REQUEST_CODE)
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
func AuthenticateClient(client net.Conn, message []byte, length int) bool {
	glog.Infoln("[proxy] Authenticating client.")
	var err error

	/* Establish a connection with the master node. */
	master, err := Connect(&config.Cfg.Master)
	defer master.Close()

	/* Relay the startup message to master node. */
	err = send(master, message[:length])

	/* Receive startup response. */
	message, length, err = receive(master)

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
			glog.V(2).Infoln("The client closed the connection.")
			glog.V(2).Infoln("If the client is 'psql' and the authentication method " +
				"was 'password', then this behavior is expected.")
			return false
		}

		send(master, message[:length])

		message, length, err = receive(master)

		messageType = getMessageType(message)
	}

	/*
	 * If the last response from the master node was AuthenticationOK, then
	 * terminate the connection and return 'true' for a successful
	 * authentication of the client.
	 */
	if isAuthenticationOk(message) {
		glog.Infoln("[proxy] Authentication successful.")
		termMsg := GetTerminateMessage()
		send(master, termMsg)
		send(client, message[:length])
		master.Close()
		return true
	} else {
		glog.Errorln("[proxy] Error occurred on client startup.")
		glog.Errorf("[proxy] Message Type: %s\n", string(getMessageType(message)))

		buffer := bytes.NewBuffer(message[5:])
		for t := buffer.Next(1)[0]; t != byte(0); t = buffer.Next(1)[0] {
			msg, _ := buffer.ReadString(byte(0))
			glog.Errorf("%s: %s", string(t), msg)
		}

		return false
	}
}
