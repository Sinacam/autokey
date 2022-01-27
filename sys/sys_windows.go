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
	KeyDown = iota + 1
	KeyUp
	LeftMouseDown
	LeftMouseUp
	RightMouseDown
	RightMouseUp
)

const (
	F1  = int(C.VK_F1)
	F2  = int(C.VK_F2)
	F3  = int(C.VK_F3)
	F4  = int(C.VK_F4)
	F5  = int(C.VK_F5)
	F6  = int(C.VK_F6)
	F7  = int(C.VK_F7)
	F8  = int(C.VK_F8)
	F9  = int(C.VK_F9)
	F10 = int(C.VK_F10)
	F11 = int(C.VK_F11)
	F12 = int(C.VK_F12)
	F13 = int(C.VK_F13)
	F14 = int(C.VK_F14)
	F15 = int(C.VK_F15)
	F16 = int(C.VK_F16)
	F17 = int(C.VK_F17)
	F18 = int(C.VK_F18)
	F19 = int(C.VK_F19)
	F20 = int(C.VK_F20)
	F21 = int(C.VK_F21)
	F22 = int(C.VK_F22)
	F23 = int(C.VK_F23)
	F24 = int(C.VK_F24)
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
