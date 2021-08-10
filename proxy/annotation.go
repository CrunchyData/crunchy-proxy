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

const (
	ReadAnnotation AnnotationType = iota
	StartAnnotation
	EndAnnotation
)

const (
	readAnnotationString    string = "read"
	startAnnotationString   string = "start"
	endAnnotationString     string = "end"
  	tdfColumnAnnotationString     string = "tdfColumn"
       	tdfDelimiterAnnotationString     string = ":"
	unknownAnnotationString string = ""
)

const (
	AnnotationStartToken = "/*"
	AnnotationEndToken   = "*/"
)

type AnnotationType int

func (a AnnotationType) String() string {
	switch a {
	case ReadAnnotation:
		return readAnnotationString
	case StartAnnotation:
		return startAnnotationString
	case EndAnnotation:
		return endAnnotationString
	}

	return unknownAnnotationString
}
