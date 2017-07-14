package protocol

func CreatePasswordMessage(password string) []byte {
	message := NewMessageBuffer([]byte{})

	/* Set the message type */
	message.WriteByte(PasswordMessageType)

	/* Initialize the message length to zero. */
	message.WriteInt32(0)

	/* Add the password to the message. */
	message.WriteString(password)

	/* Update the message length */
	message.ResetLength(PGMessageLengthOffset)

	return message.Bytes()
}
