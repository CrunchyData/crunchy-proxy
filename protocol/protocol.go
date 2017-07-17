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

package protocol

import (
	"bytes"
	"encoding/binary"
)

/* PostgreSQL Protocol Version/Code constants */
const (
	ProtocolVersion int32 = 196608
	SSLRequestCode  int32 = 80877103

	/* SSL Responses */
	SSLAllowed    byte = 'S'
	SSLNotAllowed byte = 'N'
)

/* PostgreSQL Message Type constants. */
const (
	AuthenticationMessageType  byte = 'R'
	ErrorMessageType           byte = 'E'
	EmptyQueryMessageType      byte = 'I'
	DescribeMessageType        byte = 'D'
	RowDescriptionMessageType  byte = 'T'
	DataRowMessageType         byte = 'D'
	QueryMessageType           byte = 'Q'
	CommandCompleteMessageType byte = 'C'
	TerminateMessageType       byte = 'X'
	NoticeMessageType          byte = 'N'
	PasswordMessageType        byte = 'p'
	ReadyForQueryMessageType   byte = 'Z'
)

/* PostgreSQL Authentication Method constants. */
const (
	AuthenticationOk          int32 = 0
	AuthenticationKerberosV5  int32 = 2
	AuthenticationClearText   int32 = 3
	AuthenticationMD5         int32 = 5
	AuthenticationSCM         int32 = 6
	AuthenticationGSS         int32 = 7
	AuthenticationGSSContinue int32 = 8
	AuthenticationSSPI        int32 = 9
)

func GetVersion(message []byte) int32 {
	var code int32

	reader := bytes.NewReader(message[4:8])
	binary.Read(reader, binary.BigEndian, &code)

	return code
}

/*
 * Get the message type the provided message.
 *
 * message - the message
 */
func GetMessageType(message []byte) byte {
	return message[0]
}

/*
 * Get the message length of the provided message.
 *
 * message - the message
 */
func GetMessageLength(message []byte) int32 {
	var messageLength int32

	reader := bytes.NewReader(message[1:5])
	binary.Read(reader, binary.BigEndian, &messageLength)

	return messageLength
}

/* IsAuthenticationOk
 *
 * Check an Authentication Message to determine if it is an AuthenticationOK
 * message.
 */
func IsAuthenticationOk(message []byte) bool {
	/*
	 * If the message type is not an Authentication message, then short circuit
	 * and return false.
	 */
	if GetMessageType(message) != AuthenticationMessageType {
		return false
	}

	var messageValue int32

	// Get the message length.
	messageLength := GetMessageLength(message)

	// Get the message value.
	reader := bytes.NewReader(message[5:9])
	binary.Read(reader, binary.BigEndian, &messageValue)

	return (messageLength == 8 && messageValue == AuthenticationOk)
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
