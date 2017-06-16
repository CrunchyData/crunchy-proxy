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
