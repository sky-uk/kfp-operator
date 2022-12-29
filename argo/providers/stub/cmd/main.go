package main

import (
	. "github.com/sky-uk/kfp-operator/providers/base"
	"github.com/sky-uk/kfp-operator/providers/stub"
	"github.com/urfave/cli"
	"os"
)

func main() {
	app := NewProviderApp[stub.StubProviderConfig]()
	provider := stub.StubProvider{}
	app.Run(provider, cli.Command{
		Name: "compile",
		Flags: []cli.Flag{cli.StringFlag{
			Name:     "output-file",
			Required: true,
		}},
		Action: func(c *cli.Context) error {
			return os.WriteFile(c.String("output-file"), []byte("this is a resource"), 0644)
		},
	})
}
