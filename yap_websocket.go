package yap

import (
	"encoding/json"
	"slices"
	"strings"
	"sync"
	"time"

	"github.com/ARC5RF/go-blame"
	//TODO remove hard dependency on gorilla
	"github.com/gorilla/websocket"
)

type client_subscription struct {
	Reciever string
	Keep     int
	Cursor   int
}
type client_subscription_collection []*client_subscription

func (subscriptions client_subscription_collection) Contains(reciever string) bool {
	return slices.ContainsFunc(subscriptions, func(sub *client_subscription) bool {
		return sub.Reciever == reciever
	})
}

func (subscriptions client_subscription_collection) Lookup(reciever string) (*client_subscription, bool) {
	for _, v := range subscriptions {
		if v.Reciever == reciever {
			return v, true
		}
	}
	return nil, false
}

func client_subscriptions_from_strings(input []string, default_keep int) client_subscription_collection {
	output := client_subscription_collection{}

	for _, v := range input {
		output = append(output, &client_subscription{v, default_keep, 0})
	}

	return output
}

type websocket_client_data struct {
	guard         *sync.Mutex
	subscriptions client_subscription_collection
}

type receiver_cache struct {
	entries [][]byte
	keep    int
}

type websocket_controller struct {
	guard                        *sync.Mutex
	connections                  map[*websocket.Conn]*websocket_client_data
	cache                        map[string]*receiver_cache
	changes                      map[string]int
	priority                     []string
	log                          func(args ...any) (int, error)
	default_keep                 int
	insert_spaces_between_tokens bool
	interval                     time.Duration
	handlers                     []*client_message_handler
}

func (controller *websocket_controller) snapshot_cache_filtered(subscriptions client_subscription_collection) map[string]*receiver_cache {
	controller.guard.Lock()
	defer controller.guard.Unlock()

	outupt := map[string]*receiver_cache{}
	for Receiver, cache := range controller.cache {
		if !subscriptions.Contains(Receiver) {
			continue
		}
		//this is a shallow copy with side effects for performance
		//it is known that no operation using this data mutates it
		outupt[Receiver] = cache
	}

	return outupt
}

func (controller *websocket_controller) read_forever(c *websocket.Conn, c_data *websocket_client_data, read_error chan error) {
	for {
		_, message, ws_read_err := blame.O2(c.ReadMessage())
		if ws_read_err != nil {
			if strings.Contains(ws_read_err.Error(), "1001 (going away)") {
				read_error <- nil
				return
			}
			read_error <- ws_read_err.WithAdditionalContext("websocket failed to read")
			return
		}

		temp := message_from_websocket{}
		marsh_err := blame.O0(json.Unmarshal(message, &temp))
		if marsh_err != nil {
			read_error <- marsh_err
			return
		}

		switch temp.Receiver {
		case "yap.subscribe":
			client_wants := &client_subscription{}
			umarsh_err := blame.O0(json.Unmarshal(temp.Data, client_wants))
			if umarsh_err != nil {
				c_data.guard.Unlock()
				read_error <- umarsh_err
				return
			}

			controller.guard.Lock()
			if !slices.Contains(controller.priority, client_wants.Reciever) {
				controller.priority = append(controller.priority, client_wants.Reciever)
			}
			controller.guard.Unlock()

			c_data.guard.Lock()
			if !c_data.subscriptions.Contains(client_wants.Reciever) {
				c_data.subscriptions = append(c_data.subscriptions, client_wants)
			}
			c_data.guard.Unlock()
			continue
		}

		controller.guard.Lock()
		h_snapshot := append([]*client_message_handler{}, controller.handlers...)
		controller.guard.Unlock()

		for _, h := range h_snapshot {
			if h.receiver == temp.Receiver {
				if e := blame.O0(h.callback(temp.Data)); e != nil {
					read_error <- e.WithAdditionalContext("userspace returned an error")
					return
				}
			}
		}

		<-time.After(controller.interval)
	}
}

func (controller *websocket_controller) write_forever(c *websocket.Conn, c_data *websocket_client_data, read_error chan error) error {
	for {
		c_data.guard.Lock()
		snapshot := controller.snapshot_cache_filtered(c_data.subscriptions)
		c_data.guard.Unlock()
		for _, k := range controller.priority {
			c_data.guard.Lock()
			sub, has_idx := c_data.subscriptions.Lookup(k)
			c_data.guard.Unlock()
			if !has_idx {
				continue
			}

			r, has_r := snapshot[k]
			if !has_r {
				continue
			}

			c_data.guard.Lock()
			if len(r.entries) == 0 {
				c_data.guard.Unlock()
				continue
			}
			r_snapshot := append([][]byte{}, r.entries...)
			r_len := len(r_snapshot)
			c_data.guard.Unlock()

			controller.guard.Lock()
			changes, has_changes := controller.changes[k]
			if !has_changes {
				controller.changes[k] = 0
			}
			controller.guard.Unlock()

			for changes > sub.Cursor {
				missing := changes - sub.Cursor
				jump_to := r_len - missing
				if jump_to >= 0 {
					data := r_snapshot[jump_to]
					err := blame.O0(c.WriteMessage(websocket.TextMessage, data))
					if err != nil {
						return err.WithAdditionalContext("error while writing message")
					}
				}
				sub.Cursor++
			}
		}
		select {
		case err := <-read_error:
			return err
		case <-time.After(controller.interval):
		}
	}
}

func (controller *websocket_controller) Add(c *websocket.Conn, subscriptions ...string) error {
	// snapshot := hub.snapshot_cache_filtered(subscriptions)
	controller.guard.Lock()
	c_data := &websocket_client_data{&sync.Mutex{}, client_subscriptions_from_strings(subscriptions, controller.default_keep)}
	controller.connections[c] = c_data
	controller.guard.Unlock()

	read_error := make(chan error)
	go controller.read_forever(c, c_data, read_error)
	return blame.O0(controller.write_forever(c, c_data, read_error))
}

func (controller *websocket_controller) Rem(c *websocket.Conn) {
	controller.guard.Lock()
	defer controller.guard.Unlock()
	delete(controller.connections, c)
}

func (controller *websocket_controller) lookup(receiver string) *receiver_cache {
	controller.guard.Lock()
	defer controller.guard.Unlock()
	r, has_r := controller.cache[receiver]
	if !has_r {
		r = &receiver_cache{}
		r.keep = controller.default_keep
		if !slices.Contains(controller.priority, receiver) {
			controller.priority = append(controller.priority, receiver)
		}
		controller.cache[receiver] = r
	}
	return r
}

func (controller *websocket_controller) Keep(receiver string, amount int) {
	r := controller.lookup(receiver)
	controller.guard.Lock()
	r.keep = amount
	controller.guard.Unlock()
}

func (controller *websocket_controller) Purge(receivers ...string) {
	for _, receiver := range receivers {
		r := controller.lookup(receiver)
		controller.guard.Lock()
		r.entries = make([][]byte, 0)
		for _, v := range controller.connections {
			v.guard.Lock()
			sub, has_index := v.subscriptions.Lookup(receiver)
			if has_index {
				sub.Cursor = 0
			}
			v.guard.Unlock()
		}
		controller.guard.Unlock()
	}
}

func (controller *websocket_controller) Emit(receiver string, args ...any) error {
	if controller == nil {
		return nil
	}

	flattened, recursion_err := blame.O1(recursively_for_websocket(args, controller.insert_spaces_between_tokens))
	if recursion_err != nil {
		return recursion_err.WithAdditionalContext("could not flatten args")
	}

	data, marsh_err := blame.O1(json.Marshal(message_to_websocket{receiver, flattened}))
	if marsh_err != nil {
		return marsh_err.WithAdditionalContext("could not marshal message for websocket")
	}

	r := controller.lookup(receiver)
	controller.guard.Lock()
	if len(r.entries) >= r.keep {
		r.entries = append([][]byte{}, r.entries[1:]...)
	}
	r.entries = append(r.entries, data)
	_, has_changes := controller.changes[receiver]
	if !has_changes {
		controller.changes[receiver] = 0
	}
	controller.changes[receiver]++
	controller.guard.Unlock()

	return nil
}

func (controller *websocket_controller) On(receiver string, callback func([]byte) error) {
	controller.guard.Lock()
	controller.handlers = append(controller.handlers, &client_message_handler{receiver: receiver, callback: callback})
	controller.guard.Unlock()
}

type token struct {
	T string
	V any
}

type message_to_websocket struct {
	Receiver string
	Tokens   []*token
}

type message_from_websocket struct {
	Receiver string
	Data     json.RawMessage
}

type client_message_handler struct {
	receiver string
	callback func(data []byte) error
}
