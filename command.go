package rain

import (
	"fmt"
	"log/slog"

	"github.com/spf13/cobra"

	"github.com/yrbb/rain/pkg/logger"
)

var RootCmd = &cobra.Command{
	Use:  "rain",
	Args: cobra.MinimumNArgs(1),
}

func init() {
	RootCmd.PersistentFlags().StringP("config", "c", "", "config file")
	// RootCmd.MarkPersistentFlagRequired("config")
	// RootCmd.MarkFlagFilename("config", "toml")
}

func wrapCommand(p *Rain) {
	var fn func(cmd *cobra.Command)
	fn = func(cmd *cobra.Command) {
		for _, v := range cmd.Commands() {
			wrapCommandRunE(p, v)

			if lc := len(v.Commands()); lc > 0 {
				fn(v)
			}
		}
	}

	fn(RootCmd)
}

func wrapCommandRunE(p *Rain, cmd *cobra.Command) {
	if cmd.Run == nil && cmd.RunE == nil {
		return
	}

	wrapRunE := func(method func(cmd *cobra.Command, args []string)) func(cmd *cobra.Command, args []string) error {
		return func(cmd2 *cobra.Command, args []string) error {
			p.beforeStartCallback()

			if name := cmd2.Name(); name != "version" && name != "server" {
				logger.M().Info("开始执行 Command: " + name)
			}

			go method(cmd2, args)

			p.serverReadyCheck()
			p.listenSignals()

			return nil
		}
	}

	if cmd.RunE == nil {
		cmd.RunE = wrapRunE(func(cmd *cobra.Command, args []string) {
			cmd.Run(cmd, args)
			p.stop()
		})
		return
	}

	origRunE := cmd.RunE
	cmd.RunE = wrapRunE(func(cmd *cobra.Command, args []string) {
		if err := origRunE(cmd, args); err != nil {
			logger.M().Error(
				fmt.Sprintf("Command %s 异常退出", cmd.Name()),
				slog.String("error", err.Error()),
			)

			p.stop()
		}
	})
}
