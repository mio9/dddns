package main

import (
	"context"
	"fmt"
	"os"

	"github.com/urfave/cli/v3"
)

func main() {
	cli := &cli.Command{
		Name:  "dddns",
		Usage: "DDNS for Cloudflare (for now)",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:  "email",
				Usage: "Cloudflare email",
			},
		},
		Action: func(c context.Context, cmd *cli.Command) error {
			fmt.Println("Hello, World!")
			fmt.Println(cmd.String("email"))
			return nil
		},
	}

	cli.Run(context.Background(), os.Args)

}
