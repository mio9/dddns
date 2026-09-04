package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"runtime"

	"mio9/dddns/internal/config"
	"mio9/dddns/internal/updater"

	"github.com/urfave/cli/v3"
)

var defaultConfigNames = []string{"dddns.yaml", "dddns.yml"}

func resolveConfigPath(flagPath string) (string, error) {
	if flagPath != "" {
		return flagPath, nil
	}

	workDir, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("resolve config path: %w", err)
	}

	for _, name := range defaultConfigNames {
		candidate := filepath.Join(workDir, name)
		if _, err := os.Stat(candidate); err == nil {
			return candidate, nil
		}
	}

	return "", fmt.Errorf("config file not found: pass --config or create dddns.yaml or dddns.yml in %s", workDir)
}

var Version string = "indev"

func main() {
	command := &cli.Command{
		Name:  "dddns",
		Usage: "Dynamic DNS updater",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "config",
				Aliases: []string{"c"},
				Usage:   "Path to YAML config file (default: dddns.yaml or dddns.yml in current directory)",
			},
			&cli.BoolFlag{
				Name:    "version",
				Aliases: []string{"v"},
				Usage:   "Show version",
			},
		},
		Action: func(ctx context.Context, cmd *cli.Command) error {
			if cmd.Bool("version") {
				fmt.Printf("%s %s/%s\n", Version, runtime.GOOS, runtime.GOARCH)
				return nil
			}
			configPath, err := resolveConfigPath(cmd.String("config"))
			if err != nil {
				return err
			}
			cfg, err := config.Load(configPath)
			if err != nil {
				return err
			}
			return updater.Run(ctx, cfg, configPath)
		},
	}

	if err := command.Run(context.Background(), os.Args); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
