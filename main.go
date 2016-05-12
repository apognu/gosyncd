package main

import (
	"os"

	"github.com/codegangsta/cli"
)

func main() {
	app := cli.NewApp()
	app.Author = "Antoine POPINEAU <antoine.popineau@appscho.com>"
	app.Version = "0.1"

	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:  "config",
			Usage: "configuration file to gosyncd",
			Value: "/etc/gosyncd/config.hcl",
		},
		cli.BoolFlag{
			Name:  "debug, d",
			Usage: "dial verbosity to eleven",
		},
	}

	app.Commands = []cli.Command{
		{
			Before: setupConfig,
			Name:   "daemon",
			Usage:  "run gosyncd's daemon",
			Action: daemon,
		},
		{
			Before: setupConfig,
			Name:   "sync",
			Usage:  "trigger a sync",
			Action: sync,
			Flags: []cli.Flag{
				cli.BoolFlag{
					Name:  "dry",
					Usage: "print what would be done",
				},
			},
		},
		{
			Before: setupConfig,
			Name:   "state",
			Usage:  "print what would be synced",
			Action: sync,
		},
	}

	app.Run(os.Args)
}
