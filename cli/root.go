package cli

import (
	"context"
	"flag"
	"fmt"
	"os"

	cli "github.com/pressly/cli"
	"ytsruh.com/envoy/cli/tools"
	"ytsruh.com/envoy/cli/utils"
)

var Root = &cli.Command{
	Name:      "envoy",
	ShortHelp: "Envoy CLI client for managing projects and environments",
	Usage:     "envoy <command> [flags]",
	SubCommands: []*cli.Command{
		versionCmd,
		authCmd,
		projectsCmd,
		environmentsCmd,
		environmentVariablesCmd,
		usersCmd,
		tools.ToolsCmd,
	},
}

var versionCmd = &cli.Command{
	Name:      "version",
	ShortHelp: "Print the version number",
	Exec: func(ctx context.Context, s *cli.State) error {
		fmt.Fprintf(s.Stdout, "v%s\n", utils.Version)
		return nil
	},
}

func Execute() {
	ctx := context.Background()
	if err := cli.ParseAndRun(ctx, Root, os.Args[1:], nil); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func init() {
	Root.Flags = cli.FlagsFunc(func(f *flag.FlagSet) {
		f.Bool("verbose", false, "enable verbose output")
	})
}
