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
	"crypto/md5"
	"encoding/binary"
	"fmt"
	"github.com/crunchydata/crunchy-proxy/config"
	"github.com/golang/glog"
	"net"
)

const PROTOCOL_VERSION int32 = 196608

const (
	AUTHENTICATION_OK            int32 = 0
	AUTHENTICATION_KERBEROS_V5   int32 = 2
	AUTHENTICATION_CLEAR_TEXT    int32 = 3
	AUTHENTICATION_MD5           int32 = 5
	AUTHENTICATION_SCM           int32 = 6
	AUTHENTICATION_GSS           int32 = 7
	AUTHENTICATION_GSS_CONTINUTE int32 = 8
	AUTHENTICATION_SSPI          int32 = 9
)

// Constants for the message types
const (
	AUTHENTICATION_MESSAGE_TYPE   byte = 'R'
	ERROR_MESSAGE_TYPE            byte = 'E'
	EMPTY_QUERY_MESSAGE_TYPE      byte = 'I'
	DESCRIBE_MESSAGE_TYPE         byte = 'D'
	ROW_DESCRIPTION_MESSAGE_TYPE  byte = 'T'
	DATA_ROW_MESSAGE_TYPE         byte = 'D'
	QUERY_MESSAGE_TYPE            byte = 'Q'
	COMMAND_COMPLETE_MESSAGE_TYPE byte = 'C'
	TERMINATE_MESSAGE_TYPE        byte = 'X'
	NOTICE_MESSAGE_TYPE           byte = 'N'
	PASSWORD_MESSAGE_TYPE         byte = 'p'
)

func NullTermToStrings(b []byte) (s []string) {
	var zb = []byte{0}
	for _, x := range bytes.Split(b, zb) {
		s = append(s, string(x))
	}
	if len(s) > 0 && s[len(s)-1] == "" {
		s = s[:len(s)-1]
	}
	return
}

/*
 * Handle authentication requests that are sent by the backend to the client.
 *
 * connection - the connection to authenticate against.
 * message - the authentication message sent by the backend.
 */
func handleAuthenticationRequest(connection *net.TCPConn, message []byte) bool {
	var msgLength int32
	var authType int32

	// Read message length.
	reader := bytes.NewReader(message[1:5])
	binary.Read(reader, binary.BigEndian, &msgLength)

	// Read authentication type.
	reader.Reset(message[5:9])
	binary.Read(reader, binary.BigEndian, &authType)

	switch authType {
	case AUTHENTICATION_KERBEROS_V5:
		glog.Fatalln("[protocol] KerberosV5 authentication is not currently supported.")
	case AUTHENTICATION_CLEAR_TEXT:
		glog.V(2).Infoln("[protocol] Authenticating with clear text password.")
		return handleAuthClearText(connection)
	case AUTHENTICATION_MD5:
		glog.V(2).Infoln("[protocol] Authenticating with MD5 password.")
		return handleAuthMD5(connection, message)
	case AUTHENTICATION_SCM:
		glog.Fatalln("[protocol] SCM authentication is not currently supported.")
	case AUTHENTICATION_GSS:
		glog.Fatalln("[protocol] GSS authentication is not currently supported.")
	case AUTHENTICATION_GSS_CONTINUTE:
		glog.Fatalln("[protocol] GSS authentication is not currently supported.")
	case AUTHENTICATION_SSPI:
		glog.Fatalln("[protocol] SSPI authentication is not currently supported.")
	default:
		glog.Fatalf("[protocol] Unknown authentication method: %d\n", authType)
	}

	return false
}

/*
 * Handle authentication with a clear text password. If the authentication is
 * successful, then return true otherwise false.
 *
 * connection - the connection to authenticate against.
 */
func handleAuthClearText(connection *net.TCPConn) bool {
	var writeLength, readLength int
	var response []byte
	var err error

	// Create the password message.
	passwordMessage := createPasswordMessage(config.Cfg.Credentials.Password)

	// Send the password message to the backend.
	writeLength, err = connection.Write(passwordMessage)

	// Check that write was successful.
	if err != nil {
		glog.Errorln("[protocol] Error sending password message to the backend.")
		glog.Errorf("[protocol] %s", err.Error())
	}

	glog.V(2).Infof("[protocol] %d bytes sent to the backend.\n", writeLength)

	// Read response from password message.
	readLength, err = connection.Read(response)

	// Check that read was successful.
	if err != nil {
		glog.Errorln("[protocol] Error receiving authentication response from the backend.")
		glog.Errorf("[protocol] %s\n", err.Error())
	}

	glog.V(2).Infof("[protocol] %d bytes received from the backend.\n", readLength)

	return isAuthenticationOk(response)
}

/*
 * Create a MD5 password.  The password is created using the following format:
 * md5(md5(password+username)+salt)
 *
 * username - the username of the authenticating user.
 * password - the password of the authenticating user.
 * salt - the random salt used in creation of the MD5 password.
 */
func createMD5Password(username string, password string, salt string) string {
	// Concatenate the password and the username together.
	passwordString := fmt.Sprintf("%s%s", password, username)

	// Compute the MD5 sum of the password+username string.
	passwordString = fmt.Sprintf("%x", md5.Sum([]byte(passwordString)))

	// Compute the MD5 sum of the password hash and the salt
	passwordString = fmt.Sprintf("%s%s", passwordString, salt)
	return fmt.Sprintf("md5%x", md5.Sum([]byte(passwordString)))
}

/*
 * Create a PG password message.
 *
 * password - the password to include in the payload of the message.
 */
func createPasswordMessage(password string) []byte {
	var message []byte

	// Create an MD5 password value.
	// Set the message type.
	message = append(message, PASSWORD_MESSAGE_TYPE)

	// Initialize the message length to zero.
	messageLength := make([]byte, 4)
	binary.BigEndian.PutUint32(messageLength, uint32(0))
	message = append(message, messageLength...)

	// Append the MD5 password to the message.
	message = append(message, password...)

	// null terminate the message.
	message = append(message, 0)

	/*
	 * Update the message length, subtracting the length of the message type
	 * byte.
	 */
	binary.BigEndian.PutUint32(messageLength, uint32(len(message)-1))
	copy(message[1:], messageLength)

	return message
}

/*
 * Handle authentication with a MD5 password. If the authentication is
 * successful, then return true otherwise false.
 *
 * connection - the connection to authenticate against.
 */
func handleAuthMD5(connection *net.TCPConn, message []byte) bool {
	var writeLength, readLength int
	var err error

	// Get the authentication credentials.
	username := config.Cfg.Credentials.Username
	password := config.Cfg.Credentials.Password
	salt := string(message[9:13])

	password = createMD5Password(username, password, salt)

	// Create the password message.
	passwordMessage := createPasswordMessage(password)

	// Send the password message to the backend.
	writeLength, err = connection.Write(passwordMessage)

	// Check that write was successful.
	if err != nil {
		glog.Errorln("[protocol] Error sending password message to the backend.")
		glog.Errorf("[protocol] %s", err.Error())
	}

	glog.V(2).Infof("[protocol] %d bytes sent to the backend.\n", writeLength)

	// Read response from password message.
	readLength, err = connection.Read(message)

	// Check that read was successful.
	if err != nil {
		glog.Errorln("[protocol] Error receiving authentication response from the backend.")
		glog.Errorf("[protocol] %s\n", err.Error())
	}

	glog.V(2).Infof("[protocol] %d bytes received from the backend.\n", readLength)

	return isAuthenticationOk(message)
}

/*
 * Check an Authentication Message to determine if it is an AuthenticationOK
 * message.
 *
 * It is assumed that the message passed in has already been verified to be an
 * Authentication Message.
 */
func isAuthenticationOk(message []byte) bool {
	var authenticated bool = false

	/*
	 * Determine the response message type and process accordingly. The only
	 * valid message types allowed here are Authentication and Error.
	 */
	messageType := getMessageType(message)

	if messageType == AUTHENTICATION_MESSAGE_TYPE {
		var messageValue int32

		// Get the message length.
		messageLength := getMessageLength(message)

		// Get the message value.
		reader := bytes.NewReader(message[5:9])
		binary.Read(reader, binary.BigEndian, &messageValue)

		return (messageLength == 8 && messageValue == 0)
	} else if messageType == ERROR_MESSAGE_TYPE {
		// TODO: handle error message appropriately.
	} else {
		// TODO: handle any other kind of message. This is techincally an error
		// state as well, however, not explicitly one from the backend.
	}

	return authenticated
}

/*
 * Get the message type the provided message.
 *
 * message - the message
 */
func getMessageType(message []byte) byte {
	return message[0]
}

/*
 * Get the message length of the provided message.
 *
 * message - the message
 */
func getMessageLength(message []byte) int32 {
	var messageLength int32

	reader := bytes.NewReader(message[1:5])
	binary.Read(reader, binary.BigEndian, &messageLength)

	return messageLength
}

func ErrorResponse(buf []byte) {
	var msgLen int32

	reader := bytes.NewReader(buf[1:5])
	binary.Read(reader, binary.BigEndian, &msgLen)

	glog.V(2).Infof("[protocol] ErrorResponse: msglen=%d\n", msgLen)
	var errorMessage = string(buf[5:msgLen])
	glog.V(2).Infof("[protocol] ErrorResponse: message=%s\n", errorMessage)
}

func StartupRequest(buf []byte, bufLen int) {
	var msgLen int32
	var startupProtocol int32

	reader := bytes.NewReader(buf[0:4])
	binary.Read(reader, binary.BigEndian, &msgLen)

	reader.Reset(buf[4:8])
	binary.Read(reader, binary.BigEndian, &startupProtocol)

	glog.V(2).Infof("[protocol] StartupRequest: msglen=%d protocol=%d\n", msgLen, startupProtocol)
}

func QueryRequest(buf []byte) {
	var msgLen int32
	var query string

	reader := bytes.NewReader(buf[1:5])
	binary.Read(reader, binary.BigEndian, &msgLen)

	query = string(buf[5:msgLen])

	glog.V(2).Infof("[protocol] QueryRequest: msglen=%d query=%s\n", msgLen, query)
}

func NoticeResponse(buf []byte) {
	var msgLen int32

	reader := bytes.NewReader(buf[1:5])
	binary.Read(reader, binary.BigEndian, &msgLen)

	var fieldType = buf[5]
	var fieldMsg = string(buf[6:msgLen])

	glog.V(2).Infof("[protocol] NoticeResponse: msglen=%d fieldType=%x fieldMsg=%s\n", msgLen, fieldType, fieldMsg)
}

func RowDescription(buf []byte, bufLen int) {
	var msgLen int32

	reader := bytes.NewReader(buf[1:5])
	binary.Read(reader, binary.BigEndian, &msgLen)

	glog.V(2).Infof("[protocol] RowDescription: msglen=%d\n", msgLen)
	var data []byte

	data = buf[4+msgLen : bufLen]

	var dataRowType = string(data[0])
	glog.V(2).Infof("[protocol] datarow type%s found \n", dataRowType)

}

func DataRow(buf []byte) {
	var numFields int
	var msgLen int32
	var fieldLen int32
	var fieldValue string

	reader := bytes.NewReader(buf[1:5])
	binary.Read(reader, binary.BigEndian, &msgLen)

	reader.Reset(buf[5:7])
	binary.Read(reader, binary.BigEndian, &numFields)

	reader.Reset(buf[7:11])
	binary.Read(reader, binary.BigEndian, &fieldLen)

	fieldValue = string(buf[11 : fieldLen+11])

	glog.V(2).Infof("[protocol] DataRow: numfields=%d msglen=%d fieldLen=%d fieldValue=%s\n", numFields, msgLen, fieldLen, fieldValue)
}

func CommandComplete(buf []byte) {
	var msgLen int32

	buffer := bytes.NewReader(buf[1:5])
	binary.Read(buffer, binary.BigEndian, &msgLen)

	glog.V(2).Infof("[protocol] Command Complete: msglen=%d\n", msgLen)
}

func TerminateMessage(buf []byte) {
	var msgLen int32
	buffer := bytes.NewReader(buf[1:5])
	binary.Read(buffer, binary.BigEndian, &msgLen)
	glog.V(2).Infof("[protocol] Terminate: msglen=%d\n", msgLen)
}

func GetTerminateMessage() []byte {
	var buffer []byte
	buffer = append(buffer, 'X')

	//make msg len 1 for now
	x := make([]byte, 4)
	binary.BigEndian.PutUint32(x, uint32(4))
	buffer = append(buffer, x...)
	return buffer
}

/*
 * Authenticate a connection.
 *
 * connection - the connection to authenticate against.
 */
func Authenticate(connection *net.TCPConn) bool {
	var readLength, writeLength int
	var err error
	var responseMessage []byte

	// Create the startup message.
	startupMessage := createStartupMessage()

	// Send the startup message to the backend.
	writeLength, err = connection.Write(startupMessage)

	if err == nil {
		glog.V(2).Infoln("[protocol] Startup message successfully sent to the backend.")
		glog.V(2).Infof("[protocol] %d bytes written to the backend.\n", writeLength)
	} else {
		glog.Errorf("[protocol] An error occurred sending the startup message: %s\n", err.Error())
	}

	// Receive startup reponse from the backend.
	responseMessage = make([]byte, 2048)
	readLength, err = connection.Read(responseMessage)

	if err == nil {
		glog.V(2).Infoln("[protocol] Startup response message successfully received from the backend.")
		glog.V(2).Infof("[protocol] %d bytes received from the backend.\n", readLength)
	} else {
		glog.Errorf("[protocol] An error occurred receiving the startup response: %s\n", err.Error())
	}

	/*
	 * Check the message type.
	 *
	 * The first byte of the message is always the message type. The received
	 * message should be an authentication type which has a value of 'R'.
	 */
	if responseMessage[0] != AUTHENTICATION_MESSAGE_TYPE {
		glog.Errorln("[protocol] Received incorrect message type: should receive a authentication message type (%s).\n",
			string(AUTHENTICATION_MESSAGE_TYPE))
	}

	// Handle authentication request.
	authenticated := handleAuthenticationRequest(connection, responseMessage)

	if !authenticated {
		glog.Errorln("[protocol] Authentication failed.")
	}

	return authenticated
}

/*
 * Create a PG startup message. This message is used to startup all connections
 * with a PG backend.
 */
func createStartupMessage() []byte {

	//send startup packet
	var buffer []byte

	x := make([]byte, 4)

	// Temporarily set the message length to 0.
	binary.BigEndian.PutUint32(x, uint32(1))
	buffer = append(buffer, x...)

	// Set the protocol version.
	binary.BigEndian.PutUint32(x, uint32(PROTOCOL_VERSION))
	buffer = append(buffer, x...)

	/*
	 * The protocol version number is followed by one or more pairs of
	 * parameter name and value strings. A zero byte is required as a
	 * terminator after the last name/value pair. Parameters can appear in any
	 * order. 'user' is required, others are optional.
	 */
	var key, value string

	/*
	 * Set the 'user' parameter.  This is the only *required* parameter.
	 */
	key = "user"
	buffer = append(buffer, key...)
	buffer = append(buffer, 0)

	value = config.Cfg.Credentials.Username
	buffer = append(buffer, value...)
	buffer = append(buffer, 0)

	/*
	 * Set the 'database' parameter.  If no database name has been specified,
	 * then the default value is the user's name.
	 *
	 * TODO: Determine if the default should be handled here or if it assumed
	 * by the backend.
	 */
	key = "database"
	buffer = append(buffer, key...)
	buffer = append(buffer, 0)

	value = config.Cfg.Credentials.Database
	buffer = append(buffer, value...)
	buffer = append(buffer, 0)

	/*
	 * Set the 'client_encoding' parameter.
	 *
	 * TODO: Add this as a configuration specific item.
	 */
	key = "client_encoding"
	buffer = append(buffer, key...)
	buffer = append(buffer, 0)

	value = "UTF8"
	buffer = append(buffer, value...)
	buffer = append(buffer, 0)

	/*
	 * Set the 'datestyle' parameter.
	 */
	key = "datestyle"
	buffer = append(buffer, key...)
	buffer = append(buffer, 0)

	value = "ISO, MDY"
	buffer = append(buffer, value...)
	buffer = append(buffer, 0)

	/*
	 * Set the 'application_name' parameter.
	 */
	key = "application_name"
	buffer = append(buffer, key...)
	buffer = append(buffer, 0)

	value = "proxypool"
	buffer = append(buffer, value...)
	buffer = append(buffer, 0)

	/*
	 * Set the 'extra_float_digits' parameter.
	 *
	 * TODO: Determine why this parameter is necessary.
	 */
	key = "extra_float_digits"
	buffer = append(buffer, key...)
	buffer = append(buffer, 0)

	value = "2"
	buffer = append(buffer, value...)
	buffer = append(buffer, 0)

	// TODO: Determine if this last null byte is necessary.
	buffer = append(buffer, 0)

	// update the msg len
	binary.BigEndian.PutUint32(buffer, uint32(len(buffer)))

	return buffer
}
