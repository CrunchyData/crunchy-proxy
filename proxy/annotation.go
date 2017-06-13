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
