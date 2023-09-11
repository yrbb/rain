package rain

import (
	"github.com/spf13/cobra"
)

const VERSION = "0.0.1"

var (
	Version = "0000000"
	Compile = "2006-01-02 15:04:05 +0000 by go version go1.21.1 darwin/amd64"
)

func registerVersionCommand(p *Rain) {
	p.cmd.AddCommand(&cobra.Command{
		Use: "version",
		Run: func(cmd *cobra.Command, args []string) {},
	})
}
