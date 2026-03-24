package cli

import (
	"github.com/hsqbyte/protospec/src/i18n"
)

var bundle *i18n.Bundle

func init() {
	bundle = i18n.NewBundle()
	if err := bundle.LoadFS(i18n.CliFS, "cli"); err != nil {
		panic("failed to load i18n: " + err.Error())
	}
}

// msg returns the localized message for the given key.
func msg(ctx *Context, key string) string {
	return bundle.Get(ctx.Lang, key)
}
