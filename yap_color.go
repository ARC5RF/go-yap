package yap

import (
	"github.com/ARC5RF/go-ulid"
)

var color_reset_for_term = "\033[0m"
var color_lut = map[string]string{
	"red":     "\033[31m",
	"green":   "\033[32m",
	"yellow":  "\033[33m",
	"blue":    "\033[34m",
	"magenta": "\033[35m",
	"cyan":    "\033[36m",
	"gray":    "\033[37m",
	"white":   "\033[97m",
}
var factory = ulid.NewFactory()

type colored struct {
	key string
	v   any
}

func (c *colored) String() string { return "" }

func (c *colored) TokenizeForTerminal() ([]any, error) {
	return []any{color_lut[c.key], c.v, color_reset_for_term}, nil
}
func (c *colored) TokenizeForFile() ([]any, error) {
	return []any{c.v}, nil
}
func (c *colored) TokenizeForWebsocket() ([]any, error) {
	// id, id_err := blame.O1(factory.Next())
	// if id_err != nil {
	// 	return []any{}, id_err.WithAdditionalContext("could not generate id for html pair")
	// }
	return []any{HTML_PART(`<p style="color:red">`), c.v, HTML_PART(`</p>`)}, nil
}

func Red(input any) *colored {
	return &colored{"red", input}
}
