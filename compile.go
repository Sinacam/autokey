package autokey

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

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

type Expr interface {
	Eval() interface{} // Evaluate the expression value and its side-effect
	Static() bool      // Returns true if the expression value is known statically and has no side-effect
}

// Compiles yml as a Expr recursively.
// Errors in the structure of yml is reported as an error.
// Errors in values causes a panic during execution of Expr instead.
func Compile(yml interface{}) (Expr, error) {
	fn, err := compile(yml)
	if err != "" {
		return nil, errors.New(err)
	}
	return fn, nil
}

type boolExpr bool

func (be boolExpr) Eval() interface{} {
	return bool(be)
}

func (be boolExpr) Static() bool {
	return true
}

type intExpr int

func (ie intExpr) Eval() interface{} {
	return int(ie)
}

func (ie intExpr) Static() bool {
	return true
}

type floatExpr float64

func (fe floatExpr) Eval() interface{} {
	return float64(fe)
}

func (fe floatExpr) Static() bool {
	return true
}

type stringExpr string

func (se stringExpr) Eval() interface{} {
	return string(se)
}

func (se stringExpr) Static() bool {
	return true
}

// compile uses an error string because the error trace is built up
// during recursion.
// TODO: migrate to recursive errors
func compile(yml interface{}) (Expr, string) {
	switch yml := yml.(type) {
	case bool:
		return boolExpr(yml), ""
	case int:
		return intExpr(yml), ""
	case float64:
		return floatExpr(yml), ""
	case string:
		return stringExpr(yml), ""
	case []interface{}:
		return compileSlice(yml)
	case map[interface{}]interface{}:
		return compileMap(yml)
	}
	return nil, ymlErrorString(yml)
}

type sliceExpr struct {
	subs       []Expr
	staticSubs []interface{}
	static     bool
}

func newSliceExpr(subs []Expr) *sliceExpr {
	static := true
	for _, v := range subs {
		static = static && v.Static()
	}

	se := sliceExpr{static: static}
	if static {
		for _, v := range subs {
			se.staticSubs = append(se.staticSubs, v.Eval())
		}
	} else {
		se.subs = subs
	}
	return &se
}

func (se *sliceExpr) Eval() interface{} {
	if se.Static() {
		return se.staticSubs
	}

	var ret []interface{}
	for _, v := range se.subs {
		ret = append(ret, v.Eval())
	}
	return ret
}

func (se *sliceExpr) Static() bool {
	return se.static
}

func compileSlice(yml []interface{}) (Expr, string) {
	var subs []Expr
	for i, v := range yml {
		sub, err := compile(v)
		if err != "" {
			return nil, addErrorTrace(err, i)
		}
		subs = append(subs, sub)
	}

	return newSliceExpr(subs), ""
}

func compileMap(yml map[interface{}]interface{}) (Expr, string) {
	var subs []Expr
	for k, v := range yml {
		kstr, ok := k.(string)
		if !ok {
			return nil, "key must be a string"
		}

		var sub Expr
		var err string
		switch kstr {
		case "do":
			sub, err = compileDo(v)
		case "repeat":
			sub, err = compileRepeat(v)
		case "press":
			sub, err = compilePress(v)
		case "hold":
			sub, err = compileHold(v)
		case "release":
			sub, err = compileRelease(v)
		case "file":
			sub, err = compileFile(v)
		default:
			return nil, "invalid key " + kstr
		}
		if err != "" {
			return nil, addErrorTrace(err, kstr)
		}
		subs = append(subs, sub)
	}

	// maps are treated identical to slices after compilation
	return newSliceExpr(subs), ""
}

// parseInput parses val as a slice of Inputs.
// Accepts int, string or a slice of them.
func parseInput(val interface{}, defaultFlag uint64) ([]Input, error) {
	switch val := val.(type) {
	case int:
		return []Input{
			{Key: val + '0', Flag: defaultFlag},
		}, nil
	case string:
		val = strings.ToLower(val)
		input, ok := inputMap[val]
		if !ok {
			break
		}

		if input.Flag == 0 {
			input.Flag = defaultFlag
		}
		return []Input{input}, nil
	case []interface{}:
		var inputs []Input
		for _, v := range val {
			input, err := parseInput(v, defaultFlag)
			if err != nil {
				return nil, err
			}
			inputs = append(inputs, input...)
		}
		return inputs, nil
	}
	return nil, errors.New("cannot parse as Input")
}

type doExpr struct {
	onExpr     Expr
	actionExpr Expr
	staticOn   []Input
}

func newDoExpr(onExpr, actionExpr Expr) (*doExpr, string) {
	de := &doExpr{actionExpr: actionExpr}
	if onExpr != nil && onExpr.Static() {
		val := onExpr.Eval()
		inputs, err := parseInput(val, KeyDown)
		if err != nil {
			return nil, err.Error()
		}
		de.staticOn = inputs
	} else {
		de.onExpr = onExpr
	}
	return de, ""
}

func (de *doExpr) Eval() interface{} {
	if de.Static() {
		return nil
	}

	// If there is no trigger, do is identity.
	if de.onExpr == nil && de.staticOn == nil {
		de.actionExpr.Eval()
		return nil
	}

	var err error
	inputs := de.staticOn
	if inputs == nil {
		val := de.onExpr.Eval()
		inputs, err = parseInput(val, KeyDown)
		if err != nil {
			panic(fmt.Sprintf("bad value for on: %v", val))
		}
	}

	ch := make(chan Input)
	NotifyOn(ch, inputs...)
	go func() {
		for range ch {
			de.actionExpr.Eval()
		}
	}()

	return nil
}

func (de *doExpr) Static() bool {
	return de.actionExpr == nil && de.actionExpr.Static()
}

// compileDo compiles the map value with key "do".
// Compiles by special-casing the "on" key as the trigger
// and delegating to compileMap for the remaining.
func compileDo(yml interface{}) (Expr, string) {
	var m map[interface{}]interface{}
	switch yml := yml.(type) {
	case map[interface{}]interface{}:
		m = yml
	case []interface{}:
		return compileSlice(yml)
	default:
		return nil, "value must be a mapping or sequence"
	}

	var onExpr Expr
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
			expr, err := compile(v)
			if err != "" {
				return nil, addErrorTrace(err, kstr)
			}
			onExpr = expr
		default:
			remaining[k] = v
		}
	}

	actionExpr, err := compileMap(remaining)
	if err != "" {
		return nil, err
	}

	de, err := newDoExpr(onExpr, actionExpr)
	if err != "" {
		return nil, addErrorTrace(err, "on")
	}
	return de, ""
}

type hertz float64

// parseFreq parses val as a frequency.
// Accepts strings.
func parseFreq(val interface{}) (hertz, error) {
	s, ok := val.(string)
	if !ok {
		return 0, errors.New("cannot parse as frequency")
	}

	s = strings.ToLower(s)
	if !strings.HasSuffix(s, "hz") {
		return 0, errors.New("cannot parse as frequency")
	}

	freq, err := strconv.ParseFloat(s[:len(s)-2], 64)
	if err != nil {
		return 0, err
	}
	return hertz(freq), nil
}

// parseDuration parses val as a duration.
// Accepts strings.
func parseDuration(val interface{}) (time.Duration, error) {
	s, ok := val.(string)
	if !ok {
		return 0, errors.New("cannot parse as duration")
	}

	return time.ParseDuration(s)
}

// parseTrigger parses val as a trigger which sends to ch.
// Returns an installation function.
// Accepts strings or slices.
func parseTrigger(val interface{}, ch chan<- struct{}) (func(), error) {
	// TODO: uninstall function?
	switch yml := val.(type) {
	case string:
		inputs, err := parseInput(yml, KeyDown)
		if err != nil {
			return nil, err
		}

		return func() {
			in := make(chan Input)
			NotifyOn(in, inputs...)
			go func() {
				for range in {
					select {
					case ch <- struct{}{}:
					default:
					}
				}
			}()
		}, nil
	case []interface{}:
		var installs []func()
		for _, v := range yml {
			install, err := parseTrigger(v, ch)
			if err != nil {
				return nil, err
			}
			installs = append(installs, install)
		}
		return func() {
			for _, v := range installs {
				v()
			}
		}, nil
	}

	return nil, errors.New("cannot parse as trigger")
}

type repeatExpr struct {
	atExpr      Expr
	forExpr     Expr
	untilExpr   Expr
	actionExpr  Expr
	staticAt    hertz
	staticFor   time.Duration
	staticUntil func()
	untilCh     chan struct{}
}

func newRepeatExpr(atExpr, forExpr, untilExpr, actionExpr Expr) (*repeatExpr, string) {
	re := &repeatExpr{
		actionExpr: actionExpr,
		untilCh:    make(chan struct{}),
	}

	if atExpr.Static() {
		val := atExpr.Eval()
		freq, err := parseFreq(val)
		if err != nil {
			return nil, err.Error()
		}
		re.staticAt = freq
	} else {
		re.atExpr = atExpr
	}

	if untilExpr == nil && forExpr == nil {
		return nil, "must contain either until or for"
	}

	if untilExpr != nil && untilExpr.Static() {
		val := untilExpr.Eval()
		install, err := parseTrigger(val, re.untilCh)
		if err != nil {
			return nil, err.Error()
		}
		re.staticUntil = install
	} else {
		re.untilExpr = untilExpr
	}

	if forExpr != nil && forExpr.Static() {
		val := forExpr.Eval()
		dur, err := parseDuration(val)
		if err != nil {
			return nil, err.Error()
		}
		if dur <= 0 {
			return nil, "duration has to be positive"
		}
		re.staticFor = dur
	} else {
		re.forExpr = forExpr
	}

	return re, ""
}

func (re *repeatExpr) Eval() interface{} {
	if re.Static() {
		return nil
	}

	var freq hertz
	var err error
	if re.atExpr == nil {
		freq = re.staticAt
	} else {
		val := re.atExpr.Eval()
		freq, err = parseFreq(val)
		if err != nil {
			panic(fmt.Sprintf("bad value for at: %v", val))
		}
	}

	var install func()
	if re.untilExpr == nil {
		install = re.staticUntil
	} else {
		val := re.untilExpr.Eval()
		install, err = parseTrigger(val, re.untilCh)
		if err != nil {
			panic(fmt.Sprintf("bad value for until: %v", val))
		}
	}

	var dur time.Duration
	if re.forExpr == nil {
		dur = re.staticFor
	} else {
		val := re.forExpr.Eval()
		dur, err = parseDuration(val)
		if err != nil || dur <= 0 {
			panic(fmt.Sprintf("bad value for for: %v", val))
		}
	}

	if install != nil {
		install()
	}
	ticker := time.NewTicker(time.Duration(float64(time.Second) / float64(freq)))
	defer ticker.Stop()

	if dur > 0 {
		timer := time.NewTimer(dur)
		for {
			select {
			case <-re.untilCh:
				return nil
			case <-timer.C:
				return nil
			case <-ticker.C:
				re.actionExpr.Eval()
			}
		}
	} else {
		for {
			select {
			case <-re.untilCh:
				return nil
			case <-ticker.C:
				re.actionExpr.Eval()
			}
		}
	}
}

func (re *repeatExpr) Static() bool {
	return re.actionExpr == nil || re.actionExpr.Static()
}

func compileRepeat(yml interface{}) (Expr, string) {
	m, ok := yml.(map[interface{}]interface{})
	if !ok {
		return nil, "value must be a mapping"
	}

	var (
		atExpr    Expr
		forExpr   Expr
		untilExpr Expr
	)
	remaining := make(map[interface{}]interface{})
	for k, v := range m {
		kstr, ok := k.(string)
		if !ok {
			return nil, "key must be a string"
		}

		switch kstr {
		case "at":
			Expr, err := compile(v)
			if err != "" {
				return nil, addErrorTrace(err, kstr)
			}
			atExpr = Expr
		case "for":
			Expr, err := compile(v)
			if err != "" {
				return nil, addErrorTrace(err, kstr)
			}
			forExpr = Expr
		case "until":
			Expr, err := compile(v)
			if err != "" {
				return nil, addErrorTrace(err, kstr)
			}
			untilExpr = Expr
		default:
			remaining[k] = v
		}
	}

	if atExpr == nil {
		return nil, "missing at"
	}

	actionExpr, err := compileMap(remaining)
	if err != "" {
		return nil, err
	}

	re, err := newRepeatExpr(atExpr, forExpr, untilExpr, actionExpr)
	if err != "" {
		return nil, err
	}

	return re, ""
}

type pressExpr struct {
	expr   Expr
	static []Input
}

func newPressExpr(expr Expr) (*pressExpr, string) {
	pe := &pressExpr{}
	if expr.Static() {
		val := expr.Eval()
		inputs, err := parseInput(val, 0)
		if err != nil {
			return nil, err.Error()
		}
		pe.static = inputs
	} else {
		pe.expr = expr
	}

	return pe, ""
}

func (pe *pressExpr) Eval() interface{} {
	inputs := pe.static
	var err error
	if inputs == nil {
		val := pe.expr.Eval()
		inputs, err = parseInput(val, 0)
		if err != nil {
			panic(fmt.Sprintf("bad value for press: %v", val))
		}
	}

	// Specifying nothing means keydown and keyup for press.
	// Order is keydown over all inputs before keyup to allow
	// key combinations such as ctrl + c.
	for _, input := range inputs {
		if input.Flag == 0 {
			input.Flag = KeyDown
		}
		Send(input)
	}

	for _, input := range inputs {
		if input.Flag == 0 {
			input.Flag = KeyUp
			Send(input)
		}
	}
	return nil
}

func (pe *pressExpr) Static() bool {
	return false
}

func compilePress(yml interface{}) (Expr, string) {
	expr, err := compile(yml)
	if err != "" {
		return nil, err
	}

	pe, err := newPressExpr(expr)
	if err != "" {
		return nil, err
	}

	return pe, ""
}

type holdExpr struct {
	expr   Expr
	static []Input
}

func newHoldExpr(expr Expr) (*holdExpr, string) {
	he := &holdExpr{}
	if expr.Static() {
		val := expr.Eval()
		inputs, err := parseInput(val, KeyDown)
		if err != nil {
			return nil, err.Error()
		}
		he.static = inputs
	} else {
		he.expr = expr
	}

	return he, ""
}

func (he *holdExpr) Eval() interface{} {
	inputs := he.static
	var err error
	if inputs == nil {
		val := he.expr.Eval()
		inputs, err = parseInput(val, KeyDown)
		if err != nil {
			panic(fmt.Sprintf("bad value for hold: %v", val))
		}
	}

	for _, input := range inputs {
		Send(input)
	}
	return nil
}

func (he *holdExpr) Static() bool {
	return false
}

func compileHold(yml interface{}) (Expr, string) {
	expr, err := compile(yml)
	if err != "" {
		return nil, err
	}

	he, err := newHoldExpr(expr)
	if err != "" {
		return nil, err
	}

	return he, ""
}

type releaseExpr struct {
	expr   Expr
	static []Input
}

func newReleaseExpr(expr Expr) (*releaseExpr, string) {
	re := &releaseExpr{}
	if expr.Static() {
		val := expr.Eval()
		inputs, err := parseInput(val, KeyUp)
		if err != nil {
			return nil, err.Error()
		}
		re.static = inputs
	} else {
		re.expr = expr
	}

	return re, ""
}

func (re *releaseExpr) Eval() interface{} {
	inputs := re.static
	var err error
	if inputs == nil {
		val := re.expr.Eval()
		inputs, err = parseInput(val, KeyDown)
		if err != nil {
			panic(fmt.Sprintf("bad value for hold: %v", val))
		}
	}

	for _, input := range inputs {
		Send(input)
	}
	return nil
}

func (re *releaseExpr) Static() bool {
	return false
}

func compileRelease(yml interface{}) (Expr, string) {
	expr, err := compile(yml)
	if err != "" {
		return nil, err
	}

	re, err := newReleaseExpr(expr)
	if err != "" {
		return nil, err
	}

	return re, ""
}

type fileExpr struct {
	expr       Expr        // expr is not static
	staticExpr Expr        // expr is static, but file content is not
	staticVal  interface{} // both expr and file content is static
}

func newFileExpr(expr Expr) (*fileExpr, string) {
	fe := &fileExpr{}
	if expr.Static() {
		val := expr.Eval()
		path, ok := val.(string)
		if !ok {
			return nil, "file path must be a string"
		}
		f, err := os.Open(path)
		if err != nil {
			return nil, err.Error()
		}
		defer f.Close()

		var m interface{}
		err = yaml.NewDecoder(f).Decode(&m)
		if err != nil {
			return nil, err.Error()
		}

		mexpr, err := Compile(m)
		if err != nil {
			return nil, err.Error()
		}

		if mexpr.Static() {
			fe.staticVal = mexpr.Eval()
		} else {
			fe.staticExpr = mexpr
		}
	} else {
		fe.expr = expr
	}
	return fe, ""
}

func (fe *fileExpr) Eval() interface{} {
	if fe.staticVal != nil {
		return fe.staticVal
	}

	if fe.staticExpr != nil {
		return fe.staticExpr.Eval()
	}

	val := fe.expr.Eval()
	path, ok := val.(string)
	if !ok {
		panic(fmt.Sprintf("bad value for file: %v", val))
	}
	f, err := os.Open(path)
	if err != nil {
		panic(err)
	}
	defer f.Close()

	var m map[interface{}]interface{}
	err = yaml.NewDecoder(f).Decode(&m)
	if err != nil {
		panic(err)
	}

	mexpr, err := Compile(m)
	if err != nil {
		panic(err)
	}

	return mexpr.Eval()
}

func (fe *fileExpr) Static() bool {
	return fe.staticVal != nil
}

func compileFile(yml interface{}) (Expr, string) {
	expr, err := compile(yml)
	if err != "" {
		return nil, err
	}

	fe, err := newFileExpr(expr)
	if err != "" {
		return nil, err
	}

	return fe, ""
}
