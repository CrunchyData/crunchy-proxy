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
