package yap

import (
	"encoding/json"

	"github.com/ARC5RF/go-blame"
	"github.com/ARC5RF/go-ulid"
)

type HTML_SHORT string
type HTML_PART string
type HTML_PAIR struct {
	Data    string
	Sibling ulid.ULID
}

type html_pair_placeholder struct {
	Data    string
	Sibling ulid.ULID
}

func (pair HTML_PAIR) MarshalJSON() ([]byte, error) {
	return blame.O1(json.Marshal(html_pair_placeholder{pair.Data, pair.Sibling}))
}

type shared_newline struct{}

func (_ *shared_newline) TokenizeForTerminal() ([]any, error) {
	return []any{"\n"}, nil
}
func (_ *shared_newline) TokenizeForFile() ([]any, error) {
	return []any{"\n"}, nil
}
func (_ *shared_newline) TokenizeForWebsocket() ([]any, error) {
	return []any{HTML_SHORT("<br/>")}, nil
}

var NL = &shared_newline{}
