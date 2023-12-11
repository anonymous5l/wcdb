package main

import (
	_ "github.com/mattn/go-sqlite3"
	"github.com/urfave/cli/v2"
	"os"
)

func main() {
	app := &cli.App{
		Name:  "wcdb",
		Usage: "wechat backup database tools",
		Authors: []*cli.Author{
			{
				Name:  "Anonymous5L",
				Email: "wxdxfg@hotmail.com",
			},
		},
		Commands: []*cli.Command{
			DumpCommand,
			SessionCommand,
			ChatCommand,
			DecryptCommand,
			ResourcesCommand,
		},
	}

	if err := app.Run(os.Args); err != nil {
		panic(err)
	}
}
