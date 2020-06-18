package main

import "gopkg.in/src-d/go-cli.v0"

var (
	name    = "gopherbbs"
	version = "undefined"
	build   = "undefined"
)

var app = cli.New(name, version, build, "A BBS system in Go")

func main() {
	app.RunMain()
}
