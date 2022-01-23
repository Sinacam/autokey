package autokey

/*
#include "sys_windows.hpp"
*/
import "C"
import "unsafe"

const (
	KeyDown       = 0
	KeyUp         = uint(C.KEYEVENTF_KEYUP)
	LeftMouseDown = uint(C.MOUSEEVENTF_LEFTDOWN)
	LeftMouseUp   = uint(C.MOUSEEVENTF_LEFTUP)

	Alt  = int(C.VK_MENU)
	Ctrl = int(C.VK_CONTROL)
)

func KeybdEvent(k int, flag uint) {
	C.keybd_event(C.BYTE(k), 0, C.DWORD(flag), 0)
}

func GetClipboardText() string {
	cstr := C.GetClipboardText()
	defer C.free(unsafe.Pointer(cstr))
	return C.GoString(cstr)
}

func MouseEvent(flag uint) {
	C.mouse_event(C.DWORD(flag), 0, 0, 0, 0)
}

func SetGlobalHook() uintptr {
	return uintptr(unsafe.Pointer(C.SetGlobalHook()))
}

func Unhook(hook uintptr) {
	C.UnhookWindowsHookEx(C.HHOOK(unsafe.Pointer(hook)))
}
