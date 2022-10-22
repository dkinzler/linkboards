package main

import (
	"log"
	"os"

	cli "github.com/urfave/cli/v2"
)

const version string = "1.0"

func main() {
	app := &cli.App{
		Name:    "Example",
		Usage:   "great app to do great thing",
		Version: version,
		Flags: []cli.Flag{
			&cli.IntFlag{
				Name:    "port",
				Value:   9001,
				EnvVars: []string{"PORT"},
				Aliases: []string{"p"},
			},
			&cli.StringFlag{
				Name:    "address",
				Value:   "",
				EnvVars: []string{"address"},
				Aliases: []string{"a"},
			},
			&cli.BoolFlag{
				Name:  "inmem",
				Value: false,
			},
			&cli.BoolFlag{
				Name:  "emulators",
				Value: false,
			},
			&cli.StringFlag{
				Name:  "firebaseProjectId",
				Value: "",
			},
			&cli.StringFlag{
				Name:  "firebaseServiceAccountFile",
				Value: "",
			},
			&cli.BoolFlag{
				Name:  "debug",
				Value: false,
			},
		},
		Action: func(ctx *cli.Context) error {
			config := Config{
				Port:                       ctx.Int("port"),
				Address:                    ctx.String("address"),
				UseInmemDatastores:         ctx.Bool("inmem"),
				UseFirebaseEmulators:       ctx.Bool("emulators"),
				FirebaseProjectId:          ctx.String("firebaseProjectId"),
				FirebaseServiceAccountFile: ctx.String("firebaseServiceAccountFile"),
				DebugMode:                  ctx.Bool("debug"),
			}
			return runApp(config)
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}
