package tools

import (
	cli "github.com/pressly/cli"
)

var ToolsCmd = &cli.Command{
	Name:      "tools",
	ShortHelp: "Various developer tools (password, license, pomo)",
	SubCommands: []*cli.Command{
		passwordCmd,
		licenseCmd,
		pomoCmd,
	},
}
