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

var healthCmd = &cobra.Command{
	Use:   "health [options]",
	Short: "show health check information for configured nodes",
	RunE:  runHealth,
}

func init() {
	flags := healthCmd.Flags()

	stringFlag(flags, &host, FlagAdminHost)
	stringFlag(flags, &port, FlagAdminPort)
	stringFlag(flags, &format, FlagOutputFormat)
}

func runHealth(cmd *cobra.Command, args []string) error {
	var result string
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

	response, err := client.Health(ctx, &pb.HealthRequest{})
	conn.Close()

	if err != nil {
		fmt.Printf("Error: %s", err.Error())
		return err
	}

	switch format {
	case "json":
		j, _ := json.Marshal(response.GetHealth())
		result = string(j)
	case "plain":
		result = formatPlain(response.GetHealth())
	default:
		result = fmt.Sprintf("Error: Unsupported format '%s'", format)
	}

	fmt.Println(result)

	return nil
}

func formatPlain(hc map[string]bool) string {
	var result string

	for name, healthy := range hc {
		result += fmt.Sprintf("* %s: %t\n", name, healthy)
	}

	return result
}
