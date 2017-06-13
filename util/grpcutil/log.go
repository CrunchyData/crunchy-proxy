package grpcutil

import (
	"fmt"

	"google.golang.org/grpc/grpclog"
)

func init() {
	grpclog.SetLogger(&logger{})
}

type logger struct{}

var _ grpclog.Logger = (*logger)(nil)

func (*logger) Fatal(args ...interface{}) {
	fmt.Println("Fatal", args)
}

func (*logger) Fatalf(format string, args ...interface{}) {
	fmt.Println("Fatalf", args)
}

func (*logger) Fatalln(args ...interface{}) {
	fmt.Println("Fatalln", args)
}

func (*logger) Print(args ...interface{}) {
	fmt.Println("Print", args)
}

func (*logger) Printf(format string, args ...interface{}) {
	grpclog.Printf(format, args)
}

func (*logger) Println(args ...interface{}) {
	fmt.Println("Println", args)
}
