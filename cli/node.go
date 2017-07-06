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
