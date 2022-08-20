package main

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/veandco/go-sdl2/sdl"
	rt "github.com/arnodel/golua/runtime"
	"github.com/arnodel/golua/lib/packagelib"
)

type luaExport struct {
	Function rt.GoFunctionFunc
	ArgNum int
	Variadic bool
}

func setExports(rtm *rt.Runtime, tbl *rt.Table, exports map[string]luaExport) {
	for name, export := range exports {
		rtm.SetEnvGoFunc(tbl, name, export.Function, export.ArgNum, export.Variadic)
	}
}

// %----- renderer api
var rendererLoader = packagelib.Loader{
	Name: "renderer",
	Load: rendererLoad,
}

func rendererLoad(rtm *rt.Runtime) (rt.Value, func()) {
	exports := map[string]luaExport{
		"begin_frame": {rendererBeginFrame, 0, false},
		"end_frame": {rendererEndFrame, 0, false},
		"draw_rect": {rendererDrawRect, 5, false},
		"draw_text": {rendererDrawText, 5, false},
	}
	mod := rt.NewTable()
	setExports(rtm, mod, exports)

	return rt.TableValue(mod), nil
}

func rendererBeginFrame(t *rt.Thread, c *rt.GoCont) (rt.Cont, error) {
	wnd.StartFrame()
	return c.Next(), nil
}

func rendererEndFrame(t *rt.Thread, c *rt.GoCont) (rt.Cont, error) {
	wnd.FinishFrame()
	return c.Next(), nil
}

func rendererDrawRect(t *rt.Thread, c *rt.GoCont) (rt.Cont, error) {
	if err := c.CheckNArgs(5); err != nil {
		return nil, err
	}

	x, err := c.FloatArg(0)
	if err != nil {
		return nil, err
	}
	y, err := c.FloatArg(1)
	if err != nil {
		return nil, err
	}
	w, err := c.FloatArg(2)
	if err != nil {
		return nil, err
	}
	h, err := c.FloatArg(3)
	if err != nil {
		return nil, err
	}
	color, err := c.TableArg(4)
	if err != nil {
		return nil, err
	}

	r := int(color.Get(rt.IntValue(1)).AsInt())
	g := int(color.Get(rt.IntValue(2)).AsInt())
	b := int(color.Get(rt.IntValue(3)).AsInt())
	a := int(color.Get(rt.IntValue(4)).AsInt())

	cv.SetFillStyle(r, g, b, a)
	cv.FillRect(x, y, w, h)

	return c.Next(), nil
}

func rendererDrawText(t *rt.Thread, c *rt.GoCont) (rt.Cont, error) {
	if err := c.CheckNArgs(5); err != nil {
		return nil, err
	}

	_, err := c.StringArg(0)
	if err != nil {
		return nil, err
	}
	text, err := c.StringArg(1)
	if err != nil {
		return nil, err
	}
	x, err := c.FloatArg(2)
	if err != nil {
		return nil, err
	}
	y, err := c.FloatArg(3)
	if err != nil {
		return nil, err
	}
	color, err := c.TableArg(4)
	if err != nil {
		return nil, err
	}

	r := int(color.Get(rt.IntValue(1)).AsInt())
	g := int(color.Get(rt.IntValue(2)).AsInt())
	b := int(color.Get(rt.IntValue(3)).AsInt())
	a := int(color.Get(rt.IntValue(4)).AsInt())

	f, _ := cv.LoadFont("/usr/share/fonts/dejavu-sans-fonts/DejaVuSans.ttf")
	cv.SetFont(f, 36)
	cv.SetFillStyle(r, g, b, a)
	cv.FillText(text, x, y)

	return c.Next(), nil
}

// %----- system api
var systemLoader = packagelib.Loader{
	Name: "system",
	Load: systemLoad,
}

func systemLoad(rtm *rt.Runtime) (rt.Value, func()) {
	exports := map[string]luaExport{
		"poll_event": {systemPollEvent, 0, false},
	}
	mod := rt.NewTable()
	setExports(rtm, mod, exports)

	return rt.TableValue(mod), nil
}

func systemPollEvent(t *rt.Thread, c *rt.GoCont) (rt.Cont, error) {
	event := sdl.PollEvent()

	switch event.(type) {
		//case 
	}
	return c.Next(), nil
}

// %----- process api
var processLoader = packagelib.Loader{
	Name: "process",
	Load: processLoad,
}

func processLoad(rtm *rt.Runtime) (rt.Value, func()) {
	exports := map[string]luaExport{}
	mod := rt.NewTable()
	setExports(rtm, mod, exports)

	return rt.TableValue(mod), nil
}

// %----- regex api
const (
	REGEX_ANCHORED = iota + 1 << 2
	REGEX_ENDANCHORED
)

var regexMetaKey = rt.StringValue("_lavateraRegexp")
var regexLoader = packagelib.Loader{
	Name: "regex",
	Load: regexLoad,
}

func regexLoad(rtm *rt.Runtime) (rt.Value, func()) {
	regexMethods := rt.NewTable()
	r.SetEnvGoFunc(regexMethods, "cmatch", regexMatch, 2, false)

	regexMeta := rt.NewTable()
	r.SetEnv(regexMeta, "__index", rt.TableValue(regexMethods))
	r.SetRegistry(regexMetaKey, rt.TableValue(regexMeta))

	exports := map[string]luaExport{
		"compile": {regexCompile, 1, true},
	}
	mod := rt.NewTable()
	setExports(rtm, mod, exports)

	r.SetEnv(mod, "ANCHORED", rt.IntValue(REGEX_ANCHORED))
	r.SetEnv(mod, "ENDANCHORED", rt.IntValue(REGEX_ENDANCHORED))

	return rt.TableValue(mod), nil
}

func regexArg(c *rt.GoCont, n int) (*regexp.Regexp, error) {
	re, ok := valueToRegex(c.Arg(n))
	if ok {
		return re, nil
	}
	return nil, fmt.Errorf("#%d must be a regex", n+1)
}

func valueToRegex(v rt.Value) (re *regexp.Regexp, ok bool) {
	var u *rt.UserData
	u, ok = v.TryUserData()
	if ok {
		re, ok = u.Value().(*regexp.Regexp)
	}
	return
}

func regexCompile(t *rt.Thread, c *rt.GoCont) (rt.Cont, error) {
	if err := c.Check1Arg(); err != nil {
		return nil, err
	}

	pat, err := c.StringArg(0)
	if err != nil {
		return nil, err
	}

	var flags string
	if len(c.Etc()) != 0 {
		options, ok := c.Etc()[0].TryString()
		if !ok {
			return nil, fmt.Errorf("??")
		}

		if strings.Contains(options, "i") {
			flags += "i"
		}
		if strings.Contains(options, "s") {
			flags += "s"
		}
		if strings.Contains(options, "m") {
			flags += "s"
		}
	}

	if flags != "" {
		pat = "(?" + flags + ")" + pat
	}

	re, err := regexp.Compile(pat)
	if err != nil {
		return nil, err
	}

	return c.PushingNext1(t.Runtime, t.NewUserDataValue(re, regexMetaKey.AsTable())), nil
}

func regexMatch(t *rt.Thread, c *rt.GoCont) (rt.Cont, error) {
	re, err := regexArg(c, 0)
	if err != nil {
		return nil, err
	}

	subject, err := c.StringArg(0)
	if err != nil {
		return nil, err
	}

	var indexes [][]int
	if len(c.Etc()) != 0 {
		offset, ok := c.Etc()[0].TryInt()
		if !ok {
			return nil, fmt.Errorf("??")
		}
		var opts int64
		if len(c.Etc()) > 1 {
			var ok bool
			opts, ok = c.Etc()[0].TryInt()
			if !ok {
				return nil, fmt.Errorf("??")
			}
		}

		match := true
		anchored := opts & REGEX_ANCHORED == REGEX_ANCHORED
		endAnchored := opts & REGEX_ENDANCHORED == REGEX_ENDANCHORED;

		ln := re.FindStringIndex(subject)
		if anchored {
			start := ln[0]
			if int64(start) + offset != offset {
				match = false
			}
		}

		if endAnchored {
			end := ln[1]
			if end != len(subject) {
				match = false
			}
		}

		if match {
			indexes = append(indexes, ln)
		}
	}

	luaIndexes := rt.NewTable()
	if len(indexes) == 0 {
		indexes = re.FindAllStringIndex(subject, -1)
	}

	for v, idxSlice := range indexes {
		luaIdxSlice := rt.NewTable()
		for vv, i := range idxSlice {
			luaIdxSlice.Set(rt.IntValue(int64(vv)), rt.IntValue(int64(i)))
		}

		luaIndexes.Set(rt.IntValue(int64(v)), rt.TableValue(luaIdxSlice))
	}

	return c.PushingNext1(t.Runtime, rt.TableValue(luaIndexes)), nil
}
