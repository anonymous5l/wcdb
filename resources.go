package main

import (
	"bytes"
	"fmt"
	"github.com/urfave/cli/v2"
	"io"
	"os"
	"path/filepath"
	"strconv"
)

var ResourcesCommand = &cli.Command{
	Name:   "resources",
	Usage:  "resources dump to directory",
	Action: actionDumpResource,
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:    "db",
			Usage:   "database file",
			Value:   "Backup.db",
			Aliases: []string{"d"},
		},
		&cli.StringFlag{
			Name:     "resource",
			Usage:    "BAK_0_XXX folder path",
			Required: true,
			Aliases:  []string{"r"},
		},
		&cli.StringFlag{
			Name:     "pass",
			Usage:    "decrypt media resource file chunk key",
			Required: true,
			Aliases:  []string{"p"},
		},
	},
}

func actionDumpResource(ctx *cli.Context) error {
	dbName := ctx.String("db")
	resource := ctx.String("resource")
	pass := Pass(ctx.String("pass"))
	if !pass.Valid() {
		return ErrInvalidPassKey
	}

	defer releaseFds()

	db, err := NewBackupDB(dbName)
	if err != nil {
		return err
	}
	defer db.Close()

	segments, err := db.FileSegments()
	if err != nil {
		return err
	}

	buf := bytes.NewBuffer([]byte{})

	var n int
	for i := 0; i < len(segments); i++ {
		segment := segments[i]
		fd := getFd(filepath.Join(resource, segment.FileName))
		if _, err = fd.Seek(segment.OffSet, io.SeekStart); err != nil {
			return err
		}

		data := make([]byte, segment.Length, segment.Length)

		if n, err = fd.Read(data); err != nil {
			return err
		}

		usePadding := true
		if i+1 < len(segments) {
			if segments[i+1].MapKey == segment.MapKey {
				usePadding = false
			}
		}

		if data, err = aesDecrypt([]byte(pass), data[:n], usePadding); err != nil {
			return err
		}

		buf.Write(data)

		if usePadding {
			totalBuf := buf.Bytes()
			ext := getExt(totalBuf)

			fi, err := os.Stat(segment.FileName)
			if os.IsNotExist(err) {
				if err = os.MkdirAll(segment.FileName, 0755); err != nil {
					return err
				}
			} else if err != nil {
				return err
			} else if !fi.IsDir() {
				return fmt.Errorf("%s is not dir", segment.FileName)
			}

			o, err := os.Create(filepath.Join(segment.FileName, strconv.Itoa(segment.MapKey)) + ext)
			if err != nil {
				return err
			}

			if _, err = o.Write(totalBuf); err != nil {
				o.Close()
				return err
			}
			o.Close()

			buf.Reset()
		}
	}

	return nil
}
