package yap

import (
	_ "embed"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"

	"github.com/ARC5RF/go-blame"
	"github.com/gorilla/websocket"
)

type no_space_after struct{ v any }

func NoSpaceAfter(v any) *no_space_after {
	return &no_space_after{v}
}

const (
	INSERT_SPACES_IN_WEBSOCKET int = 1 << iota
	INSERT_SPACES_IN_FILE
	INSERT_SPACES_IN_TERMINAL
)

//go:embed client.js
var ClientJS []byte

func ServeClientJS(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/javascript")
	w.Write(ClientJS)
}

type tokenize_for_websocket interface{ TokenizeForWebsocket() ([]any, error) }
type tokenize_for_terminal interface{ TokenizeForTerminal() ([]any, error) }
type tokenize_for_file interface{ TokenizeForFile() ([]any, error) }

type terminal_output struct {
	underlying                   io.Writer
	insert_spaces_between_tokens bool
}

type file_output struct {
	underlying                   io.Writer
	insert_spaces_between_tokens bool
}

type impl struct {
	File      file_output
	Terminal  terminal_output
	Websocket *websocket_controller
}

func New(logfile, terminal io.Writer, default_keep, poll_interval_ms int, flags int) *impl {
	controller := &websocket_controller{}
	controller.insert_spaces_between_tokens = flags&INSERT_SPACES_IN_WEBSOCKET == INSERT_SPACES_IN_WEBSOCKET
	controller.guard = &sync.Mutex{}
	controller.connections = map[*websocket.Conn]*websocket_client_data{}
	controller.cache = map[string]*receiver_cache{}
	controller.changes = map[string]int{}
	controller.priority = []string{}
	controller.default_keep = default_keep
	controller.interval = time.Millisecond * time.Duration(poll_interval_ms)
	controller.handlers = []*client_message_handler{}

	fout := file_output{logfile, flags&INSERT_SPACES_IN_FILE == INSERT_SPACES_IN_FILE}
	tout := terminal_output{terminal, flags&INSERT_SPACES_IN_TERMINAL == INSERT_SPACES_IN_TERMINAL}

	inst := &impl{fout, tout, controller}

	return inst
}

func (inst *terminal_output) Log(args []any) (int, error) {
	if inst.underlying == nil {
		return 0, nil
	}
	flattened, recursion_err := blame.O1(recursively_for_terminal(args, inst.insert_spaces_between_tokens))
	if recursion_err != nil {
		return 0, recursion_err.WithAdditionalContext("could not flatten args")
	}

	return fmt.Fprint(inst.underlying, flattened...)
}

func (inst *file_output) Log(args []any) (int, error) {
	if inst.underlying == nil {
		return 0, nil
	}
	flattened, recursion_err := blame.O1(recursively_for_file(args, inst.insert_spaces_between_tokens))
	if recursion_err != nil {
		return 0, recursion_err.WithAdditionalContext("could not flatten args")
	}
	return fmt.Fprint(inst.underlying, flattened...)
}

func (inst *impl) Log(args ...any) (int, error) {
	if e := blame.O0(inst.Websocket.Emit("yap.log", args...)); e != nil {
		return 0, e.WithAdditionalContext("could not forward to websocket")
	}
	if _, e := blame.O1(inst.File.Log(args)); e != nil {
		return 0, e.WithAdditionalContext("could not forward to file")
	}
	return inst.Terminal.Log(args)
}

func (inst *impl) LogLN(args ...any) (int, error) {
	return blame.O1(inst.Log(append(args, NL)...))
}

func (inst *impl) Debug(args ...any) (int, error) {
	upstream := blame.L(1).String()
	send_to_debug := append([]any{White(time.Now().Format(time.RFC3339Nano)), Magenta(upstream)}, args...)
	if e := blame.O0(inst.Websocket.Emit("yap.log", send_to_debug...)); e != nil {
		return 0, e.WithAdditionalContext("could not forward to websocket")
	}
	if _, e := blame.O1(inst.File.Log(send_to_debug)); e != nil {
		return 0, e.WithAdditionalContext("could not forward to file")
	}
	return inst.Terminal.Log(send_to_debug)
}
