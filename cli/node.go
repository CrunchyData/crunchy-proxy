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
	Use:     "node",
	Short:   "",
	Long:    "",
	Example: "",
	RunE:    runNode,
}

var nodeListCmd = &cobra.Command{}

func init() {
	flags := nodeCmd.Flags()

	flags.StringVarP(&host, "host", "", "localhost", "")
	flags.StringVarP(&port, "port", "", "8000", "")
	flags.StringVarP(&format, "format", "", "plain", "")
}

func runNode(cmd *cobra.Command, args []string) error {
	address := fmt.Sprintf("%s:%s", host, port)

	conn, err := grpc.Dial(address, grpc.WithInsecure())

	if err != nil {
		fmt.Println(err)
	}

	defer conn.Close()

	c := pb.NewAdminClient(conn)

	response, err := c.Nodes(context.Background(), &pb.NodeRequest{})

	if err != nil {
		fmt.Println(err)
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
