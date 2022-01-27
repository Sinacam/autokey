package main

import (
	"errors"
	"fmt"
	"strings"

	"github.com/Sinacam/autokey"
	"gopkg.in/yaml.v2"
)

// ymlErrorString formats yml as an error string.
func ymlErrorString(yml interface{}) string {
	str, err := yaml.Marshal(yml)
	if err != nil {
		return "bad yml argument"
	}
	return fmt.Sprint("unrecognized yaml element ", string(str))
}

// addErrorTrace adds a trace string to err.
func addErrorTrace(err string, from interface{}) string {
	return fmt.Sprintf("%v\n\tfrom %v", err, from)
}

// Fn is the result of a compiled yml config.
// Fn may be used to only return a value or it may have further side effects.
// If the object is a literal, the return value is the object itself.
// If the object is an array, the return value is the slice of return values.
// If the object is a map, the return value depends on the content.
// All values are computed lazily on execution.
type Fn func() interface{}

// Compiles yml as a Fn recursively.
// Errors in the structure of yml is reported as an error.
// Errors in values causes a panic during execution of Fn instead.
func Compile(yml interface{}) (Fn, error) {
	fn, err := compile(yml)
	if err != "" {
		return nil, errors.New(err)
	}
	return fn, nil
}

// compile uses an error string because the error trace is built up
// during recursion.
func compile(yml interface{}) (Fn, string) {
	switch yml := yml.(type) {
	case bool:
		return Fn(func() interface{} {
			return yml
		}), ""
	case int:
		return Fn(func() interface{} {
			return yml
		}), ""
	case float64:
		return Fn(func() interface{} {
			return yml
		}), ""
	case string:
		return Fn(func() interface{} {
			return yml
		}), ""
	case []interface{}:
		return compileSlice(yml)
	case map[interface{}]interface{}:
		return compileMap(yml)
	}
	return nil, ymlErrorString(yml)
}

func compileSlice(yml []interface{}) (Fn, string) {
	var subfns []Fn
	for i, v := range yml {
		fn, err := compile(v)
		if err != "" {
			return nil, addErrorTrace(err, i)
		}
		subfns = append(subfns, fn)
	}

	return Fn(func() interface{} {
		var ret []interface{}
		for _, fn := range subfns {
			ret = append(ret, fn)
		}
		return ret
	}), ""
}

func compileMap(yml map[interface{}]interface{}) (Fn, string) {
	var fns []Fn
	for k, v := range yml {
		kstr, ok := k.(string)
		if !ok {
			return nil, "key must be a string"
		}

		// TODO: refactor switch if they end up being identical
		switch kstr {
		case "do":
			fn, err := compileDo(v)
			if err != "" {
				return nil, addErrorTrace(err, kstr)
			}
			fns = append(fns, fn)
		case "repeat":
			fn, err := compileRepeat(v)
			if err != "" {
				return nil, addErrorTrace(err, kstr)
			}
			fns = append(fns, fn)
		case "press":
			fn, err := compilePress(v)
			if err != "" {
				return nil, addErrorTrace(err, kstr)
			}
			fns = append(fns, fn)
		case "hold":
			fn, err := compileHold(v)
			if err != "" {
				return nil, addErrorTrace(err, kstr)
			}
			fns = append(fns, fn)
		case "file":
			fn, err := compileFile(v)
			if err != "" {
				return nil, addErrorTrace(err, kstr)
			}
			fns = append(fns, fn)
		default:
			return nil, "invalid key " + kstr
		}
	}

	return Fn(func() interface{} {
		for _, fn := range fns {
			fn()
		}
		return nil
	}), ""
}

var (
	inputMap = makeInputMap()
)

// makeInputMap makes the map from string to Input.
// Includes alphanumeric, f1~f12 and mouse inputs.
// If there are suffixes "down" and "up", Input.Flag is KeyDown and KeyUp,
// otherwise it's 0.
// Mouse input is an exception because it is only a flag.
func makeInputMap() map[string]autokey.Input {
	m := make(map[string]autokey.Input)
	for c := '0'; c <= '9'; c++ {
		m[string(c)] = autokey.Input{Key: int(c)}
	}
	for c := 'A'; c <= 'Z'; c++ {
		m[string(c)] = autokey.Input{Key: int(c)}
	}
	for i := 1; i <= 12; i++ {
		m[fmt.Sprintf("F%v", i)] = autokey.Input{Key: autokey.F1 + i - 1}
	}

	for k, v := range m {
		v.Flag = autokey.KeyDown
		m[k+" down"] = v
		v.Flag = autokey.KeyUp
		m[k+" up"] = v
	}

	m["left click"] = autokey.Input{Flag: autokey.LeftMouseDown}
	m["left click up"] = autokey.Input{Flag: autokey.LeftMouseUp}
	m["right click"] = autokey.Input{Flag: autokey.RightMouseDown}
	m["right click up"] = autokey.Input{Flag: autokey.RightMouseUp}
	return m
}

// parseInput parses val as a slice of Inputs.
// Accepts int, string or a slice of them.
func parseInput(val interface{}) ([]autokey.Input, error) {
	switch val := val.(type) {
	case int:
		return []autokey.Input{
			{Key: val + '0'},
		}, nil
	case string:
		val = strings.ToUpper(val)
		input, ok := inputMap[val]
		if !ok {
			break
		}
		return []autokey.Input{input}, nil
	case []interface{}:
		var inputs []autokey.Input
		for _, v := range val {
			input, err := parseInput(v)
			if err != nil {
				return nil, err
			}
			inputs = append(inputs, input...)
		}
		return inputs, nil
	}
	return nil, errors.New("cannot parse as Input")
}

// compileDo compiles the map value with key "do".
// Compiles by special-casing the "on" key as the trigger
// and delegating to compileMap for the remaining.
func compileDo(yml interface{}) (Fn, string) {
	m, ok := yml.(map[interface{}]interface{})
	if !ok {
		return nil, "value must be a map"
	}

	var on Fn
	remaining := make(map[interface{}]interface{})
	for k, v := range m {
		var kstr string
		switch k.(type) {
		case string:
			kstr = k.(string)
		case bool:
			if k.(bool) == true {
				// on gets parsed to true, we pretend that doesn't happen
				kstr = "on"
			} else {
				return nil, "key must be a string"
			}
		default:
			return nil, "key must be a string"
		}

		switch kstr {
		case "on":
			fn, err := compile(v)
			if err != "" {
				return nil, addErrorTrace(err, kstr)
			}
			on = fn
		default:
			remaining[k] = v
		}
	}

	remainingFn, err := compileMap(remaining)
	if err != "" {
		return nil, err
	}

	// If there is no trigger, do is a no-op.
	if on == nil {
		return remainingFn, ""
	}

	return Fn(func() interface{} {
		inputs, err := parseInput(on())
		if err != nil {
			panic("bad value for on")
		}

		for i := range inputs {
			if inputs[i].Flag == 0 {
				inputs[i].Flag = autokey.KeyDown
			}
		}

		ch := make(chan autokey.Input)
		autokey.NotifyOn(ch, inputs...)
		go func() {
			for range ch {
				remainingFn()
			}
		}()

		return nil
	}), ""
}

func compileRepeat(yml interface{}) (Fn, string) {
	return nil, ymlErrorString(yml)
}

func compilePress(yml interface{}) (Fn, string) {
	fn, err := compile(yml)
	if err != "" {
		return nil, err
	}
	return Fn(func() interface{} {
		inputs, err := parseInput(fn())
		if err != nil {
			panic("bad value for press")
		}

		for _, input := range inputs {
			if input.Flag == 0 {
				input.Flag = autokey.KeyDown
			}
			autokey.Send(input)
		}
		for _, input := range inputs {
			if input.Flag == 0 || input.Flag == autokey.KeyDown {
				input.Flag = autokey.KeyUp
				autokey.Send(input)
			}
		}

		return nil
	}), ""
}

func compileHold(yml interface{}) (Fn, string) {
	fn, err := compile(yml)
	if err != "" {
		return nil, err
	}
	return Fn(func() interface{} {
		inputs, err := parseInput(fn())
		if err != nil {
			panic("bad value for press")
		}
		for _, input := range inputs {
			autokey.Send(input)
		}
		return nil
	}), ""
}

func compileFile(yml interface{}) (Fn, string) {
	return nil, ymlErrorString(yml)
}