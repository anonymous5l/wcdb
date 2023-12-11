package main

import (
	"fmt"
	"github.com/dustin/go-humanize"
	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/jedib0t/go-pretty/v6/text"
	"github.com/urfave/cli/v2"
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
		&cli.IntFlag{
			Name:    "limit",
			Usage:   "nickname string length limit default 0 no limit",
			Aliases: []string{"l"},
		},
	},
}

func actionSession(ctx *cli.Context) error {
	dbName := ctx.String("db")
	limit := ctx.Int("limit")

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
	t.AppendHeader(table.Row{"Talker", "NickName", "TotalSize"})

	totalSize := int64(0)
	for i := 0; i < len(sessions); i++ {
		s := sessions[i]
		nickname := []rune(s.NickName)
		if limit > 0 && len(nickname) > limit {
			nickname = append(nickname[:limit], []rune("...")...)
		}
		t.AppendRow(table.Row{
			s.Talker,
			string(nickname),
			humanize.Bytes(uint64(s.TotalSize)),
		})
		totalSize += s.TotalSize
	}

	t.AppendSeparator()
	t.AppendFooter(table.Row{"", "Total", humanize.Bytes(uint64(totalSize))})

	fmt.Println(t.Render())

	return nil
}
