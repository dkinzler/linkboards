package main

import (
	"log"
	"os"

	cli "github.com/urfave/cli/v2"
)

const version string = "1.0.0"

func main() {
	app := &cli.App{
		Name:    "Linkboards",
		Usage:   "A sample API application for sharing links.",
		Version: version,
		Flags: []cli.Flag{
			&cli.IntFlag{
				Name:    "port",
				Value:   9001,
				EnvVars: []string{"PORT"},
				Aliases: []string{"p"},
				Usage:   "port the application will listen on",
			},
			&cli.StringFlag{
				Name:    "address",
				Value:   "",
				EnvVars: []string{"address"},
				Aliases: []string{"a"},
				Usage:   "address the application will listen on",
			},
			&cli.BoolFlag{
				Name:  "inmem",
				Value: false,
				Usage: "use local in-memory dependencies for authentication and data stores, useful for development/testing",
			},
			&cli.BoolFlag{
				Name:  "emulators",
				Value: false,
				Usage: "use firebase emulators for authentication/data stores",
			},
			&cli.StringFlag{
				Name:  "firebaseProjectId",
				Value: "",
			},
			&cli.StringFlag{
				Name:  "firebaseServiceAccountFile",
				Value: "",
				Usage: "path to service account file used to authenticate with Firebase services",
			},
			&cli.BoolFlag{
				Name:  "debug",
				Value: false,
				Usage: "will pretty print JSON log messages",
			},
		},
		Action: func(ctx *cli.Context) error {
			config := Config{
				Port:                       ctx.Int("port"),
				Address:                    ctx.String("address"),
				UseInmemDependencies:       ctx.Bool("inmem"),
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
