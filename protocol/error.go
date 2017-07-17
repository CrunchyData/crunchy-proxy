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
	"fmt"
)

/* PG Error Severity Levels */
const (
	ErrorSeverityFatal   string = "FATAL"
	ErrorSeverityPanic   string = "PANIC"
	ErrorSeverityWarning string = "WARNING"
	ErrorSeverityNotice  string = "NOTICE"
	ErrorSeverityDebug   string = "DEBUG"
	ErrorSeverityInfo    string = "INFO"
	ErrorSeverityLog     string = "LOG"
)

/* PG Error Message Field Identifiers */
const (
	ErrorFieldSeverity         byte = 'S'
	ErrorFieldCode             byte = 'C'
	ErrorFieldMessage          byte = 'M'
	ErrorFieldMessageDetail    byte = 'D'
	ErrorFieldMessageHint      byte = 'H'
	ErrorFieldPosition         byte = 'P'
	ErrorFieldInternalPosition byte = 'p'
	ErrorFieldInternalQuery    byte = 'q'
	ErrorFieldWhere            byte = 'W'
	ErrorFieldSchemaName       byte = 's'
	ErrorFieldTableName        byte = 't'
	ErrorFieldColumnName       byte = 'c'
	ErrorFieldDataTypeName     byte = 'd'
	ErrorFieldConstraintName   byte = 'n'
	ErrorFieldFile             byte = 'F'
	ErrorFieldLine             byte = 'L'
	ErrorFieldRoutine          byte = 'R'
)

type Error struct {
	Severity         string
	Code             string
	Message          string
	Detail           string
	Hint             string
	Position         string
	InternalPosition string
	InternalQuery    string
	Where            string
	SchemaName       string
	TableName        string
	ColumnName       string
	DataTypeName     string
	Constraint       string
	File             string
	Line             string
	Routine          string
}

func (e *Error) Error() string {
	return fmt.Sprintf("pg: %s: %s", e.Severity, e.Message)
}

func (e *Error) GetMessage() []byte {
	msg := NewMessageBuffer([]byte{})

	msg.WriteByte(ErrorMessageType)
	msg.WriteInt32(0)

	msg.WriteByte(ErrorFieldSeverity)
	msg.WriteString(e.Severity)

	msg.WriteByte(ErrorFieldCode)
	msg.WriteString(e.Code)

	msg.WriteByte(ErrorFieldMessage)
	msg.WriteString(e.Message)

	if e.Detail != "" {
		msg.WriteByte(ErrorFieldMessageDetail)
		msg.WriteString(e.Detail)
	}

	if e.Hint != "" {
		msg.WriteByte(ErrorFieldMessageHint)
		msg.WriteString(e.Hint)
	}

	msg.WriteByte(0x00) // null terminate the message

	msg.ResetLength(PGMessageLengthOffset)

	return msg.Bytes()
}

// ParseError parses a PG error message
func ParseError(e []byte) *Error {
	msg := NewMessageBuffer(e)
	msg.Seek(5)
	err := new(Error)

	for field, _ := msg.ReadByte(); field != 0; field, _ = msg.ReadByte() {
		value, _ := msg.ReadString()
		switch field {
		case ErrorFieldSeverity:
			err.Severity = value
		case ErrorFieldCode:
			err.Code = value
		case ErrorFieldMessage:
			err.Message = value
		case ErrorFieldMessageDetail:
			err.Detail = value
		case ErrorFieldMessageHint:
			err.Hint = value
		case ErrorFieldPosition:
			err.Position = value
		case ErrorFieldInternalPosition:
			err.InternalPosition = value
		case ErrorFieldInternalQuery:
			err.InternalQuery = value
		case ErrorFieldWhere:
			err.Where = value
		case ErrorFieldSchemaName:
			err.SchemaName = value
		case ErrorFieldTableName:
			err.TableName = value
		case ErrorFieldColumnName:
			err.ColumnName = value
		case ErrorFieldDataTypeName:
			err.DataTypeName = value
		case ErrorFieldConstraintName:
			err.Constraint = value
		case ErrorFieldFile:
			err.File = value
		case ErrorFieldLine:
			err.Line = value
		case ErrorFieldRoutine:
			err.Routine = value
		}
	}
	return err
}
