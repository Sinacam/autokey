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

	// Mouse clicks are represented as keys above the keyboard key range.
	LeftMouse = int(iota + 256)
	RightMouse
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

	Alt        = int(C.VK_MENU)
	Ctrl       = int(C.VK_CONTROL)
	LeftCtrl   = int(C.VK_LCONTROL)
	RightCtrl  = int(C.VK_RCONTROL)
	Shift      = int(C.VK_SHIFT)
	LeftShift  = int(C.VK_LSHIFT)
	RightShift = int(C.VK_RSHIFT)
	Enter      = int(C.VK_RETURN)
	Esc        = int(C.VK_ESCAPE)
	Space      = int(C.VK_SPACE)
	Left       = int(C.VK_LEFT)
	Up         = int(C.VK_UP)
	Right      = int(C.VK_RIGHT)
	Down       = int(C.VK_DOWN)
	End        = int(C.VK_END)
	Home       = int(C.VK_HOME)
	Delete     = int(C.VK_DELETE)

	Num0 = int(C.VK_NUMPAD0)
	Num1 = int(C.VK_NUMPAD1)
	Num2 = int(C.VK_NUMPAD2)
	Num3 = int(C.VK_NUMPAD3)
	Num4 = int(C.VK_NUMPAD4)
	Num5 = int(C.VK_NUMPAD5)
	Num6 = int(C.VK_NUMPAD6)
	Num7 = int(C.VK_NUMPAD7)
	Num8 = int(C.VK_NUMPAD8)
	Num9 = int(C.VK_NUMPAD9)
)

var (
	InvalidFlag = errors.New("invalid flag")
)

func Send(k int, flag uint64) error {
	if k < 256 {
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

	var arg C.DWORD
	switch {
	case k == LeftMouse && flag == KeyDown:
		arg = C.MOUSEEVENTF_LEFTDOWN
	case k == LeftMouse && flag == KeyUp:
		arg = C.MOUSEEVENTF_LEFTUP
	case k == RightMouse && flag == KeyDown:
		arg = C.MOUSEEVENTF_RIGHTDOWN
	case k == RightMouse && flag == KeyUp:
		arg = C.MOUSEEVENTF_RIGHTUP
	default:
		return InvalidFlag
	}
	C.mouse_event(arg, 0, 0, 0, 0)
	return nil
}

func GetClipboardText() string {
	cstr := C.getClipboardText()
	defer C.free(unsafe.Pointer(cstr))
	return C.GoString(cstr)
}

func SetGlobalHook() {
	C.setGlobalHook()
}

func Unhook() {
	C.unhook()
}

func GetInput() (int, uint64) {
	input := C.getInput()
	var key int
	var flag uint64
	switch input.flag {
	case C.WM_KEYDOWN:
		key = int(input.key)
		flag = KeyDown
	case C.WM_KEYUP:
		key = int(input.key)
		flag = KeyUp
	case C.WM_LBUTTONDOWN:
		key = LeftMouse
		flag = KeyDown
	case C.WM_LBUTTONUP:
		key = LeftMouse
		flag = KeyUp
	case C.WM_RBUTTONDOWN:
		key = RightMouse
		flag = KeyDown
	case C.WM_RBUTTONUP:
		key = RightMouse
		flag = KeyUp
	default:
		key = int(input.key)
		flag = uint64(input.flag)
	}
	return key, uint64(flag)
}
