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
	Use:     "stats [options]",
	Short:   "",
	Long:    "",
	Example: "",
	RunE:    runStats,
}

func init() {
	flags := statsCmd.Flags()

	flags.StringVarP(&host, "host", "", "localhost", "")
	flags.StringVarP(&host, "port", "", "8000", "")
	flags.StringVarP(&format, "format", "", "plain", "")
}

func runStats(cmd *cobra.Command, args []string) error {
	address := fmt.Sprintf("%s:%s", host, port)
	conn, err := grpc.Dial(address, grpc.WithInsecure())

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
