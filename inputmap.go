package autokey

import (
	"fmt"
	"strings"
)

var (
	inputMap = makeInputMap()
)

// makeInputMap makes the map from string to Input.
// Includes alphanumeric, f1~f12, special keys (e.g. ctrl) and mouse inputs.
// If there are suffixes "down" and "up", Input.Flag is KeyDown and KeyUp,
// otherwise it's 0.
// Mouse input is an exception because it is only a flag.
func makeInputMap() map[string]Input {
	m := make(map[string]Input)
	for c := '0'; c <= '9'; c++ {
		m[string(c)] = Input{Key: int(c)}
	}
	for c := 'a'; c <= 'z'; c++ {
		m[string(c)] = Input{Key: int(c) - 'a' + 'A'}
	}
	for i := 1; i <= 12; i++ {
		m[fmt.Sprintf("f%v", i)] = Input{Key: F1 + i - 1}
	}

	str := []string{
		"Alt",
		"Ctrl",
		"LeftCtrl",
		"RightCtrl",
		"Shift",
		"LeftShift",
		"RightShift",
		"Enter",
		"Esc",
		"Space",
		"Left",
		"Up",
		"Right",
		"Down",
		"End",
		"Home",
		"Delete",
	}
	val := []int{
		Alt,
		Ctrl,
		LeftCtrl,
		RightCtrl,
		Shift,
		LeftShift,
		RightShift,
		Enter,
		Esc,
		Space,
		Left,
		Up,
		Right,
		Down,
		End,
		Home,
		Delete,
	}
	for i := range str {
		m[varToColloquial(str[i])] = Input{Key: val[i]}
	}

	for k, v := range m {
		v.Flag = KeyDown
		m[k+" down"] = v
		v.Flag = KeyUp
		m[k+" up"] = v
	}

	m["left click"] = Input{Flag: LeftMouseDown}
	m["left click up"] = Input{Flag: LeftMouseUp}
	m["right click"] = Input{Flag: RightMouseDown}
	m["right click up"] = Input{Flag: RightMouseUp}
	return m
}

// varToColloquial transforms camel and pascal case to space seperated words.
func varToColloquial(s string) string {
	ret := []byte{s[0]}
	for _, v := range []byte(s)[1:] {
		if v >= 'A' && v <= 'Z' {
			ret = append(ret, ' ')
		}
		ret = append(ret, v)
	}
	return strings.ToLower(string(ret))
}
