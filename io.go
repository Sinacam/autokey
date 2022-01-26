package autokey

import (
	"strings"
	"sync"

	"github.com/Sinacam/autokey/sys"
)

const (
	KeyDown        = sys.KeyDown
	KeyUp          = sys.KeyUp
	LeftMouseDown  = sys.LeftMouseDown
	LeftMouseUp    = sys.LeftMouseUp
	RightMouseDown = sys.RightMouseDown
	RightMouseUp   = sys.RightMouseUp
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
			input := Input{Key: k, Flag: f}

			select {
			case <-im.done:
				return
			default:
			}

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
	switch input.Flag {
	case KeyDown:
		fallthrough
	case KeyUp:
		sys.KeybdEvent(input.Key, input.Flag)
	case LeftMouseDown:
		fallthrough
	case LeftMouseUp:
		fallthrough
	case RightMouseDown:
		fallthrough
	case RightMouseUp:
		sys.MouseEvent(input.Flag)
	default:
		return InvalidFlag
	}
	return nil
}
