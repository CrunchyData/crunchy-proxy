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

	conn, err := grpc.Dial(address, grpc.WithInsecure())

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
