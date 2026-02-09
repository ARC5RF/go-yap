package yap

import "fmt"

type sized struct {
	key string
	v   any
}

func (c *sized) String() string { return "" }

func (c *sized) TokenizeForTerminal() ([]any, error) {
	return []any{c.v}, nil
}
func (c *sized) TokenizeForFile() ([]any, error) {
	return []any{c.v}, nil
}
func (c *sized) TokenizeForWebsocket() ([]any, error) {
	// id, id_err := blame.O1(factory.Next())
	// if id_err != nil {
	// 	return []any{}, id_err.WithAdditionalContext("could not generate id for html pair")
	// }
	return []any{HTML_PART(`<h` + c.key + `>`), c.v, HTML_PART(`</h` + c.key + `>`)}, nil
}

func Huge(k int, v any) *sized { return &sized{fmt.Sprint(k), v} }
