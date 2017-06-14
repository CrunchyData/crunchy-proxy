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
	"strings"
)

/* PostgreSQL message length offset constants. */
const (
	PGMessageLengthOffsetStartup int = 0
	PGMessageLengthOffset        int = 1
)

// MessageBuffer is a variable-sized byte buffer used to read and write
// PostgreSQL Frontend and Backend messages.
//
// A separate instance of a MessageBuffer should be use for reading and writing.
type MessageBuffer struct {
	buffer *bytes.Buffer
}

// NewMessageBuffer creates and intializes a new MessageBuffer using message as its
// initial contents.
func NewMessageBuffer(message []byte) *MessageBuffer {
	return &MessageBuffer{
		buffer: bytes.NewBuffer(message),
	}
}

// ReadInt32 reads an int32 from the message buffer.
//
// This function will read the next 4 available bytes from the message buffer
// and return them as an int32. If an error occurs then 0 and the error are
// returned.
func (message *MessageBuffer) ReadInt32() (int32, error) {
	value := make([]byte, 4)

	if _, err := message.buffer.Read(value); err != nil {
		return 0, err
	}

	return int32(binary.BigEndian.Uint32(value)), nil
}

// ReadInt16 reads an int16 from the message buffer.
//
// This function will read the next 2 available bytes from the message buffer
// and return them as an int16. If an error occurs then 0 and the error are
// returned.
func (message *MessageBuffer) ReadInt16() (int16, error) {
	value := make([]byte, 2)

	if _, err := message.buffer.Read(value); err != nil {
		return 0, err
	}

	return int16(binary.BigEndian.Uint16(value)), nil
}

// ReadByte reads a byte from the message buffer.
//
// This function will read and return the next available byte from the message
// buffer.
func (message *MessageBuffer) ReadByte() (byte, error) {
	return message.buffer.ReadByte()
}

// ReadBytes reads a variable size byte array defined by count from the message
// buffer.
//
// This function will read and return the number of bytes as specified by count.
func (message *MessageBuffer) ReadBytes(count int) ([]byte, error) {
	value := make([]byte, count)

	if _, err := message.buffer.Read(value); err != nil {
		return nil, err
	}

	return value, nil
}

// ReadString reads a string from the message buffer.
//
// This function will read and return the next Null terminated string from the
// message buffer.
func (message *MessageBuffer) ReadString() (string, error) {
	str, err := message.buffer.ReadString(0x00)
	return strings.Trim(str, "\x00"), err
}

// WriteByte will write the specified byte to the message buffer.
func (message *MessageBuffer) WriteByte(value byte) error {
	return message.buffer.WriteByte(value)
}

// WriteBytes writes a variable size byte array specified by 'value' to the
// message buffer.
//
// This function will return the number of bytes written, if the buffer is not
// large enough to hold the value then an error is returned.
func (message *MessageBuffer) WriteBytes(value []byte) (int, error) {
	return message.buffer.Write(value)
}

// WriteInt16 will write a 2 byte int16 to the message buffer.
func (message *MessageBuffer) WriteInt16(value int16) (int, error) {
	x := make([]byte, 2)
	binary.BigEndian.PutUint16(x, uint16(value))
	return message.WriteBytes(x)
}

// WriteInt32 will write a 4 byte int32 to the message buffer.
func (message *MessageBuffer) WriteInt32(value int32) (int, error) {
	x := make([]byte, 4)
	binary.BigEndian.PutUint32(x, uint32(value))
	return message.WriteBytes(x)
}

// WriteString will write a NULL terminated string to the buffer.  It is
// assumed that the incoming string has *NOT* been NULL terminated.
func (message *MessageBuffer) WriteString(value string) (int, error) {
	return message.buffer.WriteString((value + "\000"))
}

// ResetLength will reset the message length for the message.
//
// offset should be one of the PGMessageLengthOffset* constants.
func (message *MessageBuffer) ResetLength(offset int) {
	/* Get the contents of the buffer. */
	b := message.buffer.Bytes()

	/* Get the start of the message length bytes. */
	s := b[offset:]

	/* Determine the new length and set it. */
	binary.BigEndian.PutUint32(s, uint32(len(s)))
}

// Bytes gets the contents of the message buffer. This function is only
// useful after 'Write' operations as the underlying implementation will return
// the 'unread' portion of the buffer.
func (message *MessageBuffer) Bytes() []byte {
	return message.buffer.Bytes()
}

// Reset resets the buffer to empty.
func (message *MessageBuffer) Reset() {
	message.buffer.Reset()
}

// Seek moves the current position of the buffer.
func (message *MessageBuffer) Seek(pos int) {
	message.buffer.Next(pos)
}
