package main

import (
	"github.com/dustin/go-humanize"
	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/jedib0t/go-pretty/v6/text"
	"github.com/urfave/cli/v2"
	"os"
)

var SessionCommand = &cli.Command{
	Name:   "session",
	Usage:  "get decrypt Backup.db session list",
	Action: actionSession,
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:    "db",
			Usage:   "decrypted Backup.db file path",
			Value:   "Backup.db",
			Aliases: []string{"d"},
		},
	},
}

func actionSession(ctx *cli.Context) error {
	dbName := ctx.String("db")

	db, err := NewBackupDB(dbName)
	if err != nil {
		return err
	}
	defer db.Close()

	sessions, err := db.Sessions()
	if err != nil {
		return err
	}

	style := table.StyleDefault
	style.Name = "CustomTable"
	style.Format.Header = text.FormatDefault
	style.Format.Footer = text.FormatDefault

	t := table.NewWriter()
	t.SetStyle(style)
	t.SetOutputMirror(os.Stdout)
	t.AppendHeader(table.Row{"Talker", "NickName", "TotalSize"})

	totalSize := int64(0)
	for i := 0; i < len(sessions); i++ {
		s := sessions[i]
		t.AppendRow(table.Row{
			s.Talker,
			s.NickName,
			humanize.Bytes(uint64(s.TotalSize)),
		})
		totalSize += s.TotalSize
	}

	t.AppendSeparator()
	t.AppendFooter(table.Row{"", "Total", humanize.Bytes(uint64(totalSize))})
	t.Render()
	return nil
}
