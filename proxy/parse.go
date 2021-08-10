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
	"strings"

	"github.com/crunchydata/crunchy-proxy/protocol"
)

// GetAnnotations the annotation approach
// assume a write if there is no comment in the SQL
// or if there are no keywords in the comment
// return (write, start, finish) booleans
func getAnnotations(m []byte) (map[AnnotationType]bool, string) {
	message := protocol.NewMessageBuffer(m)
	annotations := make(map[AnnotationType]bool, 0)
        tdfColumn := ""

	/* Get the query string */
	message.ReadByte()  // read past the message type
	message.ReadInt32() // read past the message length
	query, _ := message.ReadString()

	/* Find the start and end position of the annotations. */
	startPos := strings.Index(query, AnnotationStartToken)
	endPos := strings.Index(query, AnnotationEndToken)

	/*
	 * If the start or end positions are less than zero then that means that
	 * an annotation was not found.
	 */
	if startPos < 0 || endPos < 0 {
		return annotations, tdfColumn
	}

	/* Deterimine which annotations were specified as part of the query */
	keywords := strings.Split(query[startPos+2:endPos], ",")

	for i := 0; i < len(keywords); i++ {
                tmp := strings.Split(keywords[i], tdfDelimiterAnnotationString)
                if len(tmp) == 1 {
		switch strings.TrimSpace(tmp[0]) {
                  case readAnnotationString:
          	    annotations[ReadAnnotation] = true
	        	case startAnnotationString:
	  	     annotations[StartAnnotation] = true
		  case endAnnotationString:
		     annotations[EndAnnotation] = true
		  }
                } else if len(tmp) > 1 {
	  	  switch strings.TrimSpace(tmp[0]) {
                  case tdfColumnAnnotationString:
                       tdfColumn = strings.TrimSpace(tmp[1])
                  }
                } else {
                  return annotations, tdfColumn
                }
	}

	return annotations, tdfColumn
}
