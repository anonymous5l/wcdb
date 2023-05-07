package main

import (
	"crypto/sha1"
	"encoding/hex"
	"errors"
	"github.com/urfave/cli/v2"
	"os"
)

type Pass string

func (p Pass) Valid() bool {
	if len(p) != 32 {
		return false
	}
	if _, err := hex.DecodeString(string(p)); err != nil {
		return false
	}
	return true
}

var ErrInvalidPassKey = errors.New("invalid pass key")

var DumpCommand = &cli.Command{
	Name:   "dump",
	Usage:  "decrypt and dump Backup.db to normalize sqlite3 file",
	Action: actionDump,
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:     "input",
			Usage:    "Backup.db input file",
			Required: true,
			Aliases:  []string{"i"},
		},
		&cli.StringFlag{
			Name:    "output",
			Usage:   "Backup.db output file",
			Value:   "Backup.db",
			Aliases: []string{"o"},
		},
		&cli.StringFlag{
			Name:     "pass",
			Usage:    "decrypt key",
			Required: true,
			Aliases:  []string{"p"},
		},
	},
}

const (
	DefaultPageSize = 4096
	DefaultKdfIter  = 64000
)

func actionDump(ctx *cli.Context) error {
	inputFilename := ctx.String("input")
	outputFilename := ctx.String("output")
	pass := Pass(ctx.String("pass"))
	if !pass.Valid() {
		return ErrInvalidPassKey
	}

	input, err := os.Open(inputFilename)
	if err != nil {
		return err
	}
	defer input.Close()

	output, err := os.Create(outputFilename)
	if err != nil {
		return err
	}
	defer output.Close()

	cipher, err := NewSqlcipher([]byte(pass), DefaultPageSize, DefaultKdfIter, sha1.New, input)
	if err != nil {
		return err
	}

	if _, err = output.Write(SQLiteHead); err != nil {
		return err
	}

	var buf []byte
	for cipher.Next() {
		buf, err = cipher.Data()
		if err != nil {
			return err
		}
		if _, err = output.Write(buf); err != nil {
			return err
		}
	}

	return nil
}
