package sys

/*
#cgo CXXFLAGS: -std=c++2a
#include "sys_windows.h"
*/
import "C"
import (
	"errors"
	"unsafe"
)

const (
	// These constants are mapped to their corresponding C constants internally.
	// The constant for the messages and event functions are _not_ the same,
	// so some mapping has to happen regardless.
	KeyDown = iota
	KeyUp
	LeftMouseDown
	LeftMouseUp
	RightMouseDown
	RightMouseUp
)

var (
	InvalidFlag = errors.New("invalid flag")
)

func KeybdEvent(k int, flag uint64) error {
	var arg C.DWORD
	switch flag {
	case KeyDown:
		arg = 0
	case KeyUp:
		arg = C.KEYEVENTF_KEYUP
	default:
		return InvalidFlag
	}
	C.keybd_event(C.BYTE(k), 0, arg, 0)
	return nil
}

func GetClipboardText() string {
	cstr := C.getClipboardText()
	defer C.free(unsafe.Pointer(cstr))
	return C.GoString(cstr)
}

func MouseEvent(flag uint64) error {
	var arg C.DWORD
	switch flag {
	case LeftMouseDown:
		arg = C.MOUSEEVENTF_LEFTDOWN
	case LeftMouseUp:
		arg = C.MOUSEEVENTF_LEFTUP
	case RightMouseDown:
		arg = C.MOUSEEVENTF_RIGHTDOWN
	case RightMouseUp:
		arg = C.MOUSEEVENTF_RIGHTUP
	default:
		return InvalidFlag
	}
	C.mouse_event(arg, 0, 0, 0, 0)
	return nil
}

func SetGlobalHook() {
	C.setGlobalHook()
}

func Unhook() {
	C.unhook()
}

func GetInput() (int, uint64) {
	input := C.getInput()
	var flag uint64
	switch input.flag {
	case C.WM_KEYDOWN:
		flag = KeyDown
	case C.WM_KEYUP:
		flag = KeyUp
	default:
		flag = uint64(input.flag)
	}
	return int(input.key), uint64(flag)
}
