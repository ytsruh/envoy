package tools

import (
	"context"
	"flag"
	"fmt"

	cli "github.com/pressly/cli"
	"ytsruh.com/envoy/cli/tools/license"
)

var licenseCmd = &cli.Command{
	Name:      "license",
	ShortHelp: "Generate a license file",
	Usage:     "envoy tools license [--name=NAME] [--year=YEAR] [--license=TYPE] [--output=PATH]",
	Flags: cli.FlagsFunc(func(f *flag.FlagSet) {
		f.String("name", "", "Copyright holder name")
		f.String("year", "", "Copyright year (defaults to current year)")
		f.String("license", "MIT", "License type (MIT, Apache-2.0, GPL-3.0, AGPL-3.0, MPL-2.0, CC0-1.0, Unlicense)")
		f.String("output", "", "Output file path (defaults to ./LICENSE)")
	}),
	Exec: func(ctx context.Context, s *cli.State) error {
		name := cli.GetFlag[string](s, "name")
		year := cli.GetFlag[string](s, "year")
		licenseType := cli.GetFlag[string](s, "license")
		output := cli.GetFlag[string](s, "output")

		if err := license.GenerateLicense(name, year, licenseType, output); err != nil {
			return fmt.Errorf("failed to generate license: %w", err)
		}

		fmt.Fprintf(s.Stdout, "License file generated successfully.\n")
		return nil
	},
}
