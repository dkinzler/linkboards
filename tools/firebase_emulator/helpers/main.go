package main

import (
	"fmt"
	"log"
	"os"

	"github.com/dkinzler/kit/firebase/emulator"

	cli "github.com/urfave/cli/v2"
)

const version string = "0.4.2"

func main() {
	app := &cli.App{
		Name:    "helpers",
		Usage:   "Firebase Emulator Helpers",
		Version: version,
		Flags:   []cli.Flag{},
		Commands: []*cli.Command{
			{
				Name:  "login",
				Usage: "Create and login auth user",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:  "email",
						Value: "test@test.de",
					},
					&cli.StringFlag{
						Name:    "password",
						Aliases: []string{"pw"},
						Value:   "test123",
					},
					&cli.BoolFlag{
						Name:  "verified",
						Value: true,
					},
				},
				Action: func(ctx *cli.Context) error {
					err := createAndLoginAuthUser(ctx.String("email"), ctx.String("password"), ctx.Bool("verified"))
					return err
				},
			},
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}

func createAndLoginAuthUser(email, password string, verified bool) error {
	client, err := emulator.NewAuthEmulatorClient()
	if err != nil {
		return err
	}

	uid, err := client.CreateUser(email, password, verified)
	if err != nil {
		if emulator.IsEmailAlreadyExistsError(err) {
			fmt.Println("User with email already exists")
		} else {
			return err
		}
	} else {
		fmt.Println("Created user with id:", uid)
	}

	token, err := client.SignInUser(email, password)
	if err != nil {
		return err
	}

	fmt.Println("Token:", token)
	return nil
}
