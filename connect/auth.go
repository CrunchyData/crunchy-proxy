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

package connect

import (
	"bytes"
	"crypto/md5"
	"encoding/binary"
	"fmt"
	"io"
	"net"

	"github.com/crunchydata/crunchy-proxy/config"
	"github.com/crunchydata/crunchy-proxy/protocol"
	"github.com/crunchydata/crunchy-proxy/util/log"
)

/*
 * Handle authentication requests that are sent by the backend to the client.
 *
 * connection - the connection to authenticate against.
 * message - the authentication message sent by the backend.
 */
func HandleAuthenticationRequest(connection net.Conn, message []byte) bool {
	var msgLength int32
	var authType int32

	// Read message length.
	reader := bytes.NewReader(message[1:5])
	binary.Read(reader, binary.BigEndian, &msgLength)

	// Read authentication type.
	reader.Reset(message[5:9])
	binary.Read(reader, binary.BigEndian, &authType)

	switch authType {
	case protocol.AuthenticationKerberosV5:
		log.Error("KerberosV5 authentication is not currently supported.")
	case protocol.AuthenticationClearText:
		log.Info("Authenticating with clear text password.")
		return handleAuthClearText(connection)
	case protocol.AuthenticationMD5:
		log.Info("Authenticating with MD5 password.")
		return handleAuthMD5(connection, message)
	case protocol.AuthenticationSCM:
		log.Error("SCM authentication is not currently supported.")
	case protocol.AuthenticationGSS:
		log.Error("GSS authentication is not currently supported.")
	case protocol.AuthenticationGSSContinue:
		log.Error("GSS authentication is not currently supported.")
	case protocol.AuthenticationSSPI:
		log.Error("SSPI authentication is not currently supported.")
	case protocol.AuthenticationOk:
		/* Covers the case where the authentication type is 'cert' or 'trust' */
		return true
	default:
		log.Errorf("Unknown authentication method: %d", authType)
	}

	return false
}

func createMD5Password(username string, password string, salt string) string {
	// Concatenate the password and the username together.
	passwordString := fmt.Sprintf("%s%s", password, username)

	// Compute the MD5 sum of the password+username string.
	passwordString = fmt.Sprintf("%x", md5.Sum([]byte(passwordString)))

	// Compute the MD5 sum of the password hash and the salt
	passwordString = fmt.Sprintf("%s%s", passwordString, salt)
	return fmt.Sprintf("md5%x", md5.Sum([]byte(passwordString)))
}

func handleAuthMD5(connection net.Conn, message []byte) bool {
	// Get the authentication credentials.
	creds := config.GetCredentials()
	username := creds.Username
	password := creds.Password
	salt := string(message[9:13])

	password = createMD5Password(username, password, salt)

	// Create the password message.
	passwordMessage := protocol.CreatePasswordMessage(password)

	// Send the password message to the backend.
	_, err := Send(connection, passwordMessage)

	// Check that write was successful.
	if err != nil {
		log.Error("Error sending password message to the backend.")
		log.Errorf("Error: %s", err.Error())
	}

	// Read response from password message.
	message, _, err = Receive(connection)

	// Check that read was successful.
	if err != nil {
		log.Error("Error receiving authentication response from the backend.")
		log.Errorf("Error: %s", err.Error())
	}

	return protocol.IsAuthenticationOk(message)
}

func handleAuthClearText(connection net.Conn) bool {
	password := config.GetString("credentials.password")
	passwordMessage := protocol.CreatePasswordMessage(password)

	_, err := connection.Write(passwordMessage)

	if err != nil {
		log.Error("Error sending clear text password message to the backend.")
		log.Errorf("Error: %s", err.Error())
	}

	message, _, err := Receive(connection)

	//make([]byte, 4096)
	//_, err = connection.Read(response)

	if err != nil {
		log.Error("Error receiving clear text authentication response.")
		log.Errorf("Error: %s", err.Error())
	}

	return protocol.IsAuthenticationOk(message)
}

// AuthenticateClient - Establish and authenticate client connection to the backend.
//
//  This function simply handles the passing of messages from the client to the
//  backend necessary for startup/authentication of a connection. All
//  communication is between the client and the master node. If the client
//  authenticates successfully with the master node, then 'true' is returned and
//  the authenticating connection is terminated.
func AuthenticateClient(client net.Conn, message []byte, length int) (bool, error) {
	var err error

	nodes := config.GetNodes()

	node := nodes["master"]

	/* Establish a connection with the master node. */
	log.Debug("client auth: connecting to 'master' node")
	master, err := Connect(node.HostPort)

	if err != nil {
		log.Error("An error occurred connecting to the master node")
		log.Errorf("Error %s", err.Error())
		return false, err
	}

	defer master.Close()

	/* Relay the startup message to master node. */
	log.Debug("client auth: relay startup message to 'master' node")
	_, err = master.Write(message[:length])

	/* Receive startup response. */
	log.Debug("client auth: receiving startup response from 'master' node")
	message, length, err = Receive(master)

	if err != nil {
		log.Error("An error occurred receiving startup response.")
		log.Errorf("Error %s", err.Error())
		return false, err
	}

	/*
	 * While the response for the master node is not an AuthenticationOK or
	 * ErrorResponse keep relaying the mesages to/from the client/master.
	 */
	messageType := protocol.GetMessageType(message)

	for !protocol.IsAuthenticationOk(message) &&
		(messageType != protocol.ErrorMessageType) {
		Send(client, message)
		message, length, err = Receive(client)

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
			log.Info("The client closed the connection.")
			log.Debug("If the client is 'psql' and the authentication method " +
				"was 'password', then this behavior is expected.")
			return false, err
		}

		Send(master, message)

		message, length, err = Receive(master)

		messageType = protocol.GetMessageType(message)
	}

	/*
	 * If the last response from the master node was AuthenticationOK, then we
	 * need to allow the backend to finish syncing with the frontend. When that
	 * is complete, the backend will send a ReadyForQuery message. When this is
	 * received, we can then terminate the connection to the backend and return
	 * 'true' for a successful authentication of the client.
	 */
	log.Debug("client auth: checking authentication repsonse")
	if protocol.IsAuthenticationOk(message) {
		// Send the AuthenticationOK message.
		Send(client, message)

		// Continue reading from the backend and relaying the message to the
		// client.
		done := false

		for !done {
			message, _, _ = Receive(master)
			Send(client, message)
			done = protocol.GetMessageType(message) == protocol.ReadyForQueryMessageType
		}

		// Terminate the connection with the backend.
		termMsg := protocol.GetTerminateMessage()
		Send(master, termMsg)

		return true, nil
	}

	if protocol.GetMessageType(message) == protocol.ErrorMessageType {
		err = protocol.ParseError(message)
		log.Error("Error occurred on client startup.")
		log.Errorf("Error: %s", err.Error())
	} else {
		log.Error("Unknown error occurred on client startup.")
	}

	Send(client, message)

	return false, err
}

func ValidateClient(message []byte) bool {
	var clientUser string
	var clientDatabase string

	creds := config.GetCredentials()

	startup := protocol.NewMessageBuffer(message)

	startup.Seek(8) // Seek past the message length and protocol version.

	for {
		param, err := startup.ReadString()

		if err == io.EOF || param == "\x00" {
			break
		}

		switch param {
		case "user":
			clientUser, err = startup.ReadString()
		case "database":
			clientDatabase, err = startup.ReadString()
		}
	}

	return (clientUser == creds.Username && clientDatabase == creds.Database)
}
