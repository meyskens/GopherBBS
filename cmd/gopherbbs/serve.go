package main

import (
	"context"
	"errors"

	"github.com/meyskens/GopherBBS/pkg/gopherbbs"
	"github.com/meyskens/GopherBBS/pkg/modem"
	gocli "gopkg.in/src-d/go-cli.v0"
)

func init() {
	app.AddCommand(&ServeCommand{})
}

type ServeCommand struct {
	gocli.PlainCommand `name:"serve" short-description:"serves a BBS on a serial line"`
	Port               string `long:"port" env:"GOPHERBBS_PORT" description:"Serial port of the modem to run the BBS on"`
	Baudrate           uint   `long:"baudrate" env:"GOPHERBBS_BAUDRATE" description:"Serial port baudrate" default:"115200"`
}

func (r *ServeCommand) ExecuteContext(ctx context.Context, args []string) error {
	if r.Port == "" {
		return errors.New("Port is not defined")
	}

	bbs, err := gopherbbs.NewBBS()
	if err != nil {
		return err
	}
	line, err := modem.NewModem(r.Port, r.Baudrate)
	if err != nil {
		return err
	}
	bbs.AddLine(line)

	bbs.Start(ctx)

	return nil
}
