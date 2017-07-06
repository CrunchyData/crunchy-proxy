package cli

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"
	"golang.org/x/net/context"
	"google.golang.org/grpc"

	pb "github.com/crunchydata/crunchy-proxy/server/serverpb"
)

var statsCmd = &cobra.Command{
	Use:   "stats [options]",
	Short: "show query statistics for configured nodes",
	RunE:  runStats,
}

func init() {
	flags := statsCmd.Flags()

	stringFlag(flags, &host, FlagAdminHost)
	stringFlag(flags, &port, FlagAdminPort)
	stringFlag(flags, &format, FlagOutputFormat)
}

func runStats(cmd *cobra.Command, args []string) error {
	address := fmt.Sprintf("%s:%s", host, port)

	dialOptions := []grpc.DialOption{
		grpc.WithDialer(adminServerDialer),
		grpc.WithInsecure(),
	}

	conn, err := grpc.Dial(address, dialOptions...)

	if err != nil {
		fmt.Println(err)
	}

	defer conn.Close()

	c := pb.NewAdminClient(conn)

	response, err := c.Statistics(context.Background(), &pb.StatisticsRequest{})

	if err != nil {
		fmt.Println(err)
	}

	var result string
	queries := response.GetQueries()

	switch format {
	case "json":
		j, _ := json.Marshal(queries)
		result = string(j)
	case "plain":
		for name, query := range queries {
			result += fmt.Sprintf("* %s - %d\n", name, query)
		}
	default:
		result = fmt.Sprintf("Error: Unsupported format - '%s'", format)
	}

	fmt.Println(result)

	return nil
}
