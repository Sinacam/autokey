package autokey

import (
	"strings"
	"sync"

	"github.com/Sinacam/autokey/sys"
)

const (
	KeyDown = sys.KeyDown
	KeyUp   = sys.KeyUp
)

const (
	LeftClick  = sys.LeftMouse
	RightClick = sys.RightMouse
	F1         = sys.F1
	F2         = sys.F2
	F3         = sys.F3
	F4         = sys.F4
	F5         = sys.F5
	F6         = sys.F6
	F7         = sys.F7
	F8         = sys.F8
	F9         = sys.F9
	F10        = sys.F10
	F11        = sys.F11
	F12        = sys.F12
	Alt        = sys.Alt
	Ctrl       = sys.Ctrl
	LeftCtrl   = sys.LeftCtrl
	RightCtrl  = sys.RightCtrl
	Shift      = sys.Shift
	LeftShift  = sys.LeftShift
	RightShift = sys.RightShift
	Enter      = sys.Enter
	Esc        = sys.Esc
	Space      = sys.Space
	Left       = sys.Left
	Up         = sys.Up
	Right      = sys.Right
	Down       = sys.Down
	End        = sys.End
	Home       = sys.Home
	Delete     = sys.Delete
	Num0       = sys.Num0
	Num1       = sys.Num1
	Num2       = sys.Num2
	Num3       = sys.Num3
	Num4       = sys.Num4
	Num5       = sys.Num5
	Num6       = sys.Num6
	Num7       = sys.Num7
	Num8       = sys.Num8
	Num9       = sys.Num9
)

var (
	im = newinputMonitor()

	InvalidFlag = sys.InvalidFlag
)

type Input struct {
	Key  int
	Flag uint64
}

func (input Input) asMapKey() uint64 {
	return uint64(input.Key)<<32 | input.Flag
}

// Keys is a convenience function converting an alphanumeric string
// to a slice of Inputs corresponding to those characters.
// Non-alphanumeric characters are treated as their byte values.
func Keys(s string) []Input {
	var ret []Input
	for _, v := range []byte(strings.ToUpper(s)) {
		ret = append(ret, Input{Key: int(v)})
	}
	return ret
}

type inputMonitor struct {
	hookChan chan uintptr
	done     chan struct{}

	notifyOn  map[uint64][]chan<- Input
	notify    []chan<- Input
	notifyMtx sync.RWMutex
}

func newinputMonitor() *inputMonitor {
	return &inputMonitor{
		notifyOn: make(map[uint64][]chan<- Input),
		done:     make(chan struct{}),
	}
}

func (im *inputMonitor) Init() {
	go sys.SetGlobalHook()
	go func() {
		for {
			k, f := sys.GetInput()

			select {
			case <-im.done:
				return
			default:
			}

			// GetInput may return k == 0 right after initialization and after teardown
			// as means to unblock.
			if k == 0 {
				continue
			}

			input := Input{Key: k, Flag: f}
			im.dispatch(input)
		}
	}()
}

func (im *inputMonitor) dispatch(input Input) {
	im.notifyMtx.RLock()
	defer im.notifyMtx.RUnlock()

	for _, ch := range im.notify {
		select {
		case ch <- input:
		default:
		}
	}
	for _, ch := range im.notifyOn[input.asMapKey()] {
		select {
		case ch <- input:
		default:
		}
	}
}

func (im *inputMonitor) Teardown() {
	sys.Unhook()

	im.notifyOn = make(map[uint64][]chan<- Input)
	im.notify = nil
}

func (im *inputMonitor) NotifyOn(ch chan<- Input, inputs []Input) {
	im.notifyMtx.Lock()
	defer im.notifyMtx.Unlock()
	for _, v := range inputs {
		k := v.asMapKey()
		im.notifyOn[k] = append(im.notifyOn[k], ch)
	}
}

func (im *inputMonitor) Notify(ch chan<- Input) {
	im.notifyMtx.Lock()
	defer im.notifyMtx.Unlock()
	im.notify = append(im.notify, ch)
}

// Init must be called prior to Notify and NotifyOn.
func Init() {
	im.Init()
}

// Teardown must be called after a call to Init.
func Teardown() {
	im.Teardown()
}

// Notify sends the input on ch whenever an input is detected.
func Notify(ch chan<- Input) {
	im.Notify(ch)
}

// NotifyOn sends the input on ch whenever any of the inputs is dectected.
func NotifyOn(ch chan<- Input, inputs ...Input) {
	im.NotifyOn(ch, inputs)
}

func Send(input Input) error {
	return sys.Send(input.Key, input.Flag)
}
