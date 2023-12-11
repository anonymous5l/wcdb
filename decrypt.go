package main

import (
	"crypto/aes"
	"github.com/urfave/cli/v2"
	"os"
)

var DecryptCommand = &cli.Command{
	Name:   "decrypt",
	Usage:  "decrypt raw file",
	Action: actionDecryptFile,
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:     "input",
			Usage:    "input filepath",
			Required: true,
			Aliases:  []string{"i"},
		},
		&cli.StringFlag{
			Name:    "output",
			Usage:   "output filepath",
			Aliases: []string{"o"},
		},
		&cli.StringFlag{
			Name:     "pass",
			Usage:    "decrypt media resource file chunk key",
			Required: true,
			Aliases:  []string{"p"},
		},
	},
}

func actionDecryptFile(ctx *cli.Context) error {
	input := ctx.String("input")
	output := ctx.String("output")
	pass := Pass(ctx.String("pass"))
	if !pass.Valid() {
		return ErrInvalidPassKey
	}

	if output == "" {
		output = input + "_Decrypt"
	}

	inputBytes, err := os.ReadFile(input)
	if err != nil {
		return err
	}

	cipher, err := aes.NewCipher([]byte(pass)[:16])
	if err != nil {
		return err
	}
	size := cipher.BlockSize()
	for bs, be := 0, size; bs < len(inputBytes); bs, be = bs+size, be+size {
		cipher.Decrypt(inputBytes[bs:be], inputBytes[bs:be])
	}

	if err = os.WriteFile(output, inputBytes, 0644); err != nil {
		return err
	}

	return nil
}
