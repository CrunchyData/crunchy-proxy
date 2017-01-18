/*
 Copyright 2016 Crunchy Data Solutions, Inc.
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

func ProtocolMsgType(buf []byte) (string, int) {
	var msgLen int32

	// Read the message length.
	reader := bytes.NewReader(buf[1:5])
	binary.Read(reader, binary.BigEndian, &msgLen)

	glog.V(2).Infof("[protocol] %d msgLen\n", msgLen)

	return string(buf[0]), int(msgLen)
}

func LogProtocol(direction string, hint string, buf []byte, bufLen int) {
	var msgType byte

	if hint == "startup" {
		glog.V(2).Infof("[protocol] %s %s [%s]\n", direction, hint, "startup")
		StartupRequest(buf, bufLen)
		return
	} else {
		msgType = buf[0]
		glog.V(2).Infof("[protocol] %s %s [%c]\n", direction, hint, msgType)
		switch msgType {
		case 'R':
			AuthenticationRequest(buf)
			return
		case 'E':
			ErrorResponse(buf)
			return
		case 'Q':
			QueryRequest(buf)
			return
		case 'N':
			NoticeResponse(buf)
			return
		case 'T':
			RowDescription(buf, bufLen)
			return
		case 'D':
			DataRow(buf)
			return
		case 'C':
			CommandComplete(buf)
			return
		case 'X':
			TerminateMessage(buf)
			return
		case 'p':
			PasswordMessage(buf)
			return
		default:
			glog.Errorf("[protocol] %s %s [%c] NOT handled!!\n", direction, hint, msgType)
			return
		}
	}
}

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

func AuthenticationRequest(buf []byte) []byte {
	var msgLength int32
	var authType int32

	// Read message length.
	reader := bytes.NewReader(buf[1:5])
	binary.Read(reader, binary.BigEndian, &msgLength)

	// Read authentication type.
	reader.Reset(buf[5:9])
	binary.Read(reader, binary.BigEndian, &authType)

	var salt = []byte{buf[9], buf[10], buf[11], buf[12]}
	var saltstr = string(salt)
	glog.V(2).Infof("[protocol] AuthenticationRequest: msglen=%d type=%d salt=%x saltstr=%s\n", msgLength, authType, salt, saltstr)
	return salt
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
	//parameters = string(buf[8 : bufLen-8])
	/**
	parameters = NullTermToStrings(buf[8 : bufLen-1])
	for i := 0; i < len(parameters); i++ {
		log.Printf("[protocol] startup parameter key:value: %s:%s \n", parameters[i], parameters[i+1])
		i++
	}
	*/
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

func PasswordMessage(buf []byte) {
	var msgLen int32

	reader := bytes.NewReader(buf[1:5])
	binary.Read(reader, binary.BigEndian, &msgLen)

	var hash = string(buf[5:msgLen])

	glog.V(2).Infof("[protocol] PasswordMessage: msglen=%d password hash=%s\n", msgLen, hash)
}

func PasswordMessageFake(buf []byte, salt []byte, username string, password string) {
	var msgLen int32

	reader := bytes.NewReader(buf[1:5])
	binary.Read(reader, binary.BigEndian, &msgLen)

	var hash = string(buf[5:msgLen])

	glog.V(2).Infof("[protocol] PasswordMessageFake: username=%s password=%s\n", username, password)
	glog.V(2).Infof("[protocol] PasswordMessageFake: msglen=%d password hash=%s salt=%x saltlen=%d\n", msgLen, hash, salt, len(salt))

	s := string(salt)
	hashstr := "md5" + md5s(md5s(password+username)+s)

	glog.V(2).Infof("[protocol] PasswordMessageFake: hashstr=%s\n", hashstr)
	hashbytes := []byte(hashstr)
	copy(buf[5:], hashbytes)
	glog.V(2).Infoln("generated hash " + hashstr)
}

func md5s(s string) string {
	h := md5.New()
	h.Write([]byte(s))
	return fmt.Sprintf("%x", h.Sum(nil))
}

func Authenticate(cfg *config.Config, node *config.Node, conn *net.TCPConn) {
	var readLen, writeLen int
	var err error
	var buf []byte

	startupMsg := getStartupMessage(cfg, node)

	//write to backend
	writeLen, err = conn.Write(startupMsg)
	if err != nil {
		glog.Errorln(err.Error() + " at this pt")
	}
	glog.V(2).Infof("wrote %d to backend\n", writeLen)

	//read from backend
	buf = make([]byte, 2048)
	readLen, err = conn.Read(buf)
	if err != nil {
		glog.Errorln(err.Error() + " at this pt2")
	}

	//should get back an AuthenticationRequest 'R'
	LogProtocol("<--", "pool node", buf, len(buf))
	msgType, msgLen := ProtocolMsgType(buf)
	if msgType != "R" {
		glog.Errorln("pool error: should have got R message here")
	}
	salt := AuthenticationRequest(buf)
	glog.V(2).Infof("salt from AuthenticationRequest was %s %x\n", string(salt), salt)

	//create password message and send back to backend
	pswMsg := getPasswordMessage(salt, cfg.Credentials.Username, cfg.Credentials.Password)

	//write to backend
	writeLen, err = conn.Write(pswMsg)
	glog.V(2).Infof("wrote %d to backend\n", writeLen)
	if err != nil {
		glog.Errorln(err.Error() + " at this pta")
	}

	//read from backend
	readLen, err = conn.Read(buf)
	if err != nil {
		glog.Errorln(err.Error() + " at this pt3")
	}

	msgType, msgLen = ProtocolMsgType(buf)
	glog.V(2).Infof("after passwordmsg got msgType %s msgLen=%d\n", msgType, msgLen)
	if msgType == "R" {
		LogProtocol("<--", "AuthenticationOK", buf, readLen)
	}

}

func getPasswordMessage(salt []byte, username string, password string) []byte {
	var buffer []byte

	buffer = append(buffer, 'p')

	//make msg len 1 for now
	x := make([]byte, 4)
	binary.BigEndian.PutUint32(x, uint32(1))
	buffer = append(buffer, x...)

	s := string(salt)
	hashstr := "md5" + md5s(md5s(password+username)+s)

	glog.V(2).Infof("[protocol] getPasswordMessage: hashstr=%s\n", hashstr)
	buffer = append(buffer, hashstr...)

	//null terminate the string
	buffer = append(buffer, 0)

	//update the msg len subtracting for msgType byte
	binary.BigEndian.PutUint32(x, uint32(len(buffer)-1))
	copy(buffer[1:], x)

	glog.V(2).Infof(" psw msg len=%d\n", len(buffer))
	glog.V(2).Infof(" psw msg =%s\n", string(buffer))
	return buffer

}

func getStartupMessage(cfg *config.Config, node *config.Node) []byte {

	//send startup packet
	var buffer []byte

	x := make([]byte, 4)

	//make msg len 1 for now
	binary.BigEndian.PutUint32(x, uint32(1))
	buffer = append(buffer, x...)

	// Set the protocol version.
	binary.BigEndian.PutUint32(x, uint32(PROTOCOL_VERSION))
	buffer = append(buffer, x...)

	var key, value string
	key = "database"
	buffer = append(buffer, key...)
	//null terminate the string
	buffer = append(buffer, 0)

	value = cfg.Credentials.Database
	buffer = append(buffer, value...)
	//null terminate the string
	buffer = append(buffer, 0)

	key = "user"
	buffer = append(buffer, key...)
	//null terminate the string
	buffer = append(buffer, 0)

	value = cfg.Credentials.Username
	buffer = append(buffer, value...)
	//null terminate the string
	buffer = append(buffer, 0)

	key = "client_encoding"
	buffer = append(buffer, key...)
	//null terminate the string
	buffer = append(buffer, 0)

	value = "UTF8"
	buffer = append(buffer, value...)
	//null terminate the string
	buffer = append(buffer, 0)

	key = "datestyle"
	buffer = append(buffer, key...)
	//null terminate the string
	buffer = append(buffer, 0)

	value = "ISO, MDY"
	buffer = append(buffer, value...)
	//null terminate the string
	buffer = append(buffer, 0)

	key = "application_name"
	buffer = append(buffer, key...)
	//null terminate the string
	buffer = append(buffer, 0)

	value = "proxypool"
	buffer = append(buffer, value...)
	//null terminate the string
	buffer = append(buffer, 0)

	key = "extra_float_digits"
	buffer = append(buffer, key...)
	//null terminate the string
	buffer = append(buffer, 0)

	value = "2"
	buffer = append(buffer, value...)
	//null terminate the string
	buffer = append(buffer, 0)

	buffer = append(buffer, 0)

	//update the msg len
	binary.BigEndian.PutUint32(buffer, uint32(len(buffer)))

	return buffer
}
