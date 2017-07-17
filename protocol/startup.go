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

// CreateStartupMessage creates a PG startup message. This message is used to
// startup all connections with a PG backend.
func CreateStartupMessage(username string, database string, options map[string]string) []byte {
	message := NewMessageBuffer([]byte{})

	/* Temporarily set the message length to 0. */
	message.WriteInt32(0)

	/* Set the protocol version. */
	message.WriteInt32(ProtocolVersion)

	/*
	 * The protocol version number is followed by one or more pairs of
	 * parameter name and value strings. A zero byte is required as a
	 * terminator after the last name/value pair. Parameters can appear in any
	 * order. 'user' is required, others are optional.
	 */

	/* Set the 'user' parameter.  This is the only *required* parameter. */
	message.WriteString("user")
	message.WriteString(username)

	/*
	 * Set the 'database' parameter.  If no database name has been specified,
	 * then the default value is the user's name.
	 */
	message.WriteString("database")
	message.WriteString(database)

	/* Set the remaining options as specified. */
	for option, value := range options {
		message.WriteString(option)
		message.WriteString(value)
	}

	/* The message should end with a NULL byte. */
	message.WriteByte(0x00)

	/* update the msg len */
	message.ResetLength(PGMessageLengthOffsetStartup)

	return message.Bytes()
}
