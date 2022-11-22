package main

import (
	. "github.com/sky-uk/kfp-operator/providers/base"
	"github.com/sky-uk/kfp-operator/providers/vai"
	"github.com/urfave/cli"
)

func main() {
	app := NewProviderApp[vai.VAIProviderConfig]()
	provider := vai.VAIProvider{}
	app.Run(provider, cli.Command{
		Name: "vai-run",
		Subcommands: []cli.Command{
			{
				Name: "enqueue",
				Flags: []cli.Flag{cli.StringFlag{
					Name:     "run-intent",
					Required: true,
				}},
				Action: func(c *cli.Context) error {
					providerConfig, err := app.LoadProviderConfig(c)
					if err != nil {
						return err
					}
					runIntent, err := LoadJsonFromFile[vai.RunIntent](c.String("run-intent"))
					if err != nil {
						return err
					}
					return provider.EnqueueRun(app.Context, providerConfig, runIntent)
				},
			},
			{
				Name: "submit",
				Flags: []cli.Flag{cli.StringFlag{
					Name:     "run",
					Required: true,
				}},
				Action: func(c *cli.Context) error {
					providerConfig, err := app.LoadProviderConfig(c)
					vaiRun, err := LoadJsonFromFile[vai.VAIRun](c.String("run"))
					if err != nil {
						return err
					}
					return provider.SubmitRun(app.Context, providerConfig, vaiRun)
				},
			},
		},
	})
}
