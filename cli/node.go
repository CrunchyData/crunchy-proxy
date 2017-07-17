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
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"
	"golang.org/x/net/context"
	"google.golang.org/grpc"

	pb "github.com/crunchydata/crunchy-proxy/server/serverpb"
)

var nodeCmd = &cobra.Command{
	Use:   "node",
	Short: "show information about configured nodes",
	RunE:  runNode,
}

var nodeListCmd = &cobra.Command{}

func init() {
	flags := nodeCmd.Flags()

	stringFlag(flags, &host, FlagAdminHost)
	stringFlag(flags, &port, FlagAdminPort)
	stringFlag(flags, &format, FlagOutputFormat)
}

func runNode(cmd *cobra.Command, args []string) error {
	address := fmt.Sprintf("%s:%s", host, port)

	dialOptions := []grpc.DialOption{
		grpc.WithDialer(adminServerDialer),
		grpc.WithInsecure(),
	}

	conn, err := grpc.Dial(address, dialOptions...)

	if err != nil {
		fmt.Println(err.Error())
	}

	defer conn.Close()

	c := pb.NewAdminClient(conn)

	response, err := c.Nodes(context.Background(), &pb.NodeRequest{})

	if err != nil {
		fmt.Println(err.Error())
	}

	var result string
	nodes := response.GetNodes()

	switch format {
	case "json":
		j, _ := json.Marshal(nodes)
		result = string(j)
	case "plain":
		for name, node := range nodes {
			result += fmt.Sprintf("* %s - %s\n", name, node)
		}
	default:
		result = fmt.Sprintf("Error: Unsupported format - '%s'", format)
	}

	fmt.Println(result)

	return nil
}
