package cmpl

import "github.com/codegangsta/cli"

func PropertyAdd(c *cli.Context) {
	genericPropertyAdd(c, []string{})
}

func PropertyAddValue(c *cli.Context) {
	genericPropertyAdd(c, []string{`value`})
}

func genericPropertyAdd(c *cli.Context, args []string) {
	Generic(c, append([]string{`to`, `in`, `value`, `view`, `inheritance`, `childrenonly`}, args...))
}

func PropertyCreate(c *cli.Context) {
	Generic(c, []string{`on`, `view`, `inheritance`, `childrenonly`})
}

func PropertyCreateIn(c *cli.Context) {
	Generic(c, []string{`on`, `in`, `view`, `inheritance`, `childrenonly`})
}

func PropertyCreateValue(c *cli.Context) {
	Generic(c, []string{`on`, `value`, `view`, `inheritance`, `childrenonly`})
}

func PropertyCreateInValue(c *cli.Context) {
	Generic(c, []string{`on`, `in`, `value`, `view`, `inheritance`, `childrenonly`})
}

func PropertyOnView(c *cli.Context) {
	Generic(c, []string{`on`, `view`})
}

func PropertyOnInView(c *cli.Context) {
	Generic(c, []string{`on`, `in`, `view`})
}

// vim: ts=4 sw=4 sts=4 noet fenc=utf-8 ffs=unix
