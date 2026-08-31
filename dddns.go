package main

import (
	"context"
	"fmt"
	"os"

	"mio9/dddns/internal/config"
	"mio9/dddns/internal/updater"

	"github.com/urfave/cli/v3"
)

var Version string = "indev"

func main() {
	command := &cli.Command{
		Name:  "dddns",
		Usage: "DDNS for Cloudflare",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "config",
				Aliases: []string{"c"},
				Usage:   "Path to YAML config file",
			},
			&cli.BoolFlag{
				Name:    "version",
				Aliases: []string{"v"},
				Usage:   "Show version",
			},
		},
		Action: func(ctx context.Context, cmd *cli.Command) error {
			if cmd.Bool("version") {
				fmt.Println(Version)
				return nil
			}
			configPath := cmd.String("config")
			if configPath == "" {
				return fmt.Errorf("config is required")
			}
			cfg, err := config.Load(configPath)
			if err != nil {
				return err
			}
			return updater.Update(ctx, cfg)
		},
	}

	if err := command.Run(context.Background(), os.Args); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
