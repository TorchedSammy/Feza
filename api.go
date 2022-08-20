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
var regexLoader = packagelib.Loader{
	Name: "regex",
	Load: regexLoad,
}

func regexLoad(rtm *rt.Runtime) (rt.Value, func()) {
	exports := map[string]luaExport{}
	mod := rt.NewTable()
	setExports(rtm, mod, exports)

	return rt.TableValue(mod), nil
}
