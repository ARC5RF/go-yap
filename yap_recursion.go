package yap

import (
	"encoding/json"
	"fmt"
	"reflect"

	"github.com/ARC5RF/go-blame"
)

func recursively_for_terminal(inputs []any, spaces bool) ([]any, error) {
	output := []any{}
	for i, v := range inputs {
		if decorator, has_decorator := v.(tokenize_for_terminal); has_decorator {
			decorated, decoration_err := blame.O1(decorator.TokenizeForTerminal())
			if decoration_err != nil {
				return nil, decoration_err
			}

			children, recursion_err := blame.O1(recursively_for_terminal(decorated, spaces))
			if recursion_err != nil {
				return nil, recursion_err
			}

			output = append(output, children...)
			continue
		}
		s := fmt.Sprint(v)
		if len(s) > 0 {
			output = append(output, s)
			if spaces {
				if i < len(inputs)-1 {
					output = append(output, " ")
				}
			}
		}
	}
	return output, nil
}

func recursively_for_file(inputs []any, spaces bool) ([]any, error) {
	output := []any{}
	for i, v := range inputs {
		if decorator, has_decorator := v.(tokenize_for_file); has_decorator {
			decorated, decoration_err := blame.O1(decorator.TokenizeForFile())
			if decoration_err != nil {
				return nil, decoration_err
			}

			children, recursion_err := blame.O1(recursively_for_file(decorated, spaces))
			if recursion_err != nil {
				return nil, recursion_err
			}

			output = append(output, children...)
			continue
		}
		s := fmt.Sprint(v)
		if len(s) > 0 {
			output = append(output, s)
			if spaces {
				if i < len(inputs)-1 {
					output = append(output, " ")
				}
			}
		}
	}
	return output, nil
}

func recursively_for_websocket(inputs []any, spaces bool) ([]*token, error) {
	output := []*token{}
	for i, v := range inputs {
		if decorator, has_websocket_decorator := v.(tokenize_for_websocket); has_websocket_decorator {
			decorated, decoration_err := blame.O1(decorator.TokenizeForWebsocket())
			if decoration_err != nil {
				return nil, decoration_err
			}

			children, recursion_err := blame.O1(recursively_for_websocket(decorated, spaces))
			if recursion_err != nil {
				return nil, recursion_err
			}

			output = append(output, children...)
			continue
		}
		t := reflect.TypeOf(v).String()

		var s json.RawMessage
		if marshaler, is_marshaler := v.(json.Marshaler); is_marshaler {
			marshaled, marsh_err := blame.O1(marshaler.MarshalJSON())
			if marsh_err != nil {
				return nil, marsh_err.WithAdditionalContext("could not marshal something to json")
			}
			s = marshaled
		} else if stringer, is_stringer := v.(fmt.Stringer); is_stringer {
			// t = "string"
			vs := stringer.String()
			spvd, marsh_err := blame.O1(json.Marshal(vs))
			if marsh_err != nil {
				return nil, marsh_err
			}
			s = spvd
		} else {
			vs := fmt.Sprint(v)
			spvd, marsh_err := blame.O1(json.Marshal(vs))
			if marsh_err != nil {
				return nil, marsh_err
			}
			s = spvd
		}

		if len(s) > 0 {
			output = append(output, &token{T: t, V: s})
			if spaces {
				if i < len(inputs)-1 {
					output = append(output, &token{T: "string", V: " "})
				}
			}
		}
	}
	return output, nil
}
