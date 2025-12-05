package yap

import (
	"fmt"
	"io"
	"sync"
	"time"

	"github.com/ARC5RF/go-blame"
	"github.com/gorilla/websocket"
)

type Output = io.Writer

type tokenize_for_websocket interface{ TokenizeForWebsocket() ([]any, error) }
type tokenize_for_terminal interface{ TokenizeForTerminal() ([]any, error) }
type tokenize_for_file interface{ TokenizeForFile() ([]any, error) }

type Wrapper interface {
	Log(args ...any) (int, error)
}

type impl struct {
	logfile                      Output
	terminal                     Output
	Websocket                    *websocket_controller
	insert_spaces_between_tokens bool
}

func New(logfile, terminal Output, default_keep, poll_interval_ms int, insert_spaces, insert_spaces_in_websocket bool) *impl {
	controller := &websocket_controller{}
	controller.insert_spaces_between_tokens = insert_spaces_in_websocket
	controller.guard = &sync.Mutex{}
	controller.connections = map[*websocket.Conn]*websocket_client_data{}
	controller.cache = map[string]*receiver_cache{}
	controller.changes = map[string]int{}
	controller.priority = []string{}
	controller.default_keep = default_keep
	controller.interval = time.Millisecond * time.Duration(poll_interval_ms)
	controller.handlers = []*client_message_handler{}
	inst := &impl{logfile, terminal, controller, insert_spaces}

	return inst
}

func (inst *impl) forward_to_terminal(args []any) (int, error) {
	if inst.terminal == nil {
		return 0, nil
	}
	flattened, recursion_err := blame.O1(recursively_for_terminal(args, inst.Websocket.insert_spaces_between_tokens))
	if recursion_err != nil {
		return 0, recursion_err.WithAdditionalContext("could not flatten args")
	}

	return fmt.Fprint(inst.terminal, flattened...)
}

func (inst *impl) forward_to_file(args []any) (int, error) {
	if inst.logfile == nil {
		return 0, nil
	}
	flattened, recursion_err := blame.O1(recursively_for_file(args, inst.Websocket.insert_spaces_between_tokens))
	if recursion_err != nil {
		return 0, recursion_err.WithAdditionalContext("could not flatten args")
	}
	return fmt.Fprint(inst.logfile, flattened...)
}

func (inst *impl) Log(args ...any) (int, error) {
	if e := blame.O0(inst.Websocket.Emit("yap.log", args...)); e != nil {
		return 0, e.WithAdditionalContext("could not forward to websocket")
	}
	if _, e := blame.O1(inst.forward_to_file(args)); e != nil {
		return 0, e.WithAdditionalContext("could not forward to file")
	}
	return inst.forward_to_terminal(args)
}
