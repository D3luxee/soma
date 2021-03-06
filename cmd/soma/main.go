package main

import (
	"fmt"
	"os"

	"github.com/codegangsta/cli"
	"github.com/mjolnir42/soma/internal/db"
)

//go:generate go run ../../script/render_markdown.go ../../docs/soma/command_reference ../../internal/help/rendered
//go:generate go-bindata -pkg help -ignore .gitignore -o ../../internal/help/bindata.go -prefix "../../internal/help/rendered/" ../../internal/help/rendered/...

// Cfg is the configuration that is exported for unknown reasons XXX
var Cfg Config
var store db.DB
var somaVersion string

const rfc3339Milli string = `2006-01-02T15:04:05.000Z07:00`

func main() {
	cli.CommandHelpTemplate = `{{.Description}}`

	app := cli.NewApp()
	app.Name = `soma`
	app.Usage = `SOMA Administrative Interface`
	app.Version = somaVersion
	app.EnableBashCompletion = true

	app = registerCommands(*app)
	app = registerFlags(*app)

	if err := app.Run(os.Args); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

// vim: ts=4 sw=4 sts=4 noet fenc=utf-8 ffs=unix
