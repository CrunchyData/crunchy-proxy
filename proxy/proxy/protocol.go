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
	"github.com/crunchydata/crunchy-proxy/proxy/config"
	"log"
	"net"
)

func ProtocolMsgType(buf []byte) string {
	return string(buf[0])
}

func LogProtocol(direction string, hint string, buf []byte, bufLen int) {
	var msgType byte
	if hint == "startup" {
		log.Printf("[protocol] %s %s [%s]\n", direction, hint, "startup")
		StartupRequest(buf, bufLen)
		return
	} else {
		//log.Printf("protocol dump: hex=%x char=%c all=%s\n", buf[0], buf[0], string(buf[0:bufLen-1]))
		//msgType = string(buf[0])
		msgType = buf[0]
		//msgLen = int32(buf[1])<<24 | int32(buf[2])<<16 | int32(buf[3])<<8 | int32(buf[4])
		log.Printf("[protocol] %s %s [%c]\n", direction, hint, msgType)
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
			log.Printf("[protocol] %s %s [%c] NOT handled!!\n", direction, hint, msgType)
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
	var msgLen int32
	var mtype int32
	msgLen = int32(buf[1])<<24 | int32(buf[2])<<16 | int32(buf[3])<<8 | int32(buf[4])
	mtype = int32(buf[5])<<24 | int32(buf[6])<<16 | int32(buf[7])<<8 | int32(buf[8])
	var salt = []byte{buf[9], buf[10], buf[11], buf[12]}
	var saltstr = string(salt)
	log.Printf("[protocol] AuthenticationRequest: msglen=%d type=%d salt=%x saltstr=%s\n", msgLen, mtype, salt, saltstr)
	return salt
}

func ErrorResponse(buf []byte) {
	var msgLen int32
	msgLen = int32(buf[1])<<24 | int32(buf[2])<<16 | int32(buf[3])<<8 | int32(buf[4])
	log.Printf("[protocol] ErrorResponse: msglen=%d\n", msgLen)
	var errorMessage = string(buf[5:msgLen])
	log.Printf("[protocol] ErrorResponse: message=%s\n", errorMessage)
}

func StartupRequest(buf []byte, bufLen int) {
	var msgLen int32
	var startupProtocol int32
	//var parameters []string
	msgLen = int32(buf[0])<<24 | int32(buf[1])<<16 | int32(buf[2])<<8 | int32(buf[3])
	startupProtocol = int32(buf[4])<<24 | int32(buf[5])<<16 | int32(buf[6])<<8 | int32(buf[7])
	log.Printf("[protocol] StartupRequest: msglen=%d protocol=%d\n", msgLen, startupProtocol)
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
	msgLen = int32(buf[1])<<24 | int32(buf[2])<<16 | int32(buf[3])<<8 | int32(buf[4])
	query = string(buf[5:msgLen])
	log.Printf("[protocol] QueryRequest: msglen=%d query=%s\n", msgLen, query)
}

func NoticeResponse(buf []byte) {
	var msgLen int32
	//var query string
	msgLen = int32(buf[1])<<24 | int32(buf[2])<<16 | int32(buf[3])<<8 | int32(buf[4])
	//log.Printf("[protocol] NoticeResponse: msglen=%d\n", msgLen)
	var fieldType = buf[5]
	var fieldMsg = string(buf[6:msgLen])
	log.Printf("[protocol] NoticeResponse: msglen=%d fieldType=%x fieldMsg=%s\n", msgLen, fieldType, fieldMsg)
}

func RowDescription(buf []byte, bufLen int) {
	var msgLen int32
	msgLen = int32(buf[1])<<24 | int32(buf[2])<<16 | int32(buf[3])<<8 | int32(buf[4])
	log.Printf("[protocol] RowDescription: msglen=%d\n", msgLen)
	//query = string(buf[5:msgLen])
	var data []byte

	data = buf[4+msgLen : bufLen]

	var dataRowType = string(data[0])
	log.Printf("[protocol] datarow type%s found \n", dataRowType)
	//	msgLen = int32(data[bufPtr+1])<<24 | int32(data[bufPtr+2])<<16 | int32(data[bufPtr+3])<<8 | int32(data[bufPtr+4])
	//	log.Printf("[protocol] datarow type%s found with msglen=%d\n", dataRowType, msgLen)
	//
	//	data = buf[bufPtr : msgLen+5]
	//	DataRow(data)
	//	bufPtr = bufPtr + msgLen + 5

}
func DataRow(buf []byte) {
	var numfields int
	var msgLen int32
	var fieldLen int32
	var fieldValue string
	msgLen = int32(buf[1])<<24 | int32(buf[2])<<16 | int32(buf[3])<<8 | int32(buf[4])
	numfields = int(buf[5])<<8 | int(buf[6])
	fieldLen = int32(buf[7])<<24 | int32(buf[8])<<16 | int32(buf[9])<<8 | int32(buf[10])
	fieldValue = string(buf[11 : fieldLen+11])
	log.Printf("[protocol] DataRow: numfields=%d msglen=%d fieldLen=%d fieldValue=%s\n", numfields, msgLen, fieldLen, fieldValue)
	//var data = string(buf[7:msgLen])
	//log.Printf("[protocol] DataRow: data=%s\n", data)
}
func CommandComplete(buf []byte) {
	var msgLen int32
	msgLen = int32(buf[1])<<24 | int32(buf[2])<<16 | int32(buf[3])<<8 | int32(buf[4])
	log.Printf("[protocol] Command Complete: msglen=%d\n", msgLen)
}
func TerminateMessage(buf []byte) {
	var msgLen int32
	msgLen = int32(buf[1])<<24 | int32(buf[2])<<16 | int32(buf[3])<<8 | int32(buf[4])
	log.Printf("[protocol] Terminate: msglen=%d\n", msgLen)
	//query = string(buf[5:msgLen])
	//log.Printf("[protocol] RowDescription: msglen=%d query=%s\n", msgLen, query)
}
func PasswordMessage(buf []byte) {
	var msgLen int32
	msgLen = int32(buf[1])<<24 | int32(buf[2])<<16 | int32(buf[3])<<8 | int32(buf[4])
	var hash = string(buf[5:msgLen])

	log.Printf("[protocol] PasswordMessage: msglen=%d password hash=%s\n", msgLen, hash)
}
func PasswordMessageFake(buf []byte, salt []byte, username string, password string) {
	var msgLen int32
	msgLen = int32(buf[1])<<24 | int32(buf[2])<<16 | int32(buf[3])<<8 | int32(buf[4])
	var hash = string(buf[5:msgLen])

	log.Printf("[protocol] PasswordMessageFake: username=%s password=%s\n", username, password)
	log.Printf("[protocol] PasswordMessageFake: msglen=%d password hash=%s salt=%x saltlen=%d\n", msgLen, hash, salt, len(salt))

	s := string(salt)
	hashstr := "md5" + md5s(md5s(password+username)+s)

	log.Printf("[protocol] PasswordMessageFake: hashstr=%s\n", hashstr)
	hashbytes := []byte(hashstr)
	copy(buf[5:], hashbytes)
	log.Println("generated hash " + hashstr)
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
		log.Println(err.Error() + " at this pt")
	}
	log.Printf("wrote %d to backend\n", writeLen)

	//read from backend
	buf = make([]byte, 2048)
	readLen, err = conn.Read(buf)
	if err != nil {
		log.Println(err.Error() + " at this pt2")
	}

	//should get back an AuthenticationRequest 'R'
	LogProtocol("<--", "pool node", buf, len(buf))
	msgType := ProtocolMsgType(buf)
	if msgType != "R" {
		log.Println("pool error: should have got R message here")
	}
	salt := AuthenticationRequest(buf)
	log.Printf("salt from AuthenticationRequest was %s %x\n", string(salt), salt)

	//create password message and send back to backend
	pswMsg := getPasswordMessage(salt, cfg.Credentials.Username, cfg.Credentials.Password)

	//write to backend
	writeLen, err = conn.Write(pswMsg)
	log.Printf("wrote %d to backend\n", writeLen)
	if err != nil {
		log.Println(err.Error() + " at this pta")
	}

	//read from backend
	readLen, err = conn.Read(buf)
	if err != nil {
		log.Println(err.Error() + " at this pt3")
	}

	msgType = ProtocolMsgType(buf)
	log.Println("after passwordmsg got msgType " + msgType)
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

	log.Printf("[protocol] getPasswordMessage: hashstr=%s\n", hashstr)
	buffer = append(buffer, hashstr...)

	//null terminate the string
	buffer = append(buffer, 0)

	//update the msg len subtracting for msgType byte
	binary.BigEndian.PutUint32(x, uint32(len(buffer)-1))
	copy(buffer[1:], x)

	log.Printf(" psw msg len=%d\n", len(buffer))
	log.Printf(" psw msg =%s\n", string(buffer))
	return buffer

}

func getStartupMessage(cfg *config.Config, node *config.Node) []byte {

	//send startup packet
	var buffer []byte

	x := make([]byte, 4)

	//make msg len 1 for now
	binary.BigEndian.PutUint32(x, uint32(1))
	buffer = append(buffer, x...)

	//w.int32(196608)
	binary.BigEndian.PutUint32(x, uint32(196608))
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
