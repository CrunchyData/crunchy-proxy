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
	"github.com/golang/glog"
)

var START = []byte{'/', '*'}
var END = []byte{'*', '/'}

//the annotation approach
//assume a write if there is no comment in the SQL
//or if there are no keywords in the comment
// return (write, start, finish) booleans
func IsWriteAnno(readAnno string, buf []byte) (write bool, start bool, finish bool) {
	write, start, finish = false, false, false
	var msgLen int32
	var query string

	msgLen = int32(buf[1])<<24 | int32(buf[2])<<16 | int32(buf[3])<<8 | int32(buf[4])
	query = string(buf[5:msgLen])
	glog.V(2).Infof("IsWrite: msglen=%d query=%s\n", msgLen, query)

	querybuf := buf[5:msgLen]
	startPos := bytes.Index(buf, START)
	glog.V(2).Infof("IsWrite: startPos=%d\n", startPos)
	endPos := bytes.Index(buf, END)
	//adding startPos != 5 forces the annotation to start
	//at the beginning of the SQL statement
	if startPos < 0 || startPos != 5 || endPos < 0 {
		glog.V(2).Infoln("no comment found..assuming write case and stateful")
		write = true
		return write, start, finish
	}
	startPos = startPos + 5 //add 5 for msg header length
	endPos = endPos + 5     //add 5 for msg header length

	comment := buf[bytes.Index(querybuf, START)+2+5 : bytes.Index(querybuf, END)+5]
	glog.V(3).Infof("comment=[%s]\n", string(comment))

	keywords := bytes.Split(comment, []byte(","))
	var keywordFound = false
	for i := 0; i < len(keywords); i++ {
		glog.V(3).Infof("keyword=[%s]\n", string(bytes.TrimSpace(keywords[i])))
		if string(bytes.TrimSpace(keywords[i])) == readAnno {
			glog.V(2).Infoln("read was found")
			write = false
			keywordFound = true
		}
		if string(bytes.TrimSpace(keywords[i])) == "start" {
			glog.V(2).Infoln("start was found")
			start = true
			keywordFound = true
		}
		if string(bytes.TrimSpace(keywords[i])) == "finish" {
			glog.V(2).Infoln("finish was found")
			finish = true
			keywordFound = true
		}
	}

	glog.V(3).Infof("write=%t start=%t finish=%t\n", write, start, finish)
	if keywordFound == false {
		glog.V(3).Infoln("no keywords found in SQL comment..assuming write")
		write = true
	}

	return write, start, finish
}
