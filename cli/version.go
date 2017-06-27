package cli

import (
	"fmt"

	"github.com/spf13/cobra"
	"golang.org/x/net/context"
	"google.golang.org/grpc"

	pb "github.com/crunchydata/crunchy-proxy/server/serverpb"
)

var versionCmd = &cobra.Command{
	Use:     "version",
	Short:   "show version information for instance of proxy",
	Long:    "",
	Example: "",
	RunE:    runVersion,
}

func init() {
	flags := versionCmd.Flags()

	stringFlag(flags, &host, FlagAdminHost)
	stringFlag(flags, &port, FlagAdminPort)
}

func runVersion(cmd *cobra.Command, args []string) error {
	address := fmt.Sprintf("%s:%s", host, port)

	conn, err := grpc.Dial(address, grpc.WithInsecure())

	if err != nil {
		fmt.Println(err)
	}

	defer conn.Close()

	c := pb.NewAdminClient(conn)

	response, err := c.Version(context.Background(), &pb.VersionRequest{})

	if err != nil {
		fmt.Println(err)
	}

	fmt.Println(response.Version)

	return nil
}
