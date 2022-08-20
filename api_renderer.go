package main

import (
	rt "github.com/arnodel/golua/runtime"
	"github.com/arnodel/golua/lib/packagelib"
)

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
