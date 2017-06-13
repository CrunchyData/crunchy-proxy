package grpcutil

import (
	"io"
	"strings"

	"golang.org/x/net/context"

	"github.com/pkg/errors"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/transport"
)

func IsClosedConnection(err error) bool {
	err = errors.Cause(err)

	if err == context.Canceled ||
		grpc.Code(err) == codes.Canceled ||
		grpc.Code(err) == codes.Unavailable ||
		grpc.ErrorDesc(err) == grpc.ErrClientConnClosing.Error() ||
		strings.Contains(err.Error(), "is closing") ||
		strings.Contains(err.Error(), "tls: use of closed connection") ||
		strings.Contains(err.Error(), "use of closed network connection") ||
		strings.Contains(err.Error(), io.ErrClosedPipe.Error()) ||
		strings.Contains(err.Error(), io.EOF.Error()) {
		return true
	}

	if streamErr, ok := err.(transport.StreamError); ok && streamErr.Code == codes.Canceled {
		return true
	}

	return false
}
