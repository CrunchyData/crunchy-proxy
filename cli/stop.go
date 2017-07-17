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

package cli

import (
	"fmt"

	"github.com/spf13/cobra"
	"golang.org/x/net/context"
	"google.golang.org/grpc"

	pb "github.com/crunchydata/crunchy-proxy/server/serverpb"
)

var stopCmd = &cobra.Command{
	Use:   "stop",
	Short: "stop a running instance of a proxy",
	RunE:  runStop,
}

func init() {
	flags := stopCmd.Flags()

	stringFlag(flags, &host, FlagAdminHost)
	stringFlag(flags, &port, FlagAdminPort)
}

func runStop(cmd *cobra.Command, args []string) error {
	address := fmt.Sprintf("%s:%s", host, port)

	dialOptions := []grpc.DialOption{
		grpc.WithDialer(adminServerDialer),
		grpc.WithInsecure(),
	}

	conn, err := grpc.Dial(address, dialOptions...)

	if err != nil {
		fmt.Println(err)
	}

	client := pb.NewAdminClient(conn)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	client.Shutdown(ctx, &pb.ShutdownRequest{})
	conn.Close()

	return nil
}
