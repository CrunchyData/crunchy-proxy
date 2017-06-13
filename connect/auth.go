package connect

import (
	"bytes"
	"encoding/binary"
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
func handleAuthenticationRequest(connection net.Conn, message []byte) bool {
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

func handleAuthMD5(connection net.Conn, message []byte) bool {
	return true
}

func handleAuthClearText(connection net.Conn) bool {
	password := config.GetString("credentials.password")
	passwordMessage := protocol.CreatePasswordMessage(password)

	_, err := connection.Write(passwordMessage.Bytes())

	if err != nil {
		log.Error("Error sending clear text password message to the backend.")
		log.Errorf("Error: %s", err.Error())
	}

	response := make([]byte, 4096)
	_, err = connection.Read(response)

	if err != nil {
		log.Error("Error receiving clear text authentication response.")
		log.Errorf("Error: %s", err.Error())
	}

	return protocol.IsAuthenticationOk(response)
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

	/*
	 * Validate that the client username and database are the same as that
	 * which is configured for the proxy connections.
	 *
	 * If the the client cannot be validated then send an appropriate PG error
	 * message back to the client.
	 */

	nodes := config.GetNodes()

	node := nodes["master"]

	/* Establish a connection with the master node. */
	master, err := net.Dial("tcp", node.HostPort)

	if err != nil {
		log.Error("An error occurred connecting to the master node")
		log.Errorf("Error %s", err.Error())
		return false, err
	}

	defer master.Close()

	/* Relay the startup message to master node. */
	_, err = master.Write(message[:length])

	/* Receive startup response. */
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
		Send(client, message[:length])
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

		Send(master, message[:length])

		message, length, err = Receive(master)

		messageType = protocol.GetMessageType(message)
	}

	/*
	 * If the last response from the master node was AuthenticationOK, then
	 * terminate the connection and return 'true' for a successful
	 * authentication of the client.
	 */
	if protocol.IsAuthenticationOk(message) {
		termMsg := protocol.GetTerminateMessage()
		Send(master, termMsg)
		Send(client, message[:length])
		return true, nil
	}

	if protocol.GetMessageType(message) == protocol.ErrorMessageType {
		err = protocol.ParseError(message)
		log.Error("Error occurred on client startup.")
		log.Errorf("Error: %s", err.Error())
		log.Debugf("%s", message[:length])
	} else {
		log.Error("Unknown error occurred on client startup.")
	}

	Send(client, message[:length])

	return false, err
}

func ValidateClient(message []byte) bool {
	var clientUser string
	var clientDatabase string

	creds := config.GetCredentials()

	startup := protocol.NewMessageBuffer(message)

	startup.Seek(8) // Seek past the message length and protocol version.

	for param, err := startup.ReadString(); err != io.EOF; param, err = startup.ReadString() {
		switch param {
		case "user":
			clientUser, _ = startup.ReadString()
		case "database":
			clientDatabase, _ = startup.ReadString()
		}
	}

	return (clientUser == creds.Username && clientDatabase == creds.Database)
}
