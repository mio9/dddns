package main

import (
	"context"
	"fmt"
	"os"

	"github.com/urfave/cli/v3"
)

func main() {
	command := &cli.Command{
		Name:  "dddns",
		Usage: "DDNS for Cloudflare",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:     "config",
				Aliases:  []string{"c"},
				Usage:    "Path to YAML config file",
				Required: true,
			},
		},
		Action: func(ctx context.Context, cmd *cli.Command) error {
			config, err := loadConfig(cmd.String("config"))
			if err != nil {
				return err
			}
			return updateDNS(ctx, config)
		},
	}

	if err := command.Run(context.Background(), os.Args); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
